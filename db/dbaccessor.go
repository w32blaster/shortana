package db

import (
	"fmt"
	"go.etcd.io/bbolt"
)

const (
	urlBucket = "MyBucket"
)

type Database struct {
	db *bbolt.DB
}

func Init(bdb *bbolt.DB) (*Database, error) {

	if err := bdb.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(urlBucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &Database{
		db: bdb,
	}, nil
}

func (d Database) GetAll() map[string]string {
	var mapRes map[string]string
	err := d.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(urlBucket))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			mapRes[string(k)] = string(v)
		}

		return nil
	})

	if err != nil {
		return mapRes
	}

	return mapRes
}
