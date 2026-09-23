package data_test

import (
	"context"
	"testing"

	"errors"

	"github.com/dgraph-io/badger/v4"
	v1 "videoCluster/api/node/v1"
	"videoCluster/internal/node/biz"
	"videoCluster/internal/node/data"
)

func TestVideoRepositoryCRUD(t *testing.T) {
	db, err := badger.Open(badger.DefaultOptions(t.TempDir()).WithLogger(nil))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close BadgerDB: %v", err)
		}
	})

	repo := data.NewVideoVideoRepository(db)
	ctx := context.Background()
	want := &v1.VideoMeta{
		Id:        "video-id",
		Title:     "Example",
		Path:      "/videos/example.mp4",
		Size:      1234,
		Thumbnail: "thumbnail",
	}

	if err := repo.Save(ctx, want); err != nil {
		t.Fatalf("save video: %v", err)
	}
	got, err := repo.Get(ctx, want.Id)
	if err != nil {
		t.Fatalf("get video: %v", err)
	}
	if got.Id != want.Id || got.Title != want.Title || got.Path != want.Path || got.Size != want.Size || got.Thumbnail != want.Thumbnail {
		t.Fatalf("unexpected video: got %+v, want %+v", got, want)
	}

	videos, err := repo.ListAll(ctx)
	if err != nil {
		t.Fatalf("list videos: %v", err)
	}
	if len(videos) != 1 || videos[0].Id != want.Id {
		t.Fatalf("unexpected video list: %+v", videos)
	}

	if err := repo.Delete(ctx, want.Id); err != nil {
		t.Fatalf("delete video: %v", err)
	}
	if _, err := repo.Get(ctx, want.Id); !errors.Is(err, biz.ErrVideoNotFound) {
		t.Fatalf("get deleted video: got %v, want %v", err, biz.ErrVideoNotFound)
	}
}

func TestDeleteUnseenRemovesOnlyStaleVideosInRoot(t *testing.T) {
	db, err := badger.Open(badger.DefaultOptions(t.TempDir()).WithLogger(nil))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	repo := data.NewVideoVideoRepository(db)
	ctx := context.Background()
	const root = "/videos"
	const firstVersion = "first"
	const currentVersion = "current"
	current := &v1.VideoMeta{Id: "current", Path: "/videos/current.mp4"}
	stale := &v1.VideoMeta{Id: "stale", Path: "/videos/stale.mp4"}
	otherRoot := &v1.VideoMeta{Id: "other-root", Path: "/other/video.mp4"}

	for _, video := range []*v1.VideoMeta{current, stale} {
		if err := repo.SaveScanned(ctx, video, root, firstVersion); err != nil {
			t.Fatalf("save initial video: %v", err)
		}
	}
	if err := repo.SaveScanned(ctx, current, root, currentVersion); err != nil {
		t.Fatalf("update current video version: %v", err)
	}
	if err := repo.SaveScanned(ctx, otherRoot, "/other", firstVersion); err != nil {
		t.Fatalf("save other root video: %v", err)
	}

	if err := repo.DeleteUnseen(ctx, root, currentVersion); err != nil {
		t.Fatalf("delete unseen videos: %v", err)
	}
	if _, err := repo.Get(ctx, current.Id); err != nil {
		t.Fatalf("get current video: %v", err)
	}
	if _, err := repo.Get(ctx, stale.Id); !errors.Is(err, biz.ErrVideoNotFound) {
		t.Fatalf("get stale video: got %v, want %v", err, biz.ErrVideoNotFound)
	}
	if _, err := repo.Get(ctx, otherRoot.Id); err != nil {
		t.Fatalf("get other-root video: %v", err)
	}
}
