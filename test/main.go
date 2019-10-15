package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func main() {
	st := time.Now().Unix()
	for i := 1; i <= 1; i++ {
		upload()
	}

	et := time.Now().Unix()

	log.Println("共耗时:", et-st, "秒")
}

func upload() {
	filename := "./timg.jpeg"
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("%+v\n", err)
		return
	}

	client := http.Client{}
	fname := fmt.Sprintf("http://127.0.0.1/oss/%s", "5.jpeg")
	req, err := http.NewRequest("PUT", fname, bytes.NewBuffer(b))
	if err != nil {
		log.Printf("%+v\n", err)
		return
	}
	req.Header.Add("AccessKey", "")
	req.Header.Add("SecretKey", "")
	_, err = client.Do(req)
	if err != nil {
		log.Printf("%+v\n", err)
		return
	}

	log.Println(fname)
}
