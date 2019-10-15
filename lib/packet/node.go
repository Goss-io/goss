package packet

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"github.com/Goss-io/Goss/lib/protocol"
)

type NodeTypes string

const (
	NodeTypes_Api     NodeTypes = "api"
	NodeTypes_Storage           = "stprage"
	NodeTypes_Master            = "master"
)

//NodePacket 节点管理数据.
type NodePacket struct {
	Protocol protocol.GossProtocol //协议号.
	Types    NodeTypes             //节点类型.
	IP       string                //节点ip.
	Size     int64
}

//NewNode.
func NewNode(types NodeTypes, ip string, proto protocol.GossProtocol) []byte {
	body := fmt.Sprintf("%s,%s", types, ip)
	buffer := make([]byte, HEADER_LEN+len(body)+PROROCOL_LEN)
	//0-4 为协议号.
	//4-8 为数据大小.
	//>8 为节点数据.
	binary.BigEndian.PutUint32(buffer[0:4], uint32(proto))
	binary.BigEndian.PutUint32(buffer[4:8], uint32(len(body)))
	copy(buffer[8:], body)
	return buffer
}

func ParseNode(conn net.Conn) (pkt NodePacket, err error) {
	pkt = NodePacket{}
	//获取协议号.
	var num = make([]byte, PROROCOL_LEN)
	_, err = io.ReadFull(conn, num)
	if err != nil {
		log.Printf("%+v\n", err)
		return pkt, err
	}
	pkt.Protocol = protocol.GossProtocol(int(binary.BigEndian.Uint32(num)))

	//获取数据长度.
	var bodyBuf = make([]byte, 4)
	_, err = io.ReadFull(conn, bodyBuf)
	if err != nil {
		log.Printf("%+v\n", err)
		return pkt, err
	}
	pkt.Size = int64(binary.BigEndian.Uint32(bodyBuf))

	//获取数据内容.
	var body = make([]byte, pkt.Size)
	_, err = io.ReadFull(conn, body)
	if err != nil {
		log.Printf("%+v\n", err)
		return pkt, err
	}

	//拆分字符串.
	sArr := strings.Split(string(body), ",")
	pkt.Types = NodeTypes(sArr[0])
	pkt.IP = sArr[1]

	return pkt, nil
}
