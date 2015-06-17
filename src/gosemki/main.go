package main

import (
    "os"
    "flag"
    "fmt"
    "path/filepath"
)

func ShowApplicationUsage() {
    fmt.Fprintf(os.Stderr,
            "Usage: %s [-s] [-in=<path>]\n"+
                    "       <command> [<args>]\n\n",
            os.Args[0])
    fmt.Fprintf(os.Stderr,
            "Flags:\n")
    flag.PrintDefaults()
    fmt.Fprintf(os.Stderr,
            "\nCommands:\n"+
            "  highlight <offset> [<path>]        highlight command\n" +
            "  close                              close the gocode daemon\n" +
            "  status                             gocode daemon status report\n")
}

type Application struct {
    IsServer bool
    Input string
    Server *Server
}

var g_app *Application

func (this *Application) Init() {
    flag.BoolVar(&this.IsServer, "s", false, "run a server instead of a client")
    flag.StringVar(&this.Input, "in", "", "use this file instead of stdin input")
    flag.Usage = ShowApplicationUsage
    flag.Parse()
}

func (this *Application) Exec() {
    var code int
    if this.IsServer {
        code = this.ExecServer()
    } else {
        code = this.ExecClient()
    }
    os.Exit(code)
}

func (this *Application) ExecServer() int {
    this.Server = new(Server)
    return this.Server.Exec(this.GetSocketFilename())
}

func (this *Application) ExecClient() int {
    if flag.NArg() > 0 {
        client := new(Client)
        client.Input = this.Input
        client.Command = flag.Arg(0)
        client.CommandArgs = flag.Args()[1:]
        client.Socket = this.GetSocketFilename() //this.GetSocketFilename()
        return client.Exec()
    }
    ShowApplicationUsage()
    return 0
}

func (_ *Application) GetSocketFilename() string {
    user := os.Getenv("USER")
    if len(user) == 0 {
        user = "all"
    }
    return filepath.Join(os.TempDir(), fmt.Sprintf("gosemki-daemon.%s", user))
}

func main() {
    g_app = new(Application)
    g_app.Init()
    g_app.Exec()
}
