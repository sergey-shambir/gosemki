package main

import (
    "net/rpc"
    "os"
    "fmt"
    "time"
    "go/build"
    "errors"
    "bytes"
    "bufio"
    "io"
    "path/filepath"
)

type Client struct {
    Input string
    Command string
    CommandArgs []string
    Socket string
    RpcClient *rpc.Client
}

func (this *Client) Exec() int {
    defer func() {
        if panicErr := recover(); panicErr != nil {
            PrintBacktrace(panicErr)
        }
    }()
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
        fmt.Fprintf(os.Stderr, "Unknown command %s\n", this.Command)
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
    context := PackGoBuildContext(&build.Default)
    content, path := this.PrepareFileTraits()
    ranges, errors := ClientHighlight(this.RpcClient, content, path, context)
    // FIXME: print (ranges, errors) returned by ClientHighlight
    fmt.Printf("ranges length = %d, errors length = %d\n", len(ranges), len(errors))
}

func (this *Client) ExecClose() {
    ClientCloseServer(this.RpcClient)
}

func (this *Client) ExecStatus() {
    status := ClientStatus(this.RpcClient)
    fmt.Printf("Daemon status: '%s'\n", status)
}

func (this *Client) PrepareFileTraits() ([]byte, string) {
    const BUFFER_SIZE = 64 * 1024
    var fileContent bytes.Buffer
    var filePath string
    if len(g_app.Input) > 0 {
        filePath = g_app.Input
        file, err := os.Open(g_app.Input)
        if err != nil {
            panic(err)
        }
        defer file.Close()
        buffer := make([]byte, BUFFER_SIZE)
        for  {
            readCount, err := file.Read(buffer)
            if err != nil && err != io.EOF {
                panic(err)
            }
            if readCount == 0 {
                break
            }
            fileContent.Write(buffer[:readCount])
        }
    } else {
        if (len(this.CommandArgs) == 0) {
            panic(errors.New("missed <path> parameter or -in=<path> option"))
        }
        filePath = this.CommandArgs[0]
        reader := bufio.NewReader(os.Stdin)
        for {
            line, _, err := reader.ReadLine()
            if err != nil && err != io.EOF {
                panic(err)
            }
            fileContent.Write(line)
            if err == io.EOF {
                break
            }
        }
    }
    if len(filePath) != 0 && !filepath.IsAbs(filePath) {
        cwd, _ := os.Getwd()
        filePath = filepath.Join(cwd, filePath)
    }
    return fileContent.Bytes(), filePath
}
