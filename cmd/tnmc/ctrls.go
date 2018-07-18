package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/mragiadakos/tendermoney/app/ctrls/utils"
	"github.com/urfave/cli"
)

var GenerateKeyCommand = cli.Command{
	Name:    "generate_inflator",
	Aliases: []string{"gi"},
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "filename",
			Usage: "the filename that the key will be saved",
		},
	},
	Usage: "generate the key in a file",
	Action: func(c *cli.Context) error {
		filename := c.String("filename")
		if len(filename) == 0 {
			return errors.New("Error: filename is missing")
		}

		kp, _ := utils.CreateKeyPair()
		kj := KeyPairJson{}
		pubB, _ := kp.Public.MarshalBinary()
		kj.PublicKey = hex.EncodeToString(pubB)
		privB, _ := kp.Private.MarshalBinary()
		kj.PrivateKey = hex.EncodeToString(privB)
		b, _ := json.Marshal(kj)
		err := ioutil.WriteFile(filename, b, 0644)
		if err != nil {
			return errors.New("Error: " + err.Error())
		}
		fmt.Println("The generate was successful")
		return nil
	},
}

var InflateCommand = cli.Command{
	Name:    "inflate",
	Aliases: []string{"i"},
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "key",
			Usage: "the filename of the inflator's key pair.",
		},
		cli.StringFlag{
			Name:  "vault",
			Usage: "the folder that will contain all the coins.",
		},
		cli.Float64Flag{
			Name:  "value",
			Usage: "the value of the new coin.",
		},
	},
	Usage: "Creates a new coin in the system and saves it in the vault's folder.",
	Action: func(c *cli.Context) error {
		key := c.String("key")
		if len(key) == 0 {
			return errors.New("Error: key is missing")
		}
		vault := c.String("vault")
		os.MkdirAll(vault, 0744)

		value := c.Float64("value")
		keyB, err := ioutil.ReadFile(key)
		if err != nil {
			return errors.New("Error: could not read the file that contains the key of the inflator, " + err.Error())
		}

		inflatorKpj := KeyPairJson{}
		err = json.Unmarshal(keyB, &inflatorKpj)
		if err != nil {
			return errors.New("Error: could not read the json format that contains the key of the inflator, " + err.Error())
		}
		filename, err := inflate(inflatorKpj, vault, value)
		if err != nil {
			return err
		}
		fmt.Println("The coin created successfully and saved in " + filename)
		return nil
	},
}

var SumCommand = cli.Command{
	Name:    "sum",
	Aliases: []string{"s"},
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "vault",
			Usage: "the folder that will contain all the coins.",
		},
		cli.StringFlag{
			Name:  "coins",
			Usage: "the list of coins seperated by comma.",
		},
	},
	Usage: "Sum two coins from the vault's folder and create a new one.",
	Action: func(c *cli.Context) error {
		vault := c.String("vault")
		if len(vault) == 0 {
			return errors.New("Error: vault is empty")
		}
		coinsListStr := c.String("coins")
		if len(coinsListStr) == 0 {
			return errors.New("Error: coins is empty")
		}

		coinsList := strings.Split(coinsListStr, ",")
		filename, err := sum(coinsList, vault)
		if err != nil {
			return err
		}
		fmt.Println("The coin created successfully and saved in " + filename)
		return nil
	},
}

