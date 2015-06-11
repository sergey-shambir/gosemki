package main

import (
    "net/rpc"
)

// FIXME: move to other file
type GoRange struct {
}
type GoError struct {
    line int
    column int
    lenght int
    message string
}

type ServerRPC struct {
}

// RPC for higlight

type ArgsHighlight struct {
    Content []byte
    Path string
    Cursor int
    Context GoBuildContext
}
type ReplyHighlight struct {
    ranges []GoRange
    errors []GoError
}
func (r *ServerRPC) Highlight(args *ArgsHighlight, reply *ReplyHighlight) error {
    // FIXME: do job
    return nil
}
func ClientHighlight(client *rpc.Client, content []byte, path string, cursor int, context GoBuildContext) (ranges []GoRange, errors []GoError) {
    var args ArgsHighlight
    var reply ReplyHighlight
    args.Content = content
    args.Path = path
    args.Cursor = cursor
    args.Context = context
    err := client.Call("ServerRPC.Highlight", &args, &reply)
    if err != nil {
        panic(err)
    }
    return reply.ranges, reply.errors
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
