package filetype

import "strings"

var tyeList = map[string]string{
	"47494638": "image/gif",
	"ffd8ffe":  "image/jpeg",
	"89504e47": "image/png",
	"4740111":  "video/mp2t",
	"23455854": "audio/mpegurl",
}

func Parse(str string) string {
	for k, v := range tyeList {
		if strings.Contains(str, k) {
			return v
		}
	}
	return "other"
}
