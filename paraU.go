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
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
