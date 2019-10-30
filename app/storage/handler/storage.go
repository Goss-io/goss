package handler

import (
	"fmt"
	"hash/crc32"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/Goss-io/goss/app/storage/conf"
	"github.com/Goss-io/goss/lib"
)

var storageList []string

//SelectPath 选择存储路径.
func (s *StorageService) SelectPath(hash string) string {
	n := crc32.ChecksumIEEE([]byte(hash))
	num := int64(n) % int64(len(storageList))
	return storageList[num]
}

//InitStoragePath 初始化存储目录.
func (s *StorageService) InitStoragePath(path string) error {
	//判断路径是否存在.
	if !lib.IsExists(path) {
		if err := os.MkdirAll(path, 0777); err != nil {
			return err
		}
	}
	if _, err := s.makeDir(path); err != nil {
		log.Printf("%+v\n", err)
		return err
	}
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Printf("%+v\n", err)
		return err
	}

	list := []string{}
	for _, f := range files {
		if f.IsDir() {
			//创建文件夹.
			path := fmt.Sprintf("%s%s/", path, f.Name())
			dirList, err := s.makeDir(path)
			if err != nil {
				log.Printf("%+v\n", err)
				return err
			}
			list = append(list, dirList...)
		}
	}

	storageList = list
	return nil
}

//makeDir 创建存储目录.
func (s *StorageService) makeDir(path string) (dirList []string, err error) {
	str := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	sarr := strings.Split(str, "")
	var num = 0
	for _, v := range sarr {
		for _, val := range sarr {
			if num > 255 {
				break
			}
			num++
			dirname := fmt.Sprintf("%s%s%s/", path, v, val)
			if lib.IsExists(dirname) {
				dirname = strings.Replace(dirname, conf.Conf.Node.StorageRoot, "", -1)
				dirList = append(dirList, dirname)
				continue
			}
			if err := os.Mkdir(dirname, 0777); err != nil {
				log.Printf("%+v\n", err)
				return dirList, err
			}

			dirname = strings.Replace(dirname, conf.Conf.Node.StorageRoot, "", -1)
			dirList = append(dirList, dirname)
		}
	}
	return dirList, nil
}
