package biz_test

import (
	"context"
	"net"
	"testing"

	nodev1 "videoCluster/api/node/v1"
	"videoCluster/internal/central/biz"

	"google.golang.org/grpc"
)

type staticNodeSource []biz.Node

func (s staticNodeSource) ListHealthyNodes(context.Context) ([]biz.Node, error) {
	return s, nil
}

type nodeService struct {
	nodev1.UnimplementedNodeServiceServer
}

func (nodeService) ListVideos(context.Context, *nodev1.ListVideosRequest) (*nodev1.ListVideosReply, error) {
	return &nodev1.ListVideosReply{Videos: []*nodev1.VideoMeta{{Id: "video-1", Title: "Example"}}}, nil
}

func TestVideoAggregatorListsVideosFromHealthyNode(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	grpcServer := grpc.NewServer()
	nodev1.RegisterNodeServiceServer(grpcServer, nodeService{})
	go func() { _ = grpcServer.Serve(listener) }()
	t.Cleanup(func() {
		grpcServer.Stop()
		_ = listener.Close()
	})

	aggregator := biz.NewVideoAggregator(staticNodeSource{{ID: "node-1", Address: listener.Addr().String()}})
	videos, err := aggregator.ListVideos(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(videos) != 1 || videos[0].Id != "video-1" || videos[0].NodeAddr != listener.Addr().String() {
		t.Fatalf("videos = %#v", videos)
	}
}
