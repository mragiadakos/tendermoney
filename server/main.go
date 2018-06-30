package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	kitlog "github.com/go-kit/kit/log"
	"github.com/mragiadakos/tendermoney/server/confs"
	"github.com/mragiadakos/tendermoney/server/ctrls"
	absrv "github.com/tendermint/abci/server"
	cmn "github.com/tendermint/tmlibs/common"
	tmlog "github.com/tendermint/tmlibs/log"
)

func main() {
	logger := tmlog.NewTMLogger(kitlog.NewSyncWriter(os.Stdout))
	flagAbci := "socket"
	ipfsDaemon := flag.String("ipfs", "127.0.0.1:5001", "the URL for the IPFS's daemon")
	node := flag.String("node", "tcp://0.0.0.0:26658", "the TCP URL for the ABCI daemon")
	inflatorsHash := flag.String("inflators", "", "the IPFS hash with the json for the inflators")
	flag.Parse()

	if len(*inflatorsHash) == 0 {
		fmt.Println("Error:", errors.New("The IPFS hash with the json for the inflators is missing"))
		return
	}

	err := confs.Conf.SetInflatorsFromHash(*inflatorsHash)
	if err != nil {
		fmt.Println("Error:", err.Error())
		return
	}
	confs.Conf.AbciDaemon = *node
	confs.Conf.IpfsConnection = *ipfsDaemon

	app := ctrls.NewTMApplication()
	srv, err := absrv.NewServer(confs.Conf.AbciDaemon, flagAbci, app)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	srv.SetLogger(logger.With("module", "abci-server"))
	if err := srv.Start(); err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		srv.Stop()
	})

}
