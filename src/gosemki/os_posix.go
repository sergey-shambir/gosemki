// +build !windows

package main

import (
        "os"
        "os/exec"
        "path/filepath"
        "runtime"
        "sync"
        "fmt"
)

func FileExists(filename string) bool {
    _, err := os.Stat(filename)
    return err == nil
}

// Full path of the current executable
func GetExecutableFilename() string {
    // try readlink first
    path, err := os.Readlink("/proc/self/exe")
    if err == nil {
            return path
    }
    // use argv[0]
    path = os.Args[0]
    if !filepath.IsAbs(path) {
            cwd, _ := os.Getwd()
            path = filepath.Join(cwd, path)
    }
    if FileExists(path) {
            return path
    }
    // Fallback : use "gocode" and assume we are in the PATH...
    path, err = exec.LookPath("gocode")
    if err == nil {
            return path
    }
    return ""
}

// config location

func GetXdgConfigDir() string {
    xdghome := os.Getenv("XDG_CONFIG_HOME")
    if xdghome == "" {
        xdghome = filepath.Join(os.Getenv("HOME"), ".config")
    }
    return xdghome
}

func GetGocodeConfigDir() string {
    return filepath.Join(GetXdgConfigDir(), "gocode")
}

func GetGocodeConfigFile() string {
    return filepath.Join(GetXdgConfigDir(), "gocode", "config.json")
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
