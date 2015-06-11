package main

import (
    "net/rpc"
    "os"
    "fmt"
    "time"
)

type Client struct {
    Input string
    Command string
    Socket string
    RpcClient *rpc.Client
}

func (this *Client) Exec() int {
    var err error
    this.RpcClient, err = rpc.Dial("unix", this.Socket)
    if err != nil {
        if this.Command == "close" {
            fmt.Printf("Daemon not running, nothing to close\n")
            return 0
        }
        if FileExists(this.Socket) {
            os.Remove(this.Socket)
        }
        err = this.TryRunServer()
        if err != nil {
            fmt.Printf("%s\n", err.Error())
            return 1
        }
        err = this.TryConnectServer("unix", this.Socket)
        if err != nil {
            fmt.Printf("%s\n", err.Error())
            return 1
        }
    }
    if this.RpcClient != nil {
        defer this.RpcClient.Close()
    }
    switch this.Command {
    case "highlight":
        this.ExecHighlight()
    case "close":
        this.ExecClose()
    case "status":
        this.ExecStatus()
    default:
        fmt.Printf("Unknown command %s\n", this.Command)
        return 1
    }
    return 0
}

func (this *Client) TryRunServer() error {
    path := GetExecutableFilename()
    args := []string{os.Args[0], "-s"}
    cwd, _ := os.Getwd()
    stdin, err := os.Open(os.DevNull)
    if err != nil {
        return err
    }
    stdout, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
    if err != nil {
        return err
    }
    stderr, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
    if err != nil {
        return err
    }

    procattr := os.ProcAttr{Dir: cwd, Env: os.Environ(), Files: []*os.File{stdin, stdout, stderr}}
    process, err := os.StartProcess(path, args, &procattr)
    if err != nil {
        return err
    }

    return process.Release()
}

func (this *Client) TryConnectServer(network, address string) (err error) {
    t := 0
    for {
        this.RpcClient, err = rpc.Dial(network, address)
        if err != nil && t < 1000 {
            time.Sleep(10 * time.Millisecond)
            t += 10
            continue
        }
        break
    }
    return err
}

func (this *Client) ExecHighlight() {
}

func (this *Client) ExecClose() {
    ClientCloseServer(this.RpcClient)
}

func (this *Client) ExecStatus() {
    status := ClientStatus(this.RpcClient)
    fmt.Printf("Daemon status: '%s'\n", status)
}
