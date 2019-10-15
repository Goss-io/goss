package main

import (
	"log"

	"github.com/Goss-io/Goss/app/master/conf"
	"github.com/Goss-io/Goss/app/master/handler"
	"github.com/Goss-io/Goss/lib/cmd"
)

func main() {
	cmd := cmd.New()

	//加载配置文件.
	conf.Load(cmd)
	log.Println("node name:", conf.Conf.Node.Name)

	master := handler.NewMaster()
	master.Start()
}
