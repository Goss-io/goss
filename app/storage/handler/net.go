package handler

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"

	"github.com/Goss-io/Goss/lib/logd"

	"github.com/Goss-io/Goss/lib/ini"

	"github.com/Goss-io/Goss/lib/protocol"

	"github.com/Goss-io/Goss/app/storage/conf"
	"github.com/Goss-io/Goss/lib"
	"github.com/Goss-io/Goss/lib/packet"
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
func (this *StorageService) Start() {
	this.checkStoragePath()
	go this.connMaster()
	this.listen()
}

//checkStoragePath 检查存储路径.
func (this *StorageService) checkStoragePath() {
	if !lib.IsExists(conf.Conf.Node.StorageRoot) {
		//创建存储文件夹.
		if err := os.Mkdir(conf.Conf.Node.StorageRoot, 0777); err != nil {
			log.Panicf("%+v\n", err)
		}
	}
}

//listen .
func (this *StorageService) listen() {
	listener, err := net.Listen("tcp4", this.Port)
	if err != nil {
		log.Printf("端口监听失败!%+v\n", err)
		return
	}

	for {
		conn, err := listener.Accept()
		if err != nil && err == io.EOF {
			logd.Make(logd.Level_WARNING, logd.GetLogpath(), "断开连接")
			return
		}

		ip := conn.RemoteAddr().String()
		if err := this.checkAuth(conn, ip); err != nil {
			logd.Make(logd.Level_WARNING, logd.GetLogpath(), err.Error())
			return
		}
		log.Println("api节点：" + ip + "连接")
		go this.handler(conn, ip)
	}
}

//checkAuth .
func (this *StorageService) checkAuth(conn net.Conn, ip string) error {
	pkt, err := packet.Parse(conn)
	if err != nil {
		return err
	}

	//判读协议.
	if pkt.Protocol != protocol.CONN_AUTH {
		return errors.New("协议错误")
	}

	//验证授权信息是否正确.
	if string(pkt.Body) != ini.GetString("token") {
		return errors.New("授权失败")
	}

	buf := packet.New([]byte("success"), lib.Hash("success"), protocol.MSG)
	_, err = conn.Write(buf)
	if err != nil {
		return err
	}
	this.Auth[ip] = true
	return nil
}

func (this *StorageService) handler(conn net.Conn, ip string) {
	defer conn.Close()
	for {
		//判断是否已经授权.
		if !this.Auth[ip] {
			buf := packet.New([]byte("未授权"), lib.Hash("未授权"), protocol.MSG)
			conn.Write(buf)
			return
		}
		pkt, err := packet.Parse(conn)
		if err != nil && err == io.EOF {
			logd.Make(logd.Level_WARNING, logd.GetLogpath(), ip+"断开连接")
			return
		}

		//判断协议号.
		if pkt.Protocol == protocol.SEND_FILE {
			//计算文件hash.
			fHash := lib.FileHash(pkt.Body)
			//验证文件是否损坏.
			if fHash != pkt.Hash {
				logd.Make(logd.Level_WARNING, logd.GetLogpath(), "文件hash不一致")
				conn.Write([]byte("fail"))
				return
			}

			fPath := conf.Conf.Node.StorageRoot + fHash
			err = ioutil.WriteFile(fPath, pkt.Body, 0777)
			if err != nil {
				logd.Make(logd.Level_WARNING, logd.GetLogpath(), "创建文件失败"+err.Error())
				return
			}
			conn.Write([]byte(fHash))
		}

		if pkt.Protocol == protocol.READ_FILE {
			//读取文件.
			fpath := conf.Conf.Node.StorageRoot + pkt.Hash
			b, err := ioutil.ReadFile(fpath)
			if err != nil {
				logd.Make(logd.Level_WARNING, logd.GetLogpath(), "读取文件失败:"+err.Error())
				return
			}

			//验证文件是否损坏.
			if lib.FileHash(b) != pkt.Hash {
				logd.Make(logd.Level_WARNING, logd.GetLogpath(), pkt.Hash+"文件已损坏")
				return
			}

			buf := packet.New(b, []byte(pkt.Hash), protocol.SEND_FILE)
			_, err = conn.Write(buf)
			if err != nil && err == io.EOF {
				logd.Make(logd.Level_WARNING, logd.GetLogpath(), "文件发送失败:"+err.Error())
				return
			}

			logd.Make(logd.Level_INFO, logd.GetLogpath(), "文件发送成功")
		}
	}
}
