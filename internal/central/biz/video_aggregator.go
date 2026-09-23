package biz

import (
	"context"
	"fmt"
	"sync"
	"time"

	centralv1 "videoCluster/api/central/v1"
	nodev1 "videoCluster/api/node/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// Node is a healthy Node endpoint returned by service discovery.
type Node struct {
	ID      string
	Address string
}

// NodeSource returns only the Nodes Consul currently considers healthy.
type NodeSource interface {
	ListHealthyNodes(context.Context) ([]Node, error)
}

// VideoAggregator reads video metadata directly from healthy Nodes.
type VideoAggregator struct {
	nodes NodeSource
}

func NewVideoAggregator(nodes NodeSource) *VideoAggregator {
	return &VideoAggregator{nodes: nodes}
}

func (a *VideoAggregator) ListVideos(ctx context.Context) ([]*centralv1.VideoMeta, error) {
	nodes, err := a.nodes.ListHealthyNodes(ctx)
	if err != nil {
		return nil, err
	}

	type result struct {
		videos []*centralv1.VideoMeta
		err    error
	}
	results := make(chan result, len(nodes))
	var wg sync.WaitGroup
	for _, node := range nodes {
		node := node
		wg.Add(1)
		go func() {
			defer wg.Done()
			videos, err := listNodeVideos(ctx, node)
			results <- result{videos: videos, err: err}
		}()
	}
	wg.Wait()
	close(results)

	var videos []*centralv1.VideoMeta
	var failed int
	for result := range results {
		if result.err != nil {
			failed++
			continue
		}
		videos = append(videos, result.videos...)
	}
	if len(nodes) > 0 && failed == len(nodes) {
		return nil, status.Error(codes.Unavailable, "no healthy Node could provide video metadata")
	}
	return videos, nil
}

func (a *VideoAggregator) GetVideoInfo(ctx context.Context, id string) (*centralv1.VideoMeta, error) {
	nodes, err := a.nodes.ListHealthyNodes(ctx)
	if err != nil {
		return nil, err
	}
	for _, node := range nodes {
		video, err := getNodeVideo(ctx, node, id)
		if err == nil && video != nil {
			return video, nil
		}
	}
	return nil, status.Errorf(codes.NotFound, "video %q was not found on a healthy Node", id)
}

func listNodeVideos(ctx context.Context, node Node) ([]*centralv1.VideoMeta, error) {
	client, closeClient, err := nodeClient(ctx, node.Address)
	if err != nil {
		return nil, err
	}
	defer closeClient()

	response, err := client.ListVideos(ctx, &nodev1.ListVideosRequest{})
	if err != nil {
		return nil, err
	}
	videos := make([]*centralv1.VideoMeta, 0, len(response.Videos))
	for _, video := range response.Videos {
		videos = append(videos, centralVideo(video, node.Address))
	}
	return videos, nil
}

func getNodeVideo(ctx context.Context, node Node, id string) (*centralv1.VideoMeta, error) {
	client, closeClient, err := nodeClient(ctx, node.Address)
	if err != nil {
		return nil, err
	}
	defer closeClient()

	response, err := client.GetVideoInfo(ctx, &nodev1.GetVideoInfoRequest{Id: id})
	if err != nil || response.Video == nil {
		return nil, err
	}
	return centralVideo(response.Video, node.Address), nil
}

func nodeClient(ctx context.Context, address string) (nodev1.NodeServiceClient, func() error, error) {
	dialCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	connection, err := grpc.DialContext(dialCtx, address, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return nil, nil, fmt.Errorf("connect to Node %s: %w", address, err)
	}
	return nodev1.NewNodeServiceClient(connection), connection.Close, nil
}

func centralVideo(video *nodev1.VideoMeta, nodeAddress string) *centralv1.VideoMeta {
	return &centralv1.VideoMeta{
		Id:        video.Id,
		Title:     video.Title,
		Path:      video.Path,
		Size:      video.Size,
		Thumbnail: video.Thumbnail,
		NodeAddr:  nodeAddress,
	}
}
