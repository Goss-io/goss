package handler

//basicStruct.go
//该文件仅存储各种消息结构与标识信息.

const (
	//StatusOK 返回正常标识码.
	StatusOK int8 = iota
	//StatusError 返回错误标识码.
	StatusError
)

//Response 消息返回信息.
type Response struct {
	Status  int8        `json:"status"`
	Msg     string      `json:"msg"`
	Content interface{} `json:"content"`
}

//StorageService 存储节点信息.
type StorageService struct {
	Port       string
	Addr       string
	MasterNode string
}
