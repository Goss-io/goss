package lib

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func FileHash(body []byte) string {
	h := md5.New()
	h.Write(body)
	b := h.Sum(nil)
	return hex.EncodeToString(b)
}

//Hash .
func Hash(body string) []byte {
	h := md5.New()
	h.Write([]byte(body))
	b := h.Sum(nil)
	return []byte(hex.EncodeToString(b))
}

//IsExists 判断ini是否存在.
func IsExists(ini string) bool {
	_, err := os.Stat(ini)
	if err != nil {
		return false
	}
	return true
}

//Time 获取当前时间.
func Time() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

//ParseInt .
func ParseInt(str string) int {
	num, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return num
}

func InArray(item string, list []string) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}

type Resp struct {
	Status bool
	Msg    interface{}
	Body   []byte
}

//Response .
func Response(w http.ResponseWriter, status bool, msg interface{}, body ...[]byte) {
	w.Header().Set("Content-Type", "application/json")
	dist := map[string]interface{}{
		"status": status,
		"msg":    msg,
	}
	if len(body) > 0 {
		dist["body"] = body[0]
	}
	b, err := json.Marshal(dist)
	if err != nil {
		log.Printf("err:%+v\n", err)
		return
	}

	w.Write(b)
}

//ParseMsg .
func ParseMsg(b []byte) Resp {
	r := Resp{}
	json.Unmarshal(b, &r)
	return r
}
