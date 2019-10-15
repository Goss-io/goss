package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/Goss-io/goss/lib"
	"github.com/Goss-io/goss/lib/ini"
	"github.com/Goss-io/goss/lib/logd"
	"github.com/Goss-io/goss/lib/packet"
	"github.com/Goss-io/goss/lib/protocol"
)

type NodeParams struct {
	Conn  net.Conn
	Types packet.NodeTypes
}

type MasterService struct {
	Conn map[string]net.Conn
	Auth map[string]bool
	Port string
}

//NewMaster .
func NewMaster() *MasterService {
	return &MasterService{
		Conn: make(map[string]net.Conn),
		Auth: make(map[string]bool),
		Port: fmt.Sprintf(":%d", ini.GetInt("node_port")),
	}
}

//Start.
func (this *MasterService) Start() {
	go NewAdmin()
	this.listen()
	select {}
}

//listen .
func (this *MasterService) listen() {
	listener, err := net.Listen("tcp4", this.Port)
	if err != nil {
		log.Panicln(err)
	}

	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			logd.Make(logd.Level_WARNING, logd.GetLogpath(), err.Error())
			continue
		}

		//验证授权信息.
		ip := conn.RemoteAddr().String()
		if err := this.connInit(conn, ip); err != nil {
			logd.Make(logd.Level_WARNING, logd.GetLogpath(), err.Error())
			continue
		}

		logd.Make(logd.Level_INFO, logd.GetLogpath(), "收到来自:"+ip+"的连接请求")
		go this.handler(ip, conn)
	}
}

//connInit 连接初始化.
func (this *MasterService) connInit(conn net.Conn, ip string) error {
	//验证授权信息.
	if err := this.checkAuth(conn, ip); err != nil {
		buf := packet.New([]byte("fail"), lib.Hash("fail"), protocol.MSG)
		conn.Write(buf)
		return err
	}
	buf := packet.New([]byte("success"), lib.Hash("success"), protocol.MSG)
	_, err := conn.Write(buf)
	if err != nil {
		return err
	}

	//接收节点信息.
	if err := this.parseNodeInfo(conn, ip); err != nil {
		buf := packet.New([]byte("fail"), lib.Hash("fail"), protocol.MSG)
		conn.Write(buf)
		return err
	}
	return nil
}

//parseNodeInfo .
func (this *MasterService) parseNodeInfo(conn net.Conn, ip string) error {
	pkt, err := packet.Parse(conn)
	if err != nil {
		return err
	}
	//判读协议类型.
	if pkt.Protocol == protocol.REPORT_NODE_INFO {
		n := protocol.NodeInfo{}
		if err = json.Unmarshal(pkt.Body, &n); err != nil {
			return err
		}

		node := Node{
			Name:     n.Name,
			SourceIP: n.SourceIP,
			CpuNum:   n.CpuNum,
			MemSize:  n.MemSize,
			IP:       ip,
			CreateAt: lib.Time(),
			Types:    packet.NodeTypes(n.Types),
		}

		GossNode = append(GossNode, node)
		this.Conn[node.SourceIP] = conn

		buf := packet.New([]byte("success"), lib.Hash("success"), protocol.MSG)
		_, err = conn.Write(buf)
		if err != nil {
			return err
		}

		//新存储节点上线,通知所有的api节点.
		if node.Types == packet.NodeTypes_Storage {
			//通知api节点.
			apiList := GetApiList()
			for _, v := range apiList {
				pkt := packet.New([]byte(node.SourceIP), lib.Hash(node.SourceIP), protocol.ADD_NODE)
				_, err = this.Conn[v.SourceIP].Write(pkt)
				if err != nil {
					logd.Make(logd.Level_WARNING, logd.GetLogpath(), "通知api节点:"+node.SourceIP+"新增storage节点失败，稍后重新通知")
					RemoveNode(this, ip)
					return err
				}

				logd.Make(logd.Level_INFO, logd.GetLogpath(), "通知api节点，新增存储节点:"+node.SourceIP+"成功")
			}
		}

		if node.Types == packet.NodeTypes_Api {
			//告知新上线的api节点多有的storage节点ip.
			storageList := GetStorageList()
			for _, v := range storageList {
				pktMsg := packet.New([]byte(v.SourceIP), lib.Hash(v.SourceIP), protocol.ADD_NODE)
				_, err = this.Conn[node.SourceIP].Write(pktMsg)
				if err != nil {
					logd.Make(logd.Level_WARNING, logd.GetLogpath(), "通知api节点:"+v.SourceIP+"storage节点失败，稍后重新通知")
					RemoveNode(this, ip)
					return err
				}

				logd.Make(logd.Level_INFO, logd.GetLogpath(), "通知api节点，新增存储节点:"+v.SourceIP+"成功")
			}
		}
	}
	return nil
}

//checkAuth .
func (this *MasterService) checkAuth(conn net.Conn, ip string) error {
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

	this.Auth[ip] = true
	return nil
}

//handler .
func (this *MasterService) handler(ip string, conn net.Conn) {
	defer conn.Close()
	for {
		//验证是否已经授权.
		if !this.Auth[ip] {
			conn.Write([]byte("fail"))
			return
		}

		pkt, err := packet.ParseNode(conn)
		if err != nil && err == io.EOF {
			logd.Make(logd.Level_WARNING, logd.GetLogpath(), ip+"断开连接")
			//从节点列表中移除.
			RemoveNode(this, ip)
			return
		}

		//判断协议.
		if pkt.Protocol == protocol.ADD_NODE {
			//新增节点信息.
			info := Node{
				Types:    pkt.Types,
				IP:       ip,
				SourceIP: pkt.IP,
				CreateAt: lib.Time(),
			}
			GossNode = append(GossNode, info)
		}
	}
}
