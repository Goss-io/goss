package dir

import (
	"fmt"
	"hash/crc32"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var storageList []string

//SwitchPath 选择存储路径.
func SwitchPath(hash string) string {
	n := crc32.ChecksumIEEE([]byte(hash))
	num := int64(n) % int64(len(storageList))
	return storageList[num]
}

func InitStoragePath(path string) error {
	//判断路径是否存在.
	if !isExists(path) {
		if err := os.MkdirAll(path, 0777); err != nil {
			return err
		}
	}
	if _, err := makeDir(path); err != nil {
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
			dirList, err := makeDir(path)
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

func makeDir(path string) (dirList []string, err error) {
	str := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	sarr := strings.Split(str, "")
	var num = 0
	for _, v := range sarr {
		for _, val := range sarr {
			if num > 255 {
				break
			}
			dirname := fmt.Sprintf("%s%s%s/", path, v, val)
			if isExists(dirname) {
				num++
				dirList = append(dirList, dirname)
				continue
			}
			if err := os.Mkdir(dirname, 0777); err != nil {
				log.Printf("%+v\n", err)
				return dirList, err
			}

			dirList = append(dirList, dirname)
			num++
		}
	}
	return dirList, nil
}

//isExists
func isExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true
}
