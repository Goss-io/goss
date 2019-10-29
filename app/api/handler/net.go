package handler

import (
	"errors"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/Goss-io/goss/lib/logd"

	"github.com/Goss-io/goss/lib/ini"

	"github.com/Goss-io/goss/lib"

	"github.com/Goss-io/goss/lib/packet"
	"github.com/Goss-io/goss/lib/protocol"
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
func (t *TcpService) Start(addr string) {
	go t.connStorageNode(addr)
}

//connStorageNode 连接存储节点.
func (t *TcpService) connStorageNode(addr string) {
	//判断当前节点是否已经连接.
	_, ok := t.conn[addr]
	if ok {
		return
	}
	log.Println("开始连接:", addr)
	conn := t.connection(addr)

	//建立授权.
	if err := t.sendAuth(conn); err != nil {
		log.Printf("err:%+v\n", err)
		return
	}
	t.conn[addr] = conn
	log.Println(addr, "连接成功!")
}

//auth 发送授权信息.
func (t *TcpService) sendAuth(conn net.Conn) error {
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
func (t *TcpService) connection(addr string) net.Conn {
	conn, err := net.Dial("tcp4", addr)
	if err != nil {
		log.Printf("%s:节点连接失败, 尝试重新连接!%+v\n", addr, err)
		time.Sleep(time.Second * 1)
		return t.connection(addr)
	}
	return conn
}

//SelectStoreNode 选择存储节点.
func (t *TcpService) SelectStoreNode() (nodeip string, conn net.Conn) {
	nodeipList := t.SelectNode(1)
	addr := nodeipList[0]
	return addr, t.conn[addr]
}

//SelectNode 选择节点.
//excludeipList 为排除的ip.
func (t *TcpService) SelectNode(nodenum int, excludeipList ...string) []string {
	rand.Seed(time.Now().UnixNano())
	list := []string{}
	for k, _ := range t.conn {
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
func (t *TcpService) Write(buf []byte, ip string) (storePath string, err error) {
	conn := t.conn[ip]
	_, err = conn.Write(buf)
	if err != nil {
		return storePath, err
	}

	log.Println("发送成功")
	pkt, err := packet.Parse(conn)
	if err != nil {
		return storePath, err
	}

	log.Printf("pkt:%+v\n", pkt)
	if string(pkt.Body) == "fail" {
		return storePath, errors.New("失败")
	}

	return string(pkt.Body), err
}

//Read tcp读取文件.
func (t *TcpService) Read(nodeip, fPath string) (boby []byte, err error) {
	// 建立连接.
	// conn, err := net.Dial("tcp4", nodeip)
	// if err != nil {
	// 	log.Printf("%+v\n", err)
	// 	return boby, err
	// }

	// //连接授权.
	// token := ini.GetString("token")
	// buf := packet.New([]byte(token), lib.Hash(token), protocol.CONN_AUTH)
	// _, err = conn.Write(buf)
	// log.Println("nodeip:", nodeip)
	// var conn net.Conn
	// for k, v := range t.conn {
	// 	if k == "127.0.0.1:9001" {
	// 		conn = v
	// 	}
	// }
	conn := t.conn["127.0.0.1:9001"]
	// log.Println(conn)
	// log.Println("conn")
	// conn := t.conn["127.0.0.1:9001"]
	// if err != nil {
	// 	log.Printf("%+v\n", err)
	// 	return boby, err
	// }
	// log.Printf("conn:%+v\n", conn)
	// pkt, err := packet.Parse(conn)
	// if err != nil {
	// 	log.Printf("%+v\n", err)
	// 	return boby, err
	// }
	// log.Printf("pkt:%+v\n", pkt)

	// if string(pkt.Body) == "fail" {
	// 	return nil, errors.New("文件获取失败")
	// }

	//读取文件.
	// log.Println("fPath:", fPath)
	buf := packet.New([]byte(fPath), lib.Hash(fPath), protocol.READ_FILE)
	_, err = conn.Write(buf)
	if err != nil {
		log.Printf("%+v\n", err)
		return boby, err
	}

	// log.Printf("buf:%+v\n", string(buf))
	// log.Println("write")

	pkt, err := packet.Parse(conn)
	if err != nil {
		log.Printf("%+v\n", err)
		return boby, err
	}

	// log.Println("over")
	return pkt.Body, nil
}
