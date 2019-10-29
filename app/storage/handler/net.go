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
	Auth       map[string]bool
}

func NewStorageService() *StorageService {
	s := &StorageService{
		Port:       fmt.Sprintf(":%d", conf.Conf.Node.Port),
		Addr:       fmt.Sprintf("%s:%d", ini.GetString("node_ip"), ini.GetInt("node_port")),
		MasterNode: ini.GetString("master_node"),
		Auth:       make(map[string]bool),
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

// func (s *StorageService) handler(conn net.Conn, ip string) {
// 	defer conn.Close()
// 	for {
//判断是否已经授权.
// mx := sync.RWMutex{}
// mx.RLock()
// _, ok := s.Auth[ip]
// mx.RUnlock()
// if !ok {
// 	log.Println("sAuth[ip]:", s.Auth[ip])
// 	buf := packet.New([]byte("未授权"), lib.Hash("未授权"), protocol.MSG)
// 	conn.Write(buf)
// 	return
// }
// pkt, err := packet.Parse(conn)
// if err != nil {
// 	log.Printf("err:%+v\n", err)
// 	logd.Make(logd.Level_WARNING, logd.GetLogpath(), ip+"断开连接")
// 	return
// }
// log.Printf("pkt:%+v\n", pkt)

//判断协议号.
// 		if pkt.Protocol == protocol.SEND_FILE {
// 			//计算文件hash.
// 			fHash := lib.FileHash(pkt.Body)
// 			//验证文件是否损坏.
// 			if fHash != pkt.Hash {
// 				logd.Make(logd.Level_WARNING, logd.GetLogpath(), "文件hash不一致")
// 				buf := packet.New([]byte("fail"), lib.Hash("fail"), protocol.MSG)
// 				conn.Write(buf)
// 				return
// 			}

// 			fPath := s.SelectPath(fHash) + fHash
// 			log.Println("fPath:", fPath)
// 			storePath := conf.Conf.Node.StorageRoot + fPath
// 			log.Println("storePath:", storePath)
// 			err = ioutil.WriteFile(storePath, pkt.Body, 0777)
// 			if err != nil {
// 				log.Printf("err:%+v\n", err)
// 				logd.Make(logd.Level_WARNING, logd.GetLogpath(), "创建文件失败"+err.Error())
// 				buf := packet.New([]byte("fail"), lib.Hash("fail"), protocol.MSG)
// 				conn.Write(buf)
// 				return
// 			}
// 			buf := packet.New([]byte(fPath), lib.Hash(fHash), protocol.MSG)
// 			conn.Write(buf)
// 		}

// 		if pkt.Protocol == protocol.READ_FILE {
// 			//读取文件.
// 			fpath := conf.Conf.Node.StorageRoot + string(pkt.Body)
// 			b, err := ioutil.ReadFile(fpath)
// 			if err != nil {
// 				log.Printf("err:%+v\n", err)
// 				logd.Make(logd.Level_WARNING, logd.GetLogpath(), "读取文件失败:"+err.Error())
// 				buf := packet.New([]byte("fail"), lib.Hash("fail"), protocol.MSG)
// 				conn.Write(buf)
// 				return
// 			}

// 			//验证文件是否损坏.
// 			// if lib.FileHash(b) != pkt.Hash {
// 			// 	logd.Make(logd.Level_WARNING, logd.GetLogpath(), pkt.Hash+"文件已损坏")
// 			// 	buf := packet.New([]byte("fail"), lib.Hash("fail"), protocol.MSG)
// 			// 	conn.Write(buf)
// 			// 	return
// 			// }

// 			buf := packet.New(b, []byte(pkt.Hash), protocol.SEND_FILE)
// 			_, err = conn.Write(buf)
// 			if err != nil {
// 				log.Printf("err:%+v\n", err)
// 				logd.Make(logd.Level_WARNING, logd.GetLogpath(), "文件发送失败:"+err.Error())
// 				buf := packet.New([]byte("fail"), lib.Hash("fail"), protocol.MSG)
// 				conn.Write(buf)
// 				return
// 			}

// 			logd.Make(logd.Level_INFO, logd.GetLogpath(), "文件发送成功")
// 		}

// 	}
// }
