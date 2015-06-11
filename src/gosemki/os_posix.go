// +build !windows

package main

import (
        "os"
        "os/exec"
        "path/filepath"
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
