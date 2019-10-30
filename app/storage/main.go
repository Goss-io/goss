package main

import (
	"log"

	"github.com/Goss-io/goss/app/storage/conf"
	"github.com/Goss-io/goss/app/storage/handler"
	"github.com/Goss-io/goss/lib/cmd"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	cmd := cmd.New()

	//加载配置文件.
	conf.Load(cmd)
	log.Println("node name:", conf.Conf.Node.Name)

	storage := handler.NewStorageService()
	storage.Start()
}
