package handler

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/Goss-io/goss/app/storage/conf"
	"github.com/Goss-io/goss/lib"
	"github.com/Goss-io/goss/lib/logd"
)

//httpSrv .
func (s *StorageService) httpSrv() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//检查token.
		if !s.checkAuth(r.Header.Get("token")) {
			//todo 未授权的访问.
			return
		}
		if r.Method == "PUT" {
			//文件内容.
			fbody, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Printf("%+v\n", err)
				return
			}
			log.Printf("r.Body:%+v\n", string(fbody))

			fhash := r.Header.Get("fhash")
			fPath, err := s.put(fhash, fbody)
			if err != nil {
				lib.Response(w, false, err.Error())
				return
			}

			lib.Response(w, true, fPath)
			return
		}

		if r.Method == "GET" {
			fpath := r.Header.Get("fpath")
			b, err := s.get(fpath)
			if err != nil {
				log.Printf("err:%+v\n", err)
				lib.Response(w, false, err.Error())
				return
			}
			lib.Response(w, true, "获取成功", b)
			return
		}

		if r.Method == "DELETE" {
			// s.delete(fname)
			return
		}

		lib.Response(w, false, "禁止访问")
	})
	http.ListenAndServe(s.Port, nil)
}

//checkAuth .
func (s *StorageService) checkAuth(token string) bool {
	if token != conf.Conf.Node.Token {
		return false
	}
	return true
}

//put 记录文件.
func (s *StorageService) put(hash string, fbody []byte) (fPath string, err error) {
	//计算文件hash.
	fHash := lib.FileHash(fbody)
	//验证文件是否损坏.
	if fHash != hash {
		logd.Make(logd.Level_WARNING, logd.GetLogpath(), "文件hash不一致")
		return fPath, errors.New("文件hash不一致")
	}

	fPath = s.SelectPath(fHash) + fHash
	storePath := conf.Conf.Node.StorageRoot + fPath
	log.Println("storePath:", storePath)
	err = ioutil.WriteFile(storePath, fbody, 0777)
	if err != nil {
		logd.Make(logd.Level_WARNING, logd.GetLogpath(), "创建文件失败"+err.Error())
		return fPath, err
	}

	return fPath, nil
}

//获取文件.
func (s *StorageService) get(fpath string) (fbody []byte, err error) {
	fpath = conf.Conf.Node.StorageRoot + fpath
	fbody, err = ioutil.ReadFile(fpath)
	if err != nil {
		log.Printf("err:%+v\n", err)
		logd.Make(logd.Level_WARNING, logd.GetLogpath(), "读取文件失败:"+err.Error())
		return fbody, err
	}

	logd.Make(logd.Level_INFO, logd.GetLogpath(), "文件发送成功")
	return fbody, nil
}
