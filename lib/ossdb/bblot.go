package ossdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Goss-io/goss/lib"
	"github.com/Goss-io/goss/lib/logd"

	"go.etcd.io/bbolt"
)

//BucketInfo.
type BucketInfo struct {
	Name       string
	Host       string
	CreateTime time.Time
}

func NewDB(path string) (db *bbolt.DB, err error) {
	db, err = bbolt.Open(fmt.Sprintf("%s", "goss.db"), 0777, nil)
	if err != nil {
		logd.Make(logd.Level_ERROR, logd.GetLogpath(), err.Error())
		return db, err
	}

	return db, nil
}

//Read.
func Read(db *bbolt.DB, bktname string, key string) (buf []byte, err error) {
	err = db.View(func(tx *bbolt.Tx) error {
		bkt := tx.Bucket([]byte(bktname))
		buf = bkt.Get([]byte(key))
		return nil
	})

	if err != nil {
		logd.Make(logd.Level_ERROR, logd.GetLogpath(), err.Error())
		return buf, err
	}
	return buf, err
}

func CreateBucket(db *bbolt.DB, bktinfo BucketInfo) error {
	return db.Update(func(tx *bbolt.Tx) error {
		//判断当前bucket是否存在.
		list, err := BucketList(db)
		if err != nil {
			logd.Make(logd.Level_ERROR, logd.GetLogpath(), err.Error())
			return err
		}

		if lib.InArray(bktinfo.Name, list) {
			return errors.New("当前bucket已经存在")
		}

		//创建bucket并且记录bucket信息.
		bkt, err := tx.CreateBucket([]byte(bktinfo.Name))
		if err != nil {
			logd.Make(logd.Level_ERROR, logd.GetLogpath(), err.Error())
			return err
		}

		b, err := json.Marshal(bktinfo)
		if err != nil {
			logd.Make(logd.Level_ERROR, logd.GetLogpath(), err.Error())
			return err
		}

		if err := bkt.Put([]byte(bktinfo.Name), b); err != nil {
			logd.Make(logd.Level_ERROR, logd.GetLogpath(), err.Error())
			return err
		}

		return nil
	})
}

//BucketList .
func BucketList(db *bbolt.DB) (list []string, err error) {
	err = db.View(func(tx *bbolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bbolt.Bucket) error {
			list = append(list, string(name))
			return nil
		})
	})
	if err != nil {
		return list, err
	}
	return list, nil
}

//Insert.
func Insert(db *bbolt.DB, bktname string, key string, val string) error {
	return db.Update(func(tx *bbolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists([]byte(bktname))
		if err != nil {
			logd.Make(logd.Level_ERROR, logd.GetLogpath(), err.Error())
			return err
		}

		return bkt.Put([]byte(key), []byte(val))
	})
}
