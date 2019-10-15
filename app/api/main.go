package main

import (
	"log"

	"github.com/Goss-io/Goss/app/api/conf"
	"github.com/Goss-io/Goss/app/api/handler"
	"github.com/Goss-io/Goss/db"
	"github.com/Goss-io/Goss/lib/cmd"
)

func main() {
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
