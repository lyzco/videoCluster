package biz

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	v1 "videoCluster/api/node/v1"
	"videoCluster/internal/node/conf"
	"videoCluster/internal/node/media"
)

type VideoScanner struct {
	repo       VideoRepository
	basePaths  []string
	generation atomic.Uint64
}

var videoExtensions = map[string]struct{}{
	".mp4": {}, ".mkv": {}, ".avi": {}, ".mov": {}, ".flv": {},
}

func NewVideoScanner(c *conf.Data, repo VideoRepository) *VideoScanner {
	return &VideoScanner{repo: repo, basePaths: normalizeBasePaths(c.Paths)}
}

func (v *VideoScanner) ScanAndCache(ctx context.Context) error {
	var errs []error
	for _, root := range v.basePaths {
		if err := v.scanRoot(ctx, root); err != nil {
			errs = append(errs, fmt.Errorf("scan %s: %w", root, err))
		}
	}
	return errors.Join(errs...)
}

func (v *VideoScanner) scanRoot(ctx context.Context, root string) error {
	version := fmt.Sprintf("%d-%d", time.Now().UnixNano(), v.generation.Add(1))
	files := make(chan string, 100)
	var workers sync.WaitGroup
	var processErr error
	var processErrMu sync.Mutex

	for range 8 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for path := range files {
				if err := v.processFile(ctx, path, root, version); err != nil {
					log.Context(ctx).Warnf("process file failed: %v", err)
					processErrMu.Lock()
					processErr = errors.Join(processErr, err)
					processErrMu.Unlock()
				}
			}
		}()
	}

	walkErr := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !isVideo(path) || v.ownerRoot(path) != root {
			return nil
		}
		select {
		case files <- path:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
	close(files)
	workers.Wait()

	if err := errors.Join(walkErr, processErr); err != nil {
		return err
	}
	return v.repo.DeleteUnseen(ctx, root, version)
}

func (v *VideoScanner) processFile(ctx context.Context, path, root, version string) error {
	ext := strings.ToLower(filepath.Ext(path))
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	fingerprintID, err := media.FileFingerprint(path)
	if err != nil {
		return err
	}

	thumbnail := ""
	existing, err := v.repo.Get(ctx, fingerprintID)
	if err == nil && existing != nil {
		thumbnail = existing.Thumbnail
	} else if !errors.Is(err, ErrVideoNotFound) {
		return err
	}
	if thumbnail == "" {
		thumbnail, err = media.GenerateThumbnail(ctx, path)
		if err != nil {
			log.Context(ctx).Warnf("thumbnail failed: %v", err)
		}
	}

	video := &v1.VideoMeta{
		Id:        fingerprintID,
		Title:     strings.TrimSuffix(info.Name(), ext),
		Path:      path,
		Size:      info.Size(),
		Thumbnail: thumbnail,
	}

	return v.repo.SaveScanned(ctx, video, root, version)
}

func normalizeBasePaths(paths []string) []string {
	roots := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		absolute, err := filepath.Abs(path)
		if err != nil {
			absolute = path
		}
		roots[filepath.Clean(absolute)] = struct{}{}
	}
	result := make([]string, 0, len(roots))
	for root := range roots {
		result = append(result, root)
	}
	sort.Slice(result, func(i, j int) bool {
		return len(result[i]) > len(result[j])
	})
	return result
}

func (v *VideoScanner) ownerRoot(path string) string {
	for _, root := range v.basePaths {
		relative, err := filepath.Rel(root, path)
		if err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return root
		}
	}
	return ""
}

func isVideo(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	_, ok := videoExtensions[ext]
	return ok
}
