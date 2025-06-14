package biz

import (
	"context"
	v1 "videoCluster/api/node/v1"
	"videoCluster/internal/data"
)

type VideoRepository interface {
	Save(ctx context.Context, video *v1.VideoMeta) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*v1.VideoMeta, error)
	ListAll(ctx context.Context) ([]*v1.VideoMeta, error)
}

type VideoUsecase struct {
	repo VideoRepository
}

func NewVideoUsecase(repo *data.VideoVideoRepository) *VideoUsecase {
	return &VideoUsecase{repo: repo}
}

func (uc *VideoUsecase) ListVideos(ctx context.Context) ([]*v1.VideoMeta, error) {
	return uc.repo.ListAll(ctx)
}

func (uc *VideoUsecase) DeleteVideo(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}

func (uc *VideoUsecase) GetVideoInfo(ctx context.Context, id string) (*v1.VideoMeta, error) {
	return uc.repo.Get(ctx, id)
}

func (uc *VideoUsecase) SaveVideo(ctx context.Context, video *v1.VideoMeta) error {
	return uc.repo.Save(ctx, video)
}
