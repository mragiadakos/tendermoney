package main

type configurations struct {
	TendermintNode string
}

var Confs = configurations{}

func init() {
	Confs.TendermintNode = "tcp://localhost:26657"
}
