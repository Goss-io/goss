package main

import (
	"log"

	"github.com/Goss-io/goss/app/api/conf"
	"github.com/Goss-io/goss/app/api/handler"
	"github.com/Goss-io/goss/db"
	"github.com/Goss-io/goss/lib/cmd"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	cmd := cmd.New()

	//加载配置文件.
	conf.Load(cmd)
	log.Println("node name:", conf.Conf.Node.Name)

	if err := db.Connection(); err != nil {
		log.Panicln(err)
	}

	apiSrv := handler.NewApi()
	apiSrv.Start()
}
