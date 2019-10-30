package handler

import (
	"fmt"

	"github.com/Goss-io/goss/lib/logd"

	"github.com/Goss-io/goss/lib/ini"

	"github.com/Goss-io/goss/app/storage/conf"
)

type StorageService struct {
	Port       string
	Addr       string
	MasterNode string
}

func NewStorageService() *StorageService {
	s := &StorageService{
		Port:       fmt.Sprintf(":%d", conf.Conf.Node.Port),
		Addr:       fmt.Sprintf("%s:%d", ini.GetString("node_ip"), ini.GetInt("node_port")),
		MasterNode: ini.GetString("master_node"),
	}
	return s
}

//Start .
func (s *StorageService) Start() {
	s.checkStoragePath()
	go s.connMaster()
	s.httpSrv()
}

//checkStoragePath 检查存储路径.
func (s *StorageService) checkStoragePath() {
	logd.Make(logd.Level_INFO, logd.GetLogpath(), "初始化存储路径")
	if err := s.InitStoragePath(conf.Conf.Node.StorageRoot); err != nil {
		panic(err)
	}
}
