package protocol

//GossProtocol 传输协议号.
type GossProtocol int

const (
	CONN_AUTH        GossProtocol = 1000 //连接授权.
	REPORT_NODE_INFO GossProtocol = 1001 //上报节点信息.
	MSG              GossProtocol = 1002 //发送消息.
	ADD_NODE         GossProtocol = 1003 //新增节点.
	REMOVE_NODE      GossProtocol = 1004 //删除节点.
	SEND_FILE        GossProtocol = 1005 //发送文件.
	READ_FILE        GossProtocol = 1006 //读取文件.
)

//NodeInfo 节点信息.
type NodeInfo struct {
	Types    string `json:"types"`
	CpuNum   string `json:"cpu_num"`
	MemSize  string `json:"mem_size"`
	SourceIP string `json:"source_ip"`
	Name     string `json:"name"`
}