var DivideCommand = cli.Command{
	Name:    "divide",
	Aliases: []string{"d"},
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "vault",
			Usage: "the folder that will contain all the coins.",
		},
		cli.StringFlag{
			Name:  "coin",
			Usage: "the coin that will be divided.",
		},
		cli.StringFlag{
			Name:  "values",
			Usage: "the values for the new coins that you expect.",
		},
	},
	Usage: "Divide a coin into smaller coins",
	Action: func(c *cli.Context) error {
		vault := c.String("vault")
		if len(vault) == 0 {
			return errors.New("Error: vault is empty")
		}

		coin := c.String("coin")
		if len(coin) == 0 {
			return errors.New("Error: coin is empty")
		}

		valuesStr := c.String("values")
		if len(valuesStr) == 0 {
			return errors.New("Error: values is empty")
		}

		valuesStrList := strings.Split(valuesStr, ",")
		values := []float64{}
		for _, vs := range valuesStrList {
			val, err := strconv.ParseFloat(vs, 64)
			if err != nil {
				return errors.New("Error: The value " + vs + " does not parse.")
			}
			fxVal := utils.ToFixed(val, 2)
			values = append(values, fxVal)
		}
		filenames, err := divide(vault, coin, values)
		if err != nil {
			return err
		}
		fmt.Println(len(filenames), " new coins have been created:")
		for _, v := range filenames {
			fmt.Println(v)
		}
		return nil
	},
}

var TaxCommand = cli.Command{
	Name: "tax",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "key",
			Usage: "the filename of the inflator's key pair.",
		},
		cli.IntFlag{
			Name:  "percent",
			Usage: "the percentage of the transactions.",
		},
	},
	Usage: "Create the tax for the transactions after it.",
	Action: func(c *cli.Context) error {
		key := c.String("key")
		if len(key) == 0 {
			return errors.New("Error: key is missing")
		}

		keyB, err := ioutil.ReadFile(key)
		if err != nil {
			return errors.New("Error: could not read the file that contains the key of the inflator, " + err.Error())
		}

		inflatorKpj := KeyPairJson{}
		err = json.Unmarshal(keyB, &inflatorKpj)
		if err != nil {
			return errors.New("Error: could not read the json format that contains the key of the inflator, " + err.Error())
		}

		err = tax(inflatorKpj, c.Int("percent"))
		if err != nil {
			return err
		}
		fmt.Println("The tax has been submitted.")
		return nil
	},
}

var GetLatestTaxCommand = cli.Command{
	Name:  "get_latest_tax",
	Usage: "Get latest tax.",
	Action: func(c *cli.Context) error {
		qmt, err := getLatestTax()
		if err != nil {
			return err
		}
		fmt.Println("Percent: ", qmt.Percentage)
		fmt.Println("Inflator: ", qmt.Inflator)
		return nil
	},
}

var SendCommand = cli.Command{
	Name:  "send",
	Usage: "Send the coins with the fee.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "vault",
			Usage: "the folder that will contain all the coins.",
		},
		cli.StringFlag{
			Name:  "coins",
			Usage: "the list of coins seperated by comma.",
		},
		cli.StringFlag{
			Name:  "fee",
			Usage: "the list of coins for the fee seperated by comma.",
		},
	},
	Action: func(c *cli.Context) error {
		vault := c.String("vault")
		if len(vault) == 0 {
			return errors.New("Error: vault is empty")
		}
		coinsListStr := c.String("coins")
		if len(coinsListStr) == 0 {
			return errors.New("Error: coins is empty")
		}

		feeListStr := c.String("fee")
		if len(feeListStr) == 0 {
			return errors.New("Error: coins for the fee is empty")
		}
		coinsList := strings.Split(coinsListStr, ",")
		feeList := strings.Split(feeListStr, ",")

		hash, secret, err := send(coinsList, feeList, vault)
		if err != nil {
			return err
		}
		fmt.Println("Hash: ", hash)
		fmt.Println("Secret: ", secret)
		return nil
	},
}

var GetTransactionCommand = cli.Command{
	Name:  "get_transaction",
	Usage: "Get transaction.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "hash",
			Usage: "the hash of the transaction.",
		},
	},
	Action: func(c *cli.Context) error {
		qmt, err := getTransaction(c.String("hash"))
		if err != nil {
			return err
		}
		fmt.Println("Hash: ", qmt.Hash)
		fmt.Println("Coins: ", qmt.Coins)
		fmt.Println("Fee: ", qmt.Fee)
		fmt.Println("The coins have been received: ", qmt.IsCoinsReceived)
		fmt.Println("The fee have been received: ", qmt.IsFeeReceived)
		return nil
	},
}

