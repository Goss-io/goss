package ini

import (
	"bufio"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

var config map[string]string

//Load 加载.
func Load(name string) error {
	conf := map[string]string{}
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	br := bufio.NewReader(f)
	for {
		b, _, err := br.ReadLine()
		if err != nil && err == io.EOF {
			break
		}
		line := string(b)

		//替换掉当前行里面所有的空格.
		line = strings.Replace(line, " ", "", -1)

		//判断是否为注释.
		if strings.HasPrefix(line, "#") {
			continue
		}

		//去掉换行符.
		if len(line) < 1 {
			continue
		}

		key, value := parseLine(line)
		if err != nil {
			return err
		}
		conf[key] = value
	}
	config = conf
	return nil
}

//解析每一行的配置.
func parseLine(line string) (key string, value string) {
	key = line
	index := strings.Index(line, "=")
	if index > 0 {
		key = line[0:index]
		value = line[index+1 : len(line)]
	}
	return key, value
}

func GetString(key string) string {
	return config[key]
}

func GetInt(key string) int {
	if len(config[key]) < 1 {
		return 0
	}
	num, err := strconv.Atoi(config[key])
	if err != nil {
		log.Printf("获取int错误:%+v\n", err)
	}
	return num
}

func GetBool(key string) bool {
	if len(config[key]) < 1 {
		return false
	}
	b, err := strconv.ParseBool(config[key])
	if err != nil {
		log.Printf("获取bool错误%+v\n", err)
	}
	return b
}
