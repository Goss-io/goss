package handler

import (
	"github.com/Goss-io/goss/lib"
	"github.com/Goss-io/goss/lib/packet"
	"github.com/Goss-io/goss/lib/protocol"
)

//GossNode 节点信息.
var GossNode = []Node{}

type Node struct {
	Types    packet.NodeTypes
	IP       string
	SourceIP string //所属ip.
	CreateAt string
	Name     string
	CpuNum   string
	MemSize  string
}

//GetStoreList 获取所有的存储节点.
func GetStorageList() []Node {
	list := []Node{}
	for _, v := range GossNode {
		if v.Types == packet.NodeTypes_Storage {
			list = append(list, v)
		}
	}

	return list
}

//获取所有的api节点.
func GetApiList() []Node {
	list := []Node{}
	for _, v := range GossNode {
		if v.Types == packet.NodeTypes_Api {
			list = append(list, v)
		}
	}

	return list
}

//RemoveNode 移除某一个切片.
func RemoveNode(n *MasterService, ip string) {
	//根据访问ip获取节点ip.
	for index, v := range GossNode {
		if v.IP == ip {
			//通知对应的节点与故障节点断开连接.
			pkt := packet.New([]byte(v.SourceIP), lib.Hash(v.SourceIP), protocol.REMOVE_NODE)
			n.Conn[v.SourceIP].Write(pkt)

			//从NodeInfo中移除当前.
			GossNode = append(GossNode[:index], GossNode[index+1:]...)

			//删除对应的连接数据.
			delete(n.Conn, v.SourceIP)
		}
	}
}
