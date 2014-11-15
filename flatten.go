package main

import (
	"os"
	"runtime"

	"github.com/FZambia/flatten/cmd"
	"github.com/codegangsta/cli"
)

const APP_VER = "0.1"

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	app := cli.NewApp()
	app.Name = "Flatten"
	app.Usage = "URL content flatten tool"
	app.Version = APP_VER
	app.Commands = []cli.Command{
		cmd.CmdWeb,
	}
	app.Flags = append(app.Flags, []cli.Flag{}...)
	app.Run(os.Args)
}
