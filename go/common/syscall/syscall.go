package syscall

import (
	"syscall"
)

// 提升程序可打开的文件描述符上限
func init() {
	var rlim syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlim); err != nil {
		panic(err.Error())
	}
	rlim.Cur = 1000000
	rlim.Max = 1000000
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rlim); err != nil {
		panic(err.Error())
	}
}
