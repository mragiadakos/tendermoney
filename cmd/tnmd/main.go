package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	kitlog "github.com/go-kit/kit/log"
	"github.com/mragiadakos/tendermoney/app/confs"
	"github.com/mragiadakos/tendermoney/app/ctrls"
	absrv "github.com/tendermint/abci/server"
	cmn "github.com/tendermint/tmlibs/common"
	tmlog "github.com/tendermint/tmlibs/log"
)

func main() {
	logger := tmlog.NewTMLogger(kitlog.NewSyncWriter(os.Stdout))
	flagAbci := "socket"
	ipfsDaemon := flag.String("ipfs", "127.0.0.1:5001", "the URL for the IPFS's daemon")
	node := flag.String("node", "tcp://0.0.0.0:26658", "the TCP URL for the ABCI daemon")
	inflatorsHash := flag.String("inflators-hash", "", "the IPFS hash with the json for the inflators")
	inflatorsFile := flag.String("inflators-file", "", "the file with json array of public keys")
	flag.Parse()

	if len(*inflatorsHash) > 0 {

		err := confs.Conf.SetInflatorsFromHash(*inflatorsHash)
		if err != nil {
			fmt.Println("Error:", err.Error())
			return
		}
	} else {
		if len(*inflatorsFile) == 0 {
			fmt.Println("Error: need the list of inflators to continue.")
			return
		}
		b, err := ioutil.ReadFile(*inflatorsFile)
		if err != nil {
			fmt.Println("Error: the file " + *inflatorsFile + " has problem:" + err.Error())
			return
		}
		arr := []string{}
		err = json.Unmarshal(b, &arr)
		if err != nil {
			fmt.Println("Error: the file " + *inflatorsFile + " has problem with encoding:" + err.Error())
			return
		}
		confs.Conf.Inflators = arr
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
