package linux

import (
	"fmt"

	"github.com/Goss-io/goss/lib"
	"github.com/Goss-io/goss/lib/command"
)

//MemSize 内存大小.
func MemSize() string {
	num, err := command.Exec("cat /proc/meminfo | grep 'MemTotal' | awk -F ':' '{print $2}'")
	if err != nil {
		num = "0"
	}

	size := (lib.ParseInt(num) / (1000 * 1000))
	mem := fmt.Sprintf("%dG", size)

	return mem
}
