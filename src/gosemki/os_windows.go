// +build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"unsafe"
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
)

var (
	proc_get_module_file_name = kernel32.NewProc("GetModuleFileNameW")
)

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// Full path of the current executable
func GetExecutableFilename() string {
	b := make([]uint16, syscall.MAX_PATH)
	ret, _, err := syscall.Syscall(proc_get_module_file_name.Addr(), 3,
		0, uintptr(unsafe.Pointer(&b[0])), uintptr(len(b)))
	if int(ret) == 0 {
		panic(fmt.Sprintf("GetModuleFileNameW : err %d", int(err)))
	}
	return syscall.UTF16ToString(b)
}

//-------------------------------------------------------------------------
// print_backtrace
//
// a nicer backtrace printer than the default one
//-------------------------------------------------------------------------

var g_backtraceMutex sync.Mutex

func PrintBacktrace(err interface{}) {
	g_backtraceMutex.Lock()
	defer g_backtraceMutex.Unlock()
	fmt.Fprintf(os.Stderr, "panic: %v\n", err)
	i := 2
	for {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		f := runtime.FuncForPC(pc)
		fmt.Fprintf(os.Stderr, "%d.\t(%s): %s:%d\n", i-1, f.Name(), file, line)
		i++
	}
	fmt.Fprintln(os.Stderr, "")
}
