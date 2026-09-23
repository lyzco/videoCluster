package service

import (
	"context"

	v1 "videoCluster/api/central/v1"
	"videoCluster/internal/central/biz"
)

type VideoService struct {
	v1.UnimplementedCentralServiceServer
	aggregator *biz.VideoAggregator
}

func NewVideoService(aggregator *biz.VideoAggregator) *VideoService {
	return &VideoService{aggregator: aggregator}
}

func (s *VideoService) ListVideos(ctx context.Context, _ *v1.ListVideosRequest) (*v1.ListVideosReply, error) {
	videos, err := s.aggregator.ListVideos(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.ListVideosReply{Videos: videos}, nil
}

func (s *VideoService) GetVideoInfo(ctx context.Context, request *v1.GetVideoInfoRequest) (*v1.GetVideoInfoReply, error) {
	video, err := s.aggregator.GetVideoInfo(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	return &v1.GetVideoInfoReply{Video: video}, nil
}
