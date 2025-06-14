package biz

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	v1 "videoCluster/api/node/v1"
	"videoCluster/internal/common"
	"videoCluster/internal/conf"
	"videoCluster/internal/data"

	"github.com/go-kratos/kratos/v2/log"
)

type VideoScanner struct {
	repo     VideoRepository
	basePath []string
}

var videoExtensions = map[string]struct{}{
	".mp4": {}, ".mkv": {}, ".avi": {}, ".mov": {}, ".flv": {},
}

func NewVideoScanner(c *conf.Data, repo *data.VideoVideoRepository) *VideoScanner {
	return &VideoScanner{repo: repo, basePath: c.Paths}
}

func (v *VideoScanner) ScanAndCache(ctx context.Context) error {
	fileChan := make(chan string, 100)
	var wg sync.WaitGroup

	// 生产者扫描目录
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, path := range v.basePath {
			filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
				if err == nil && !info.IsDir() && isVideo(p) {
					fileChan <- p
				}
				return nil
			})
		}
		close(fileChan)
	}()

	// 多协程消费者处理
	numWorkers := 8
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range fileChan {
				if err := v.processFile(ctx, path); err != nil {
					log.Context(ctx).Warnf("process file failed: %v", err)
				}
			}
		}()
	}

	wg.Wait()
	return nil
}

func (v *VideoScanner) processFile(ctx context.Context, path string) error {
	ext := strings.ToLower(filepath.Ext(path))
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	hashId, err := common.ParallelFileHash(path)
	if err != nil {
		return err
	}

	existing, err := v.repo.Get(ctx, hashId)
	if err == nil && existing != nil {
		// 存在则更新部分字段
		existing.Path = path
		existing.Title = strings.TrimSuffix(info.Name(), ext)
		return v.repo.Save(ctx, existing)
	}

	// 不存在才生成缩略图
	thumbnail, err := common.GenerateThumbnail(ctx, path)
	if err != nil {
		log.Context(ctx).Warnf("thumbnail failed: %v", err)
	}

	video := &v1.VideoMeta{
		Id:        hashId,
		Title:     strings.TrimSuffix(info.Name(), ext),
		Path:      path,
		Size:      info.Size(),
		Thumbnail: thumbnail,
	}

	return v.repo.Save(ctx, video)
}

func isVideo(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	_, ok := videoExtensions[ext]
	return ok
}
