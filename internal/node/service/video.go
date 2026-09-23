package service

import (
	"context"

	v1 "videoCluster/api/node/v1"
	"videoCluster/internal/node/biz"
)

// VideoService exposes Node video metadata APIs.
type VideoService struct {
	v1.UnimplementedNodeServiceServer

	vu *biz.VideoUsecase
}

// NewVideoService creates a Node video service.
func NewVideoService(vu *biz.VideoUsecase) *VideoService {
	return &VideoService{vu: vu}
}

func (ns VideoService) ListVideos(ctx context.Context, req *v1.ListVideosRequest) (*v1.ListVideosReply, error) {
	videoList, err := ns.vu.ListVideos(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.ListVideosReply{
		Videos: videoList,
	}, nil
}

func (ns VideoService) GetVideoInfo(ctx context.Context, req *v1.GetVideoInfoRequest) (*v1.GetVideoInfoReply, error) {
	video, err := ns.FindVideo(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.GetVideoInfoReply{
		Video: video,
	}, nil
}

// FindVideo returns metadata used by the native HTTP streaming handler.
func (ns *VideoService) FindVideo(ctx context.Context, id string) (*v1.VideoMeta, error) {
	return ns.vu.GetVideoInfo(ctx, id)
}
