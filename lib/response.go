package lib

import (
	"encoding/json"
	"log"
	"net/http"
)

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
