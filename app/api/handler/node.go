package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log"
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
func (this *ApiService) connMaster() {
	conn := this.conn(this.MasterNode)
	//连接初始化
	if err := this.connInit(conn); err != nil {
		logd.Make(logd.Level_WARNING, logd.GetLogpath(), err.Error())
		time.Sleep(time.Second * 1)
		this.connMaster()
		return
	}

	for {
		pkt, err := packet.Parse(conn)
		if err != nil && err == io.EOF {
			logd.Make(logd.Level_WARNING, logd.GetLogpath(), "master节点断开连接，稍后重新连接master节点")
			this.connMaster()
			return
		}
		log.Printf("pkt:%+v\n", pkt)
		log.Printf("pkt.body:%+v\n", string(pkt.Body))

		//判断协议类型.
		if pkt.Protocol == protocol.ADD_NODE {
			ip := string(pkt.Body)
			log.Println("ip:", ip)
			//新增节点.
			logd.Make(logd.Level_INFO, logd.GetLogpath(), "新增存储节点:"+ip)
			this.Tcp.Start(ip)
		}

		if pkt.Protocol == protocol.REMOVE_NODE {
			ip := string(pkt.Body)
			log.Println("ip:", ip)
			//删除节点.
			logd.Make(logd.Level_INFO, logd.GetLogpath(), "收到master节点要求与:"+ip+"节点断开的消息")
			if err := this.Tcp.conn[ip].Close(); err != nil {
				logd.Make(logd.Level_INFO, logd.GetLogpath(), "断开与:"+ip+"节点的连接失败")
				return
			}
			logd.Make(logd.Level_INFO, logd.GetLogpath(), "断开成功")
		}
	}
}

//conn .
func (this *ApiService) conn(node string) net.Conn {
	conn, err := net.Dial("tcp4", node)
	if err != nil {
		logd.Make(logd.Level_WARNING, logd.GetLogpath(), "master节点连接失败，稍后重新连接")
		time.Sleep(time.Second * 1)
		return this.conn(node)
	}

	return conn
}

//connInit 连接初始化.
func (this *ApiService) connInit(conn net.Conn) error {
	//向主节点发送授权信息.
	if err := this.sendAuth(conn); err != nil {
		return err
	}

	//发送节点信息.
	if err := this.sendNodeInfo(conn); err != nil {
		return err
	}
	return nil
}

//auth 发送授权信息.
func (this *ApiService) sendAuth(conn net.Conn) error {
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
func (this *ApiService) sendNodeInfo(conn net.Conn) error {
	h := hardware.New()
	nodeInfo := protocol.NodeInfo{
		Types:    string(packet.NodeTypes_Api),
		CpuNum:   h.Cpu.Num,
		MemSize:  h.Mem.Total,
		SourceIP: this.Addr,
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
