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
func (this *ApiService) Start() {
	go this.connMaster()
	this.httpSrv()
}

//httpSrv .
func (this *ApiService) httpSrv() {
	http.HandleFunc("/oss/", this.handler)
	if err := http.ListenAndServe(this.Port, nil); err != nil {
		log.Panicf("%+v\n", err)
	}
}

//handler .
func (this *ApiService) handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		this.get(w, r)
		return
	}

	if r.Method == http.MethodPut {
		this.put(w, r)
		return
	}

	if r.Method == http.MethodDelete {
		this.delete(w, r)
		return
	}

	w.WriteHeader(http.StatusNotFound)
}

//get.
func (this *ApiService) get(w http.ResponseWriter, r *http.Request) {
	name, err := this.getParse(r.URL.EscapedPath())
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	meta := db.Metadata{
		Name: name,
	}
	if err = meta.Query(); err != nil {
		log.Printf("%+v\n", err)
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	b, err := this.Tcp.Read(meta.StoreNode, meta.Hash, meta.Size)
	if err != nil {
		log.Printf("%+v\n", err)
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Write([]byte(b))
}

//getParse get请求解析文件名.
func (this *ApiService) getParse(url string) (name string, err error) {
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
func (this *ApiService) put(w http.ResponseWriter, r *http.Request) {
	//获取文件名称，文件大小，文件类型，文件hash.
	//元数据.
	name, err := this.getParse(r.URL.EscapedPath())
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
	_, nodeip, err := this.Tcp.Write(pkt)
	if err != nil {
		log.Printf("%+v\n", err)
		w.Write([]byte("fail"))
		return
	}

	//记录三条元数据，一条当前这条元数据可用, 其余两条元数据不可用.
	//选择三个存储节点(节点不能相同).
	nodeipList := this.Tcp.SelectNode(2, nodeip)
	//开启事物操作，防止节点数据不一致.
	tx := db.Db.Begin()
	log.Printf("nodeipList:%+v\n", nodeipList)

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

	//记录需要异步拉取数据的节点元数据.
	for _, ip := range nodeipList {
		metadata = db.Metadata{
			Name:      name,
			Type:      ft,
			Size:      int64(len(fBody)),
			Hash:      fhash,
			StoreNode: ip,
			Usable:    false,
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
		tx.Rollback()
		return
	}

	w.Write([]byte("success"))
}

//delete.
func (this *ApiService) delete(w http.ResponseWriter, r *http.Request) {

}
