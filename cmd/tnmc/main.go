package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Commands = []cli.Command{
		GenerateKeyCommand,
		InflateCommand,
		SumCommand,
		DivideCommand,
		TaxCommand,
		GetLatestTaxCommand,
		SendCommand,
		GetTransactionCommand,
		GetCoin,
		ReceiveCoinsCommand,
		GetTransactionsWithUnreceivedFeeCommand,
		ReceiveFeeCommand,
	}
	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
