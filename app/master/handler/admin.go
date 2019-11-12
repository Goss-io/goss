package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/Goss-io/goss/lib/ossdb"

	"github.com/Goss-io/goss/app/master/conf"

	"github.com/Goss-io/goss/lib/ini"
	"github.com/gin-gonic/gin"
)

type AdminService struct {
	WebPort string
}

func NewAdmin() {
	adm := &AdminService{
		WebPort: fmt.Sprintf(":%d", ini.GetInt("node_web_port")),
	}
	adm.Start()
}

//Start .
func (a *AdminService) Start() {
	r := gin.Default()
	r.Static("/img", "./admin/static/img/")
	r.Static("/css", "./admin/static/css/")
	r.Static("/vendor", "./admin/static/vendor/")
	r.LoadHTMLGlob("./admin/views/*")

	r.GET("/admin/login", a.handleConsole)
	r.GET("/admin/console", a.handleConsole)
	r.GET("/admin/bucket", a.handleBucketList)
	r.GET("/admin/monitor", a.handleMonitor)
	if err := r.Run(a.WebPort); err != nil {
		log.Panicln(err)
	}
}

//handleConsole .
func (a *AdminService) handleConsole(c *gin.Context) {
	//获取所有的api节点.
	apiList := GetApiList()

	//获取所有的存储节点.
	storageList := GetStorageList()

	obj := map[string]interface{}{
		"apiList":     apiList,
		"storageList": storageList,
		"menu":        "console",
	}
	c.HTML(http.StatusOK, "console.html", obj)
}

//handleBucketList  bucket list.
func (a *AdminService) handleBucketList(c *gin.Context) {
	//获取存储桶列表.
	list, err := a.getBucketList()
	if err != nil {
		log.Printf("err123:%+v\n", err)
		return
	}

	obj := map[string]interface{}{
		"menu":       "bucket",
		"bucketList": list,
	}
	c.HTML(http.StatusOK, "bucket_list.html", obj)
}

func (a *AdminService) getBucketList() (list []string, err error) {
	cf := conf.Conf.Node
	bkturl := fmt.Sprintf("http://%s:%d/bucket/list", cf.IP, cf.MetadataPort)
	req, err := http.NewRequest("GET", bkturl, nil)
	if err != nil {
		return list, err
	}
	req.Header.Set("token", cf.Token)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("%+v\n", err)
		return list, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("%+v\n", err)
		return list, err
	}

	if string(b) == "fail" {
		return list, errors.New("fail")
	}
	bktlist := []ossdb.BucketInfo{}
	if err := json.Unmarshal(b, &bktlist); err != nil {
		return list, err
	}

	return list, nil
}

//handleMonitor  monitor.
func (a *AdminService) handleMonitor(c *gin.Context) {
	obj := map[string]interface{}{
		"menu": "monitor",
	}
	c.HTML(http.StatusOK, "monitor.html", obj)
}
