package biz_test

import (
	"context"
	"testing"

	v1 "videoCluster/api/node/v1"
	"videoCluster/internal/node/biz"
	"videoCluster/internal/node/conf"
)

func TestScanRemovesUnseenVideosOnlyAfterSuccessfulWalk(t *testing.T) {
	t.Run("successful empty scan cleans stale videos", func(t *testing.T) {
		repo := &scannerRepository{}
		scanner := biz.NewVideoScanner(&conf.Data{Paths: []string{t.TempDir()}}, repo)

		if err := scanner.ScanAndCache(context.Background()); err != nil {
			t.Fatalf("scan: %v", err)
		}
		if repo.deleteCalls != 1 {
			t.Fatalf("delete calls = %d, want 1", repo.deleteCalls)
		}
	})

	t.Run("failed walk preserves stale videos", func(t *testing.T) {
		repo := &scannerRepository{}
		scanner := biz.NewVideoScanner(&conf.Data{Paths: []string{t.TempDir() + "/missing"}}, repo)

		if err := scanner.ScanAndCache(context.Background()); err == nil {
			t.Fatal("scan succeeded for missing root")
		}
		if repo.deleteCalls != 0 {
			t.Fatalf("delete calls = %d, want 0", repo.deleteCalls)
		}
	})
}

type scannerRepository struct {
	deleteCalls int
}

func (r *scannerRepository) Save(context.Context, *v1.VideoMeta) error { return nil }
func (r *scannerRepository) SaveScanned(context.Context, *v1.VideoMeta, string, string) error {
	return nil
}
func (r *scannerRepository) DeleteUnseen(context.Context, string, string) error {
	r.deleteCalls++
	return nil
}
func (r *scannerRepository) Delete(context.Context, string) error { return nil }
func (r *scannerRepository) Get(context.Context, string) (*v1.VideoMeta, error) {
	return nil, biz.ErrVideoNotFound
}
func (r *scannerRepository) ListAll(context.Context) ([]*v1.VideoMeta, error) { return nil, nil }
