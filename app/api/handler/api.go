package handler

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/Goss-io/goss/lib/ini"

	"github.com/Goss-io/goss/db"
	"github.com/Goss-io/goss/lib/filetype"
	"github.com/Goss-io/goss/lib/protocol"

	"github.com/Goss-io/goss/app/api/conf"
	"github.com/Goss-io/goss/lib"
	"github.com/Goss-io/goss/lib/packet"
)

//ApiService.
type ApiService struct {
	Port       string
	Tcp        *TcpService
	Addr       string
	MasterNode string
	// Backups chan
}

// type Backups

//NewApi .
func NewApi() *ApiService {
	cf := conf.Conf.Node
	apiSrv := ApiService{
		Port:       fmt.Sprintf(":%d", cf.Port),
		Tcp:        NewTcpService(),
		Addr:       fmt.Sprintf("%s:%d", ini.GetString("node_ip"), ini.GetInt("node_port")),
		MasterNode: ini.GetString("master_node"),
	}
	return &apiSrv
}

//Start .
func (a *ApiService) Start() {
	go a.connMaster()
	a.httpSrv()
}

//httpSrv .
func (a *ApiService) httpSrv() {
	http.HandleFunc("/oss/", a.handler)
	if err := http.ListenAndServe(a.Port, nil); err != nil {
		log.Panicf("%+v\n", err)
	}
}

//handler .
func (a *ApiService) handler(w http.ResponseWriter, r *http.Request) {
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
func (a *ApiService) get(w http.ResponseWriter, r *http.Request) {
	name, err := a.getParse(r.URL.EscapedPath())
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	meta := db.Metadata{
		Name: name,
	}
	list, err := meta.QueryNodeIP()
	if err != nil {
		log.Printf("%+v\n", err)
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if len(list) < 1 {
		w.Write([]byte("not found"))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	buf := make(chan []byte, meta.Size)
	for _, nodeip := range list {
		b, err := a.Tcp.Read(nodeip, meta.Hash)
		if err != nil {
			log.Printf("%+v\n", err)
			continue
		}
		buf <- b
		break
	}

	msg := <-buf
	w.Write(msg)
}

//getParse get请求解析文件名.
func (a *ApiService) getParse(url string) (name string, err error) {
	sArr := strings.Split(url, "/")
	if len(sArr) != 3 {
		return name, errors.New("not fount")
	}
	if sArr[2] == "" {
		return name, errors.New("not fount")
	}

	return sArr[2], nil
}

//put.
func (a *ApiService) put(w http.ResponseWriter, r *http.Request) {
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

	fhash := lib.FileHash(fBody)
	pkt := packet.New(fBody, []byte(fhash), protocol.SEND_FILE)

	//采用用强一致性来记录文件.
	nodeipList := a.Tcp.SelectNode(3)
	log.Printf("nodeipList:%+v\n", nodeipList)

	//开启事物操作，防止节点数据不一致.
	tx := db.Db.Begin()
	for _, nodeip := range nodeipList {
		err := a.Tcp.Write(pkt, nodeip)
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
			Usable:    true,
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
func (a *ApiService) delete(w http.ResponseWriter, r *http.Request) {

}
