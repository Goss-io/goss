package ossdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/Goss-io/goss/lib"
	"github.com/Goss-io/goss/lib/logd"

	"go.etcd.io/bbolt"
)

//BucketInfo.
type BucketInfo struct {
	Name       string `form:"name" json:"name"`
	Host       string `form:"host" json:"host"`
	CreateTime string `form:"create_time" json:"create_time"`
}

func NewDB(path string) (db *bbolt.DB, err error) {
	path = fmt.Sprintf("%sgoss.db", path)
	log.Println("path:", path)
	db, err = bbolt.Open(path, 0777, nil)
	if err != nil {
		logd.Make(logd.Level_ERROR, logd.GetLogpath(), err.Error())
		return db, err
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("goss_bucket_info"))
		if err != nil {
			logd.Make(logd.Level_ERROR, logd.GetLogpath(), err.Error())
			return err
		}
		return nil
	})
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
		bktlist, err := BucketList(db)
		if err != nil {
			logd.Make(logd.Level_ERROR, logd.GetLogpath(), err.Error())
			return err
		}
		list := []string{}
		for _, v := range bktlist {
			list = append(list, v.Name)
		}

		if lib.InArray(bktinfo.Name, list) {
			return errors.New("当前bucket已经存在")
		}

		//创建bucket
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

		//记录当前bucket的信息.
		bkt, err = tx.CreateBucketIfNotExists([]byte("goss_bucket_info"))
		if err != nil {
			logd.Make(logd.Level_ERROR, logd.GetLogpath(), err.Error())
			return err
		}

		if err := bkt.Put([]byte(bktinfo.Name), b); err != nil {
			logd.Make(logd.Level_ERROR, logd.GetLogpath(), err.Error())
			return err
		}

		log.Println("put bktinfo:", string(b))
		return nil
	})
}

//BucketList .
func BucketList(db *bbolt.DB) (list []BucketInfo, err error) {
	err = db.View(func(tx *bbolt.Tx) error {
		bkt := tx.Bucket([]byte("goss_bucket_info"))
		return bkt.ForEach(func(name []byte, val []byte) error {
			bktinfo := BucketInfo{}
			if err := json.Unmarshal(val, &bktinfo); err != nil {
				logd.Make(logd.Level_ERROR, logd.GetLogpath(), err.Error())
				return err
			}
			list = append(list, bktinfo)
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
