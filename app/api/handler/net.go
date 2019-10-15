package handler

import (
	"errors"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/Goss-io/Goss/lib/logd"

	"github.com/Goss-io/Goss/lib/ini"

	"github.com/Goss-io/Goss/lib"

	"github.com/Goss-io/Goss/lib/packet"
	"github.com/Goss-io/Goss/lib/protocol"
)

type TcpService struct {
	conn map[string]net.Conn
}

//NewTcpService .
func NewTcpService() *TcpService {
	return &TcpService{
		conn: make(map[string]net.Conn),
	}
}

//Start .
func (this *TcpService) Start(addr string) {
	go this.connStorageNode(addr)
}

//connStorageNode 连接存储节点.
func (this *TcpService) connStorageNode(addr string) {
	//判断当前节点是否已经连接.
	_, ok := this.conn[addr]
	if ok {
		return
	}
	log.Println("开始连接:", addr)
	conn := this.connection(addr)

	//建立授权.
	if err := this.sendAuth(conn); err != nil {
		log.Printf("err:%+v\n", err)
		return
	}
	this.conn[addr] = conn
	log.Println(addr, "连接成功!")
}

//auth 发送授权信息.
func (this *TcpService) sendAuth(conn net.Conn) error {
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

//connection .
func (this *TcpService) connection(addr string) net.Conn {
	conn, err := net.Dial("tcp4", addr)
	if err != nil {
		log.Printf("%s:节点连接失败, 尝试重新连接!%+v\n", addr, err)
		time.Sleep(time.Second * 1)
		return this.connection(addr)
	}
	return conn
}

//SelectStoreNode 选择存储节点.
func (this *TcpService) SelectStoreNode() (nodeip string, conn net.Conn) {
	nodeipList := this.SelectNode(1)
	addr := nodeipList[0]
	return addr, this.conn[addr]
}

//SelectNode 选择节点.
//excludeipList 为排除的ip.
func (this *TcpService) SelectNode(nodenum int, excludeipList ...string) []string {
	rand.Seed(time.Now().UnixNano())
	list := []string{}
	for k, _ := range this.conn {
		list = append(list, k)
	}

	nodeipList := []string{}
	num := 0
	for {
		if num >= nodenum {
			break
		}
		index := rand.Int() % len(list)
		addr := list[index]

		//判读当前ip是否需要排除.
		if lib.InArray(addr, excludeipList) {
			continue
		}
		if !lib.InArray(addr, nodeipList) {
			num++
			nodeipList = append(nodeipList, addr)
		}
	}
	return nodeipList
}

//Write tcp 发送消息.
func (this *TcpService) Write(b []byte) (msg []byte, nodeip string, err error) {
	nodeip, conn := this.SelectStoreNode()
	_, err = conn.Write(b)
	if err != nil {
		log.Printf("%+v\n", err)
		return msg, nodeip, err
	}

	for {
		var buf = make([]byte, 1024)
		_, err = conn.Read(buf)
		if err != nil {
			log.Printf("%+v\n", err)
			return msg, nodeip, err
		}

		return buf, nodeip, nil
	}
}

//Read tcp读取文件.
func (this *TcpService) Read(nodeip, fHash string, bodylen int64) (boby []byte, err error) {
	//建立连接.
	conn, err := net.Dial("tcp4", nodeip)
	if err != nil {
		log.Printf("%+v\n", err)
		return boby, err
	}

	//连接授权.
	token := ini.GetString("token")
	buf := packet.New([]byte(token), lib.Hash(token), protocol.CONN_AUTH)
	_, err = conn.Write(buf)
	if err != nil {
		log.Printf("%+v\n", err)
		return boby, err
	}
	pkt, err := packet.Parse(conn)
	if err != nil {
		log.Printf("%+v\n", err)
		return boby, err
	}

	if string(pkt.Body) == "fail" {
		return nil, errors.New("文件获取失败")
	}

	//读取文件.
	buf = packet.New(nil, []byte(fHash), protocol.READ_FILE)
	_, err = conn.Write(buf)
	if err != nil {
		log.Printf("%+v\n", err)
		return boby, err
	}

	pkt, err = packet.Parse(conn)
	if err != nil {
		log.Printf("%+v\n", err)
		return boby, err
	}

	return pkt.Body, nil
}