var GetCoin = cli.Command{
	Name:  "get_coin",
	Usage: "Get coin.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "coin",
			Usage: "the uuid of the coin.",
		},
	},
	Action: func(c *cli.Context) error {
		qmc, err := getCoin(c.String("coin"))
		if err != nil {
			return err
		}
		fmt.Println("Coin:", qmc.Coin)
		fmt.Println("Value:", qmc.Value)
		fmt.Println("Owner:", qmc.Owner)
		fmt.Println("Locked:", qmc.IsLocked)
		return nil
	},
}

var ReceiveCoinsCommand = cli.Command{
	Name:  "receive",
	Usage: "Receive the coins from the transaction.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "hash",
			Usage: "the hash of the transaction.",
		},
		cli.StringFlag{
			Name:  "secret",
			Usage: "the secret that will prove the new owners of the coins .",
		},
		cli.StringFlag{
			Name:  "vault",
			Usage: "the folder that will contain all the coins.",
		},
	},
	Action: func(c *cli.Context) error {
		vault := c.String("vault")
		if len(vault) == 0 {
			return errors.New("Error: vault is empty")
		}
		os.MkdirAll(vault, 0744)

		secret := c.String("secret")
		if len(secret) == 0 {
			return errors.New("Error: secret is empty")
		}

		hash := c.String("hash")
		if len(hash) == 0 {
			return errors.New("Error: hash is empty")
		}

		filenames, err := receive(vault, hash, secret)
		if err != nil {
			return err
		}
		fmt.Println(len(filenames), " new coins have been created:")
		for _, v := range filenames {
			fmt.Println(v)
		}
		return nil
	},
}

var GetTransactionsWithUnreceivedFeeCommand = cli.Command{
	Name:  "get_transactions_with_unreceived_fee",
	Usage: "Get transactions that their fee have not been yet received.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "hash",
			Usage: "the hash of the transaction.",
		},
	},
	Action: func(c *cli.Context) error {
		qmts, err := getTransactionsWithUnreceivedFee(c.String("hash"))
		if err != nil {
			return err
		}
		for _, qmt := range qmts {
			fmt.Println("Hash: ", qmt.Hash)
			fmt.Println("Coins: ", qmt.Coins)
			fmt.Println("Fee: ", qmt.Fee)
			fmt.Println("The coins have been received: ", qmt.IsCoinsReceived)
			fmt.Println("The fee have been received: ", qmt.IsFeeReceived)
			fmt.Println("")
		}
		return nil
	},
}

var ReceiveFeeCommand = cli.Command{
	Name:  "receive_fee",
	Usage: "Receive the fee from the transaction.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "hash",
			Usage: "the hash of the transaction.",
		},
		cli.StringFlag{
			Name:  "vault",
			Usage: "the folder that will contain all the coins.",
		},
		cli.StringFlag{
			Name:  "key",
			Usage: "the filename of the inflator's key pair.",
		},
	},
	Action: func(c *cli.Context) error {
		vault := c.String("vault")
		if len(vault) == 0 {
			return errors.New("Error: vault is empty")
		}
		os.MkdirAll(vault, 0744)

		hash := c.String("hash")
		if len(hash) == 0 {
			return errors.New("Error: hash is empty")
		}

		key := c.String("key")
		if len(key) == 0 {
			return errors.New("Error: key is missing")
		}

		keyB, err := ioutil.ReadFile(key)
		if err != nil {
			return errors.New("Error: could not read the file that contains the key of the inflator, " + err.Error())
		}

		inflatorKpj := KeyPairJson{}
		err = json.Unmarshal(keyB, &inflatorKpj)
		if err != nil {
			return errors.New("Error: could not read the json format that contains the key of the inflator, " + err.Error())
		}

		filenames, err := receiveFee(inflatorKpj, vault, hash)
		if err != nil {
			return err
		}
		fmt.Println(len(filenames), " new coins have been created:")
		for _, v := range filenames {
			fmt.Println(v)
		}
		return nil
	},
}
