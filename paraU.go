package main

import (
	"github.com/ipfs/go-log/v2"
	"github.com/vvarma/gotalk/cmd"
)

func main() {
	log.SetAllLoggers(log.LevelWarn)
	log.SetLogLevel("commands", "debug")
	log.SetLogLevel("gotalk", "info")
	log.SetLogLevel("paraU", "debug")
	log.SetLogLevel("subcmd", "debug")
	log.SetLogLevel("chat", "debug")
	log.SetLogLevel("client", "debug")
	log.SetLogLevel("control", "debug")
	log.SetLogLevel("dost", "debug")

	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
