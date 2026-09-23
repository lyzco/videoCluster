package biz

import (
	"context"
	"errors"

	v1 "videoCluster/api/node/v1"
)

var ErrVideoNotFound = errors.New("video not found")

type VideoRepository interface {
	Save(ctx context.Context, video *v1.VideoMeta) error
	SaveScanned(ctx context.Context, video *v1.VideoMeta, scanRoot, scanVersion string) error
	DeleteUnseen(ctx context.Context, scanRoot, scanVersion string) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*v1.VideoMeta, error)
	ListAll(ctx context.Context) ([]*v1.VideoMeta, error)
}

type VideoUsecase struct {
	repo VideoRepository
}

func NewVideoUsecase(repo VideoRepository) *VideoUsecase {
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
