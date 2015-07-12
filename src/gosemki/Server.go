package main

import (
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"runtime"
)

const (
	CommandCloseDaemon = iota
)

type Server struct {
	Socket   string
	Listener net.Listener
	CmdInput chan int
}

func (this *Server) Exec(socket string) int {
	this.Socket = socket
	if FileExists(this.Socket) {
		fmt.Printf("unix socket: '%s' already exists\n", this.Socket)
		return 1
	}
	var err error
	this.Listener, err = net.Listen("unix", this.Socket)
	if err != nil {
		fmt.Printf("failed to start listen socket: '%s'\n", err.Error())
		return 1
	}
	defer os.Remove(this.Socket)
	err = rpc.Register(new(ServerRPC))
	if err != nil {
		fmt.Printf("failed to register RPC: '%s'\n", err.Error())
		return 1
	}
	this.CmdInput = make(chan int, 1)
	this.Loop()
	return 0
}

func (this *Server) Loop() {
	connInput := make(chan net.Conn, 2)
	go func() {
		for {
			conn, err := this.Listener.Accept()
			if err != nil {
				panic(errors.New("Daemon socket connection failure: " + err.Error()))
			}
			connInput <- conn
		}
	}()
	for {
		// handle connections or server CMDs (currently one CMD)
		select {
		case conn := <-connInput:
			rpc.ServeConn(conn)
			runtime.GC()
		case cmd := <-this.CmdInput:
			if cmd == CommandCloseDaemon {
				return
			}
		}
	}
}

func (this *Server) DropCache() {
	// Currently does nothing
}

func (this *Server) Reindex(file []byte, filePath string, packedContext GoBuildContext, result *IndexerResult) {
	defer func() {
		if err := recover(); err != nil {
			PrintBacktrace(err)
			result.InPanic = true
			this.DropCache()
		}
	}()
	indexer := new(PackageIndexer)
	indexer.context = UnpackGoBuildContext(&packedContext)
	indexer.result = result
	indexer.Reindex(filePath, file)
}

func (this *Server) Close() {
	this.CmdInput <- CommandCloseDaemon
}
