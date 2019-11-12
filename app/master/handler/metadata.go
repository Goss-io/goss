package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Goss-io/goss/lib/logd"

	"go.etcd.io/bbolt"

	"github.com/Goss-io/goss/app/master/conf"
	"github.com/Goss-io/goss/lib/ossdb"
)

type MetadataService struct {
	Port string
	DB   *bbolt.DB
}

//NewMetadata.
func NewMetadata() {
	m := &MetadataService{
		Port: fmt.Sprintf(":%d", conf.Conf.Node.MetadataPort),
	}
	m.Start()
}

//Start .
func (m *MetadataService) Start() {
	if err := m.newdb(); err != nil {
		logd.Make(logd.Level_ERROR, logd.GetLogpath(), err.Error())
		return
	}
	m.newhttp()
}

func (m *MetadataService) newhttp() {
	http.HandleFunc("/medatadta/get", m.MetadataGet)
	http.HandleFunc("/medatadta/set", m.MetadataSet)
	http.HandleFunc("/bucket/list", m.BucketList)
	http.HandleFunc("/bucket/set", m.BucketSet)
	http.HandleFunc("/bucket/get", m.BucketGet)
	if err := http.ListenAndServe(m.Port, nil); err != nil {
		logd.Make(logd.Level_ERROR, logd.GetLogpath(), err.Error())
		return
	}
}

func (m *MetadataService) newdb() error {
	db, err := ossdb.NewDB(conf.Conf.Node.MetadataPath)
	if err != nil {
		logd.Make(logd.Level_ERROR, logd.GetLogpath(), err.Error())
		return err
	}
	m.DB = db
	return nil
}

//MetadataGet.
func (m *MetadataService) MetadataGet(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	if token != conf.Conf.Node.Token {
		w.Write([]byte("fail"))
		return
	}
	key := r.FormValue("key")
	buf, err := ossdb.Read(m.DB, "", key)
	if err != nil {
		logd.Make(logd.Level_ERROR, logd.GetLogpath(), err.Error())
		return
	}
	w.Write(buf)
}

//MetadataSet .
func (m *MetadataService) MetadataSet(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	if token != conf.Conf.Node.Token {
		w.Write([]byte("fail"))
		return
	}

	key := r.FormValue("key")
	val := r.FormValue("value")

	if err := ossdb.Insert(m.DB, "", key, val); err != nil {
		logd.Make(logd.Level_ERROR, logd.GetLogpath(), err.Error())
		w.Write([]byte("fail"))
		return
	}

	w.Write([]byte("success"))
}

//BucketList  .
func (m *MetadataService) BucketList(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	if token != conf.Conf.Node.Token {
		w.Write([]byte("fail"))
		return
	}
	bktlist, err := ossdb.BucketList(m.DB)
	if err != nil {
		logd.Make(logd.Level_ERROR, logd.GetLogpath(), err.Error())
		w.Write([]byte("fail"))
		return
	}

	b, err := json.Marshal(bktlist)
	if err != nil {
		logd.Make(logd.Level_ERROR, logd.GetLogpath(), err.Error())
		w.Write([]byte("fail"))
		return
	}
	w.Write(b)
}

//BucketSet  .
func (m *MetadataService) BucketSet(w http.ResponseWriter, r *http.Request) {

}

//BucketGet  .
func (m *MetadataService) BucketGet(w http.ResponseWriter, r *http.Request) {

}
