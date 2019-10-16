package handler

import (
	"encoding/json"
	"errors"
	"net"
	"time"

	"github.com/Goss-io/goss/lib"
	"github.com/Goss-io/goss/lib/hardware"
	"github.com/Goss-io/goss/lib/ini"
	"github.com/Goss-io/goss/lib/logd"
	"github.com/Goss-io/goss/lib/packet"
	"github.com/Goss-io/goss/lib/protocol"
)

//connMaster 连接管理节点.
func (s *StorageService) connMaster() {
	//上报节点信息
	conn := s.conn(s.MasterNode)

	//连接初始化.
	if err := s.connInit(conn); err != nil {
		logd.Make(logd.Level_WARNING, logd.GetLogpath(), err.Error())
		s.connMaster()
		return
	}

	for {
		var buf = make([]byte, 1024)
		_, err := conn.Read(buf)
		if err != nil {
			s.connMaster()
			return
		}
	}
}

func (s *StorageService) conn(node string) net.Conn {
	conn, err := net.Dial("tcp4", node)
	if err != nil {
		logd.Make(logd.Level_WARNING, logd.GetLogpath(), "master节点连接失败，稍后重新连接")
		time.Sleep(time.Second * 1)
		return s.conn(node)
	}

	return conn
}

//connInit 连接初始化.
func (s *StorageService) connInit(conn net.Conn) error {
	if err := s.sendAuth(conn); err != nil {
		return err
	}

	if err := s.sendNodeInfo(conn); err != nil {
		return err
	}
	return nil
}

//auth 发送授权信息.
func (s *StorageService) sendAuth(conn net.Conn) error {
	tokenBuf := []byte(ini.GetString("token"))
	buf := packet.New(tokenBuf, tokenBuf, protocol.CONN_AUTH)
	_, err := conn.Write(buf)
	if err != nil {
		return errors.New("授权信息发送失败")
	}

	pkt, err := packet.Parse(conn)
	if err != nil {
		return errors.New("授权失败")
	}

	if string(pkt.Body) == "fail" {
		return errors.New("授权信息验证失败")
	}

	return nil
}

//sendNodeInfo 上报节点信息.
func (s *StorageService) sendNodeInfo(conn net.Conn) error {
	h := hardware.New()
	nodeInfo := protocol.NodeInfo{
		Types:    string(packet.NodeTypes_Storage),
		CpuNum:   h.Cpu.Num,
		MemSize:  h.Mem.Total,
		SourceIP: s.Addr,
		Name:     ini.GetString("node_name"),
	}

	content, err := json.Marshal(nodeInfo)
	if err != nil {
		return err
	}

	hashBuf := lib.Hash(string(content))
	buf := packet.New(content, hashBuf, protocol.REPORT_NODE_INFO)
	_, err = conn.Write(buf)
	if err != nil {
		return err
	}

	pkt, err := packet.Parse(conn)
	if err != nil {
		return err
	}

	if string(pkt.Body) == "fail" {
		return errors.New("发送节点信息失败")
	}

	logd.Make(logd.Level_INFO, logd.GetLogpath(), "上报节点信息成功")
	return nil
}
