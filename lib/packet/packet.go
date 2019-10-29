package packet

import (
	"encoding/binary"
	"io"
	"net"

	"github.com/Goss-io/goss/lib/protocol"
)

const PROROCOL_LEN = 4
const HEADER_LEN = 4
const HASH_LEN = 32

type Packet struct {
	Protocol protocol.GossProtocol
	Size     int64
	Hash     string
	Body     []byte
}

func New(content, fileHash []byte, num protocol.GossProtocol) []byte {
	buffer := make([]byte, HEADER_LEN+len(content)+len(fileHash)+PROROCOL_LEN)
	//0-4 为协议号.
	//4-8 为文件大小.
	//8-40 为文件hash.
	//>40 为文件内容.
	binary.BigEndian.PutUint32(buffer[0:4], uint32(num))
	binary.BigEndian.PutUint32(buffer[4:8], uint32(len(content)))
	copy(buffer[8:40], fileHash)
	copy(buffer[40:], content)
	return buffer
}

//Parse 解析网络数据包.
func Parse(conn net.Conn) (pkt Packet, err error) {
	//获取协议号.
	var num = make([]byte, PROROCOL_LEN)
	_, err = io.ReadFull(conn, num)
	if err != nil {
		return pkt, err
	}
	pkt.Protocol = protocol.GossProtocol(int(binary.BigEndian.Uint32(num)))

	//获取文件长度.
	var header = make([]byte, HEADER_LEN)
	_, err = io.ReadFull(conn, header)
	if err != nil {
		return pkt, err
	}
	pkt.Size = int64(binary.BigEndian.Uint32(header))

	//获取hash.
	var fhash = make([]byte, HASH_LEN)
	_, err = io.ReadFull(conn, fhash)
	if err != nil {
		return pkt, err
	}
	pkt.Hash = string(fhash)

	//获取文件
	var buf = make([]byte, pkt.Size)
	_, err = io.ReadFull(conn, buf)
	if err != nil {
		return pkt, err
	}
	pkt.Body = buf

	return pkt, nil
}
