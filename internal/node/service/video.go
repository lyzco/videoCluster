package service

import (
	"context"
	v1 "videoCluster/api/node/v1"
	"videoCluster/internal/biz"
)

// VideoService is a greeter service.
type VideoService struct {
	v1.NodeServiceServer

	vu *biz.VideoUsecase
}

// NewVideoService new a greeter service.
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
	video, err := ns.vu.GetVideoInfo(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.GetVideoInfoReply{
		Video: video,
	}, nil
}
