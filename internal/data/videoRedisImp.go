package data

import (
	"context"
	"github.com/redis/go-redis/v9"
	v1 "videoCluster/api/node/v1"
	"videoCluster/internal/common"
	"videoCluster/internal/conf"
)

type VideoRepo struct {
	redis *redis.Client
	key   string
}

func NewVideoRepo(redis *redis.Client, c *conf.Server) *VideoRepo {
	return &VideoRepo{redis: redis, key: c.ServerName + ":" + c.ServerId + ":"}
}

func (r *VideoRepo) ListAllVideos(ctx context.Context) ([]*v1.VideoMeta, error) {
	var cursor uint64
	var videos []*v1.VideoMeta

	for {
		keys, nextCursor, err := r.redis.Scan(ctx, cursor, r.key+"*", 100).Result()
		if err != nil {
			return nil, err
		}

		for _, key := range keys {
			data, err := r.redis.HGetAll(ctx, key).Result()
			if err != nil {
				continue
			}

			size, _ := common.ConvertToInt64(data["size"])

			v := &v1.VideoMeta{
				Id:        data["id"],
				Title:     data["title"],
				Path:      data["path"],
				Size:      size,
				Thumbnail: data["thumbnail"],
			}

			videos = append(videos, v)
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return videos, nil
}

func (r *VideoRepo) SaveVideo(ctx context.Context, video *v1.VideoMeta) error {
	return r.redis.HSet(ctx, r.key+video.Id, "id", video.Id, "title", video.Title, "path", video.Path, "size", video.Size, "thumbnail", video.Thumbnail).Err()
}

func (r *VideoRepo) DeleteVideo(ctx context.Context, id string) (int64, error) {
	return r.redis.Del(ctx, r.key+id).Result()
}

func (r *VideoRepo) GetVideoInfo(ctx context.Context, id string) (map[string]string, error) {
	return r.redis.HGetAll(ctx, r.key+id).Result()
}
