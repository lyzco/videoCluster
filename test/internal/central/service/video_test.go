package service_test

import (
	"context"
	"testing"

	v1 "videoCluster/api/central/v1"
	"videoCluster/internal/central/biz"
	"videoCluster/internal/central/service"
)

type emptyNodeSource struct{}

func (emptyNodeSource) ListHealthyNodes(context.Context) ([]biz.Node, error) {
	return nil, nil
}

func TestVideoServiceListsVideosFromHealthyNodes(t *testing.T) {
	videoService := service.NewVideoService(biz.NewVideoAggregator(emptyNodeSource{}))

	response, err := videoService.ListVideos(context.Background(), &v1.ListVideosRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Videos) != 0 {
		t.Fatalf("videos = %d, want 0", len(response.Videos))
	}
}
