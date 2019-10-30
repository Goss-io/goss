package handler

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Goss-io/goss/app/api/conf"
	"github.com/Goss-io/goss/db"
	"github.com/Goss-io/goss/lib"
	"github.com/Goss-io/goss/lib/filetype"
	"github.com/Goss-io/goss/lib/ini"
)

//NewAPI .
func NewAPI() *APIService {
	cf := conf.Conf.Node
	apiSrv := APIService{
		Port:       fmt.Sprintf(":%d", cf.Port),
		Addr:       fmt.Sprintf("%s:%d", ini.GetString("node_ip"), ini.GetInt("node_port")),
		MasterNode: ini.GetString("master_node"),
	}
	return &apiSrv
}

//Start .
func (a *APIService) Start() {
	go a.connMaster()
	a.httpSrv()
}

//httpSrv .
func (a *APIService) httpSrv() {
	http.HandleFunc("/", a.handler)
	if err := http.ListenAndServe(a.Port, nil); err != nil {
		log.Panicf("%+v\n", err)
	}
}

//handler .
func (a *APIService) handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		a.get(w, r)
		return
	}

	if r.Method == http.MethodPut {
		a.put(w, r)
		return
	}

	if r.Method == http.MethodDelete {
		a.delete(w, r)
		return
	}

	w.WriteHeader(http.StatusNotFound)
}

//get.
func (a *APIService) get(w http.ResponseWriter, r *http.Request) {
	//验证bucket是否存在.
	bkt := db.Bucket{
		Host: r.Host,
	}
	if err := bkt.Query(); err != nil {
		log.Printf("err:%+v\n", err)
		w.Write([]byte(err.Error()))
		return
	}
	if bkt.ID < 1 {
		w.Write([]byte("不存在"))
		return
	}

	//获取访问的文件.
	name, err := a.getParse(r.URL.EscapedPath())
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	meta := db.Metadata{
		Name:     name,
		BucketID: bkt.ID,
	}
	list, err := meta.QueryNodeIP()
	if err != nil {
		log.Printf("%+v\n", err)
		w.Write([]byte(err.Error()))
		return
	}
	if len(list) < 1 {
		w.Write([]byte("not found"))
		return
	}

	buf := make(chan []byte, meta.Size)
	var errnum = 0
	for _, nodeip := range list {
		b, err := a.Read(meta.StorePath, nodeip)
		if err != nil {
			errnum++
			log.Printf("%+v\n", err)
			continue
		}
		buf <- b
		break
	}

	msg := <-buf
	//如果msg为空的话，则判断是否errnum > 0.
	if len(msg) > 0 {
		w.Header().Set("Content-Type", meta.Type)
		_, err = w.Write(msg)
		if err != nil {
			log.Printf("err:%+v\n", err)
		}
		return
	}
	if errnum > 0 {
		w.Write([]byte("获取失败"))
		return
	}
	w.Write([]byte("not found"))
}

//getParse get请求解析文件名.
//兼容目录结构，host以后的路径都为文件名.
func (a *APIService) getParse(url string) (name string, err error) {
	path := strings.TrimLeft(url, "/")
	if len(path) < 1 {
		return name, errors.New("not fount")
	}

	return path, nil
}

//put.
func (a *APIService) put(w http.ResponseWriter, r *http.Request) {
	//验证bucket是否存在.
	bkt := db.Bucket{
		Host: r.Host,
	}
	if err := bkt.Query(); err != nil {
		log.Printf("err:%+v\n", err)
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if bkt.ID < 1 {
		w.Write([]byte("不存在"))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	//验证AccessKey和SecretKey是否正确.
	ak := r.Header.Get("AccessKey")
	sk := r.Header.Get("SecretKey")
	if len(ak) != 32 || len(sk) != 32 || bkt.AccessKey != ak || bkt.SecretKey != sk {
		w.Write([]byte("授权失败"))
		w.WriteHeader(http.StatusForbidden)
		return
	}

	//获取文件名称，文件大小，文件类型，文件hash.
	//元数据.
	name, err := a.getParse(r.URL.EscapedPath())
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	fBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("%+v\n", err)
		w.Write([]byte("fail"))
		return
	}

	//获取文件类型.
	f16 := fmt.Sprintf("%x", fBody)
	ft := filetype.Parse(f16[:10])

	//计算文件hash.
	fhash := lib.FileHash(fBody)

	//采用用强一致性来记录文件.
	nodeipList := a.SelectNode(3)
	log.Printf("nodeipList:%+v\n", nodeipList)

	//开启事物操作，防止节点数据不一致.
	tx := db.Db.Begin()
	for _, nodeip := range nodeipList {
		storePath, err := a.Write(fhash, fBody, nodeip)
		if err != nil {
			log.Printf("%+v\n", err)
			w.Write([]byte("fail"))
			tx.Rollback()
			//todo 删除已经记录的文件.
			return
		}

		//记录文件元数据.
		metadata := db.Metadata{
			Name:      name,
			Type:      ft,
			Size:      int64(len(fBody)),
			Hash:      fhash,
			StoreNode: nodeip,
			StorePath: storePath,
			Usable:    true,
			BucketID:  bkt.ID,
		}
		if err = tx.Create(&metadata).Error; err != nil {
			log.Printf("%+v\n", err)
			w.Write([]byte("fail"))
			tx.Rollback()
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("%+v\n", err)
		w.Write([]byte("fail"))
		return
	}

	w.Write([]byte("success"))
}

//delete.
func (a *APIService) delete(w http.ResponseWriter, r *http.Request) {

}

//Write 发送消息.
func (a *APIService) Write(fhash string, body []byte, nodeip string) (storePath string, err error) {
	url := fmt.Sprintf("http://%s/", nodeip)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		log.Printf("err:%+v\n", err)
		return storePath, err
	}
	req.Header.Set("token", conf.Conf.Node.Token)
	req.Header.Set("fhash", fhash)
	client := http.Client{}
	response, err := client.Do(req)
	if err != nil {
		log.Printf("err:%+v\n", err)
		return storePath, err
	}

	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("err:%+v\n", err)
		return storePath, err
	}

	resp := lib.ParseMsg(b)
	if !resp.Status {
		return storePath, errors.New(resp.Msg.(string))
	}

	return resp.Msg.(string), nil
}

//Read 读取消息.
func (a *APIService) Read(fpath, nodeip string) (fbody []byte, err error) {
	url := fmt.Sprintf("http://%s/", nodeip)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Printf("err:%+v\n", err)
		return fbody, err
	}
	req.Header.Set("fpath", fpath)
	req.Header.Set("token", conf.Conf.Node.Token)
	client := http.Client{
		Timeout: time.Second * 1,
	}
	response, err := client.Do(req)
	if err != nil {
		log.Printf("err:%+v\n", err)
		return fbody, nil
	}

	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("err:%+v\n", err)
		return fbody, nil
	}

	resp := lib.ParseMsg(b)
	if !resp.Status {
		return fbody, errors.New(resp.Msg.(string))
	}
	return resp.Body, nil
}
