package data

import (
	"context"
	"net"
	"strconv"

	"github.com/hashicorp/consul/api"
	"videoCluster/internal/central/biz"
)

const nodeServiceName = "videoCluster-media-node"

// NodeRepository reads healthy Node endpoints from Consul.
type NodeRepository struct {
	client *api.Client
}

func NewNodeRepository(client *api.Client) *NodeRepository {
	return &NodeRepository{client: client}
}

func (r *NodeRepository) ListHealthyNodes(ctx context.Context) ([]biz.Node, error) {
	entries, _, err := r.client.Health().Service(nodeServiceName, "", true, (&api.QueryOptions{}).WithContext(ctx))
	if err != nil {
		return nil, err
	}
	nodes := make([]biz.Node, 0, len(entries))
	for _, entry := range entries {
		address := entry.Service.Address
		if address == "" {
			address = entry.Node.Address
		}
		if address == "" || entry.Service.Port == 0 {
			continue
		}
		nodes = append(nodes, biz.Node{
			ID:      entry.Service.ID,
			Address: net.JoinHostPort(address, strconv.Itoa(entry.Service.Port)),
		})
	}
	return nodes, nil
}
