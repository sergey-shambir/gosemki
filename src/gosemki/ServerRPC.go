package main

import (
    "net/rpc"
)

type ServerRPC struct {
}

// RPC for higlight

type ArgsHighlight struct {
    Content []byte
    Path string
    Context GoBuildContext
}
type ReplyHighlight struct {
    Ranges []GoRange
    Errors []GoError
}
func (r *ServerRPC) Highlight(args *ArgsHighlight, reply *ReplyHighlight) error {
    reply.Ranges, reply.Errors = g_app.Server.Highlight(args.Content, args.Path, args.Context)
    return nil
}
func ClientHighlight(client *rpc.Client, content []byte, path string, context GoBuildContext) (ranges []GoRange, errors []GoError) {
    args := &ArgsHighlight{content, path, context}
    var reply ReplyHighlight
    err := client.Call("ServerRPC.Highlight", args, &reply)
    if err != nil {
        panic(err)
    }
    return reply.Ranges, reply.Errors
}

// RPC for close server
type ArgsCloseServer struct {
    Unused int
}
type ReplyCloseServer struct {
    Unused int
}
func (r *ServerRPC) CloseServer(args *ArgsCloseServer, reply *ReplyCloseServer) error {
    g_app.Server.Close()
    reply.Unused = 0
    return nil
}
func ClientCloseServer(client *rpc.Client) {
    args := &ArgsCloseServer{0}
    var reply ReplyCloseServer
    err := client.Call("ServerRPC.CloseServer", args, &reply)
    if err != nil {
        panic(err)
    }
}

// RPC for status
type ArgsStatus struct {
    Unused int
}
type ReplyStatus struct {
    Status string
}
func (r *ServerRPC) GetStatus(args *ArgsStatus, reply *ReplyStatus) error {
    reply.Status = "daemon running OK"
    return nil
}
func ClientStatus(client *rpc.Client) string {
    args := &ArgsStatus{0}
    var reply ReplyStatus
    args.Unused = 0
    err := client.Call("ServerRPC.GetStatus", args, &reply)
    if err != nil {
        panic(err)
    }
    return reply.Status
}
