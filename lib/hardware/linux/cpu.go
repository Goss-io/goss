package linux

import "github.com/Goss-io/Goss/lib/command"

//CpuNum Cpu核数.
func CpuNum() string {
	num, err := command.Exec("cat /proc/cpuinfo | grep 'cpu cores' | head -1 | awk -F ':' '{print $2}'")
	if err != nil {
		num = "0"
	}

	num = num + "核"
	return num
}
