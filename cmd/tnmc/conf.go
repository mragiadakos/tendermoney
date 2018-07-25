package main

import (
	"os"
)

type configurations struct {
	TendermintNode string
}

var Confs = configurations{}

func init() {
	node := os.Getenv("NODE")
	if len(node) > 0 {
		Confs.TendermintNode = node
	} else {
		Confs.TendermintNode = "tcp://localhost:26657"
	}
}
