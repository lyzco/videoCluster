package data

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/dgraph-io/badger/v4"
	v1 "videoCluster/api/node/v1"
	"videoCluster/internal/node/biz"
)

type VideoVideoRepository struct {
	db *badger.DB
}

type storedVideo struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Path        string `json:"path"`
	Size        int64  `json:"size"`
	Thumbnail   string `json:"thumbnail"`
	ScanRoot    string `json:"scanRoot,omitempty"`
	ScanVersion string `json:"scanVersion,omitempty"`
}

func NewVideoVideoRepository(db *badger.DB) *VideoVideoRepository {
	return &VideoVideoRepository{db: db}
}

func (r *VideoVideoRepository) Save(ctx context.Context, video *v1.VideoMeta) error {
	return r.save(video, "", "")
}

func (r *VideoVideoRepository) SaveScanned(ctx context.Context, video *v1.VideoMeta, scanRoot, scanVersion string) error {
	return r.save(video, scanRoot, scanVersion)
}

func (r *VideoVideoRepository) save(video *v1.VideoMeta, scanRoot, scanVersion string) error {
	data, err := json.Marshal(storedVideo{
		ID:          video.Id,
		Title:       video.Title,
		Path:        video.Path,
		Size:        video.Size,
		Thumbnail:   video.Thumbnail,
		ScanRoot:    scanRoot,
		ScanVersion: scanVersion,
	})
	if err != nil {
		return err
	}
	return r.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(video.Id), data)
	})
}

func (r *VideoVideoRepository) DeleteUnseen(ctx context.Context, scanRoot, scanVersion string) error {
	var ids []string
	err := r.db.View(func(txn *badger.Txn) error {
		iter := txn.NewIterator(badger.DefaultIteratorOptions)
		defer iter.Close()
		for iter.Rewind(); iter.Valid(); iter.Next() {
			item := iter.Item()
			value, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			var video storedVideo
			if err := json.Unmarshal(value, &video); err != nil {
				return err
			}
			if video.ScanRoot == scanRoot && video.ScanVersion != scanVersion {
				ids = append(ids, video.ID)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return r.db.Update(func(txn *badger.Txn) error {
		for _, id := range ids {
			if err := txn.Delete([]byte(id)); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *VideoVideoRepository) Delete(ctx context.Context, id string) error {
	return r.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(id))
	})
}

func (r *VideoVideoRepository) Get(ctx context.Context, id string) (*v1.VideoMeta, error) {
	var video storedVideo
	err := r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(id))
		if err != nil {
			return err
		}
		val, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		return json.Unmarshal(val, &video)
	})
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil, biz.ErrVideoNotFound
		}
		return nil, err
	}
	return video.meta(), nil
}

func (r *VideoVideoRepository) ListAll(ctx context.Context) ([]*v1.VideoMeta, error) {
	var list []*v1.VideoMeta
	err := r.db.View(func(txn *badger.Txn) error {
		iter := txn.NewIterator(badger.DefaultIteratorOptions)
		defer iter.Close()
		for iter.Rewind(); iter.Valid(); iter.Next() {
			item := iter.Item()
			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			var video storedVideo
			if err := json.Unmarshal(val, &video); err != nil {
				return err
			}
			list = append(list, video.meta())
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (v storedVideo) meta() *v1.VideoMeta {
	return &v1.VideoMeta{
		Id:        v.ID,
		Title:     v.Title,
		Path:      v.Path,
		Size:      v.Size,
		Thumbnail: v.Thumbnail,
	}
}
