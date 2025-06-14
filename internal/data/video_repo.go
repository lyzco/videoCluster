package data

import (
	"context"
	"encoding/json"
	"github.com/dgraph-io/badger/v4"
	v1 "videoCluster/api/node/v1"
)

type VideoVideoRepository struct {
	db *badger.DB
}

func NewVideoVideoRepository(db *badger.DB) *VideoVideoRepository {
	return &VideoVideoRepository{db: db}
}

func (r *VideoVideoRepository) Save(ctx context.Context, video *v1.VideoMeta) error {
	data, _ := json.Marshal(video)
	return r.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(video.Id), data)
	})
}

func (r *VideoVideoRepository) Delete(ctx context.Context, id string) error {
	return r.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(id))
	})
}

func (r *VideoVideoRepository) Get(ctx context.Context, id string) (*v1.VideoMeta, error) {
	var video v1.VideoMeta
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
		return nil, err
	}
	return &video, nil
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
			var video v1.VideoMeta
			if err := json.Unmarshal(val, &video); err != nil {
				return err
			}
			list = append(list, &video)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return list, nil
}
