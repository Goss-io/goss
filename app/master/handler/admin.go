package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Goss-io/Goss/lib/ini"
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
func (this *AdminService) Start() {
	r := gin.Default()
	r.Static("/img", "./admin/static/img/")
	r.Static("/css", "./admin/static/css/")
	r.Static("/vendor", "./admin/static/vendor/")
	r.LoadHTMLGlob("./admin/views/*")

	r.GET("/console", this.handleConsole)
	if err := r.Run(this.WebPort); err != nil {
		log.Panicln(err)
	}
}

//handleConsole .
func (this *AdminService) handleConsole(c *gin.Context) {

	//获取所有的api节点.
	apiList := GetApiList()

	//获取所有的存储节点.
	storageList := GetStorageList()

	c.HTML(http.StatusOK, "console.html", map[string]interface{}{"apiList": apiList, "storageList": storageList})
}
