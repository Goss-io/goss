package hardware

import (
	"runtime"

	"github.com/Goss-io/Goss/lib/hardware/linux"
)

type cpu struct {
	Num string
}
type mem struct {
	Total string
}

//Hardware .
type Hardware struct {
	Cpu cpu
	Mem mem
}

func New() Hardware {
	h := Hardware{
		Cpu: cpu{
			Num: "8核",
		},
		Mem: mem{
			Total: "16G",
		},
	}

	system := checkSystem()
	if system == "linux" {
		h.Cpu.Num = linux.CpuNum()
		h.Mem.Total = linux.MemSize()
	}

	return h
}

func checkSystem() string {
	system := runtime.GOOS
	return system
}
