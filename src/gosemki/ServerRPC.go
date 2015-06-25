package main

import (
    "net/rpc"
)

type ServerRPC struct {
}

// RPC for higlight

type ArgsReindex struct {
    Content []byte
    Path string
    Context GoBuildContext
}

func (r *ServerRPC) Reindex(args *ArgsReindex, result *IndexerResult) error {
    g_app.Server.Reindex(args.Content, args.Path, args.Context, result)
    return nil
}

func ClientReindex(client *rpc.Client, content []byte, path string, context GoBuildContext) IndexerResult {
    args := &ArgsReindex{content, path, context}
    var result IndexerResult
    err := client.Call("ServerRPC.Reindex", args, &result)
    if err != nil {
        panic(err)
    }
    return result
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
