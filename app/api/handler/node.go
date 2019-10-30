package handler

import (
	"encoding/json"
	"errors"
	"io"
	"math/rand"
	"net"
	"time"

	"github.com/Goss-io/goss/lib"
	"github.com/Goss-io/goss/lib/hardware"
	"github.com/Goss-io/goss/lib/ini"
	"github.com/Goss-io/goss/lib/logd"
	"github.com/Goss-io/goss/lib/packet"
	"github.com/Goss-io/goss/lib/protocol"
)

//connMaster .
func (a *APIService) connMaster() {
	conn := a.conn(a.MasterNode)
	//连接初始化
	if err := a.conninit(conn); err != nil {
		logd.Make(logd.Level_WARNING, logd.GetLogpath(), err.Error())
		time.Sleep(time.Second * 1)
		a.connMaster()
		return
	}

	for {
		pkt, err := packet.Parse(conn)
		if err != nil && err == io.EOF {
			logd.Make(logd.Level_WARNING, logd.GetLogpath(), "master节点断开连接，稍后重新连接master节点")
			a.connMaster()
			return
		}

		//判断协议类型.
		if pkt.Protocol == protocol.ADD_NODE {
			ip := string(pkt.Body)
			//新增节点.
			logd.Make(logd.Level_INFO, logd.GetLogpath(), "新增存储节点:"+ip)
			a.Storage = append(a.Storage, ip)
		}

		if pkt.Protocol == protocol.REMOVE_NODE {
			ip := string(pkt.Body)
			//删除节点.
			logd.Make(logd.Level_INFO, logd.GetLogpath(), "收到master节点要求与:"+ip+"节点断开的消息")
			a.RemoveStorageNode(ip)
			logd.Make(logd.Level_INFO, logd.GetLogpath(), "断开成功")
		}
	}
}

func (a *APIService) RemoveStorageNode(nodeip string) {
	for index, v := range a.Storage {
		if v == nodeip {
			a.Storage = append(a.Storage[:index], a.Storage[index+1:]...)
		}
	}
}

//conn .
func (a *APIService) conn(node string) net.Conn {
	conn, err := net.Dial("tcp4", node)
	if err != nil {
		logd.Make(logd.Level_WARNING, logd.GetLogpath(), "master节点连接失败，稍后重新连接")
		time.Sleep(time.Second * 1)
		return a.conn(node)
	}

	return conn
}

//connInit 连接初始化.
func (a *APIService) conninit(conn net.Conn) error {
	//向主节点发送授权信息.
	if err := a.sendAuth(conn); err != nil {
		return err
	}

	//发送节点信息.
	if err := a.sendNodeInfo(conn); err != nil {
		return err
	}
	return nil
}

//auth 发送授权信息.
func (a *APIService) sendAuth(conn net.Conn) error {
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

	logd.Make(logd.Level_INFO, logd.GetLogpath(), "授权成功")
	return nil
}

//sendNodeInfo 上报节点信息.
func (a *APIService) sendNodeInfo(conn net.Conn) error {
	h := hardware.New()
	nodeInfo := protocol.NodeInfo{
		Types:    string(packet.NodeTypes_Api),
		CpuNum:   h.Cpu.Num,
		MemSize:  h.Mem.Total,
		SourceIP: a.Addr,
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

func (a *APIService) SelectNode(nodenum int) []string {
	rand.Seed(time.Now().UnixNano())
	list := []string{}
	for _, v := range a.Storage {
		list = append(list, v)
	}

	nodeipList := []string{}
	num := 0
	for {
		if num >= nodenum {
			break
		}
		index := rand.Int() % len(list)
		addr := list[index]

		if !lib.InArray(addr, nodeipList) {
			num++
			nodeipList = append(nodeipList, addr)
		}
	}
	return nodeipList
}
