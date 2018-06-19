package confs

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"

	"github.com/ipfs/go-ipfs-api"

	uuid "github.com/satori/go.uuid"
)

type configuration struct {
	IpfsConnection string
	AbciDaemon     string
	Inflators      []string
}

var Conf = configuration{}

func init() {
	Conf.IpfsConnection = "127.0.0.1:5001"
	Conf.AbciDaemon = "tcp://0.0.0.0:46658"
	Conf.Inflators = []string{}
}

func (c *configuration) SetInflatorsFromHash(hash string) error {
	sh := shell.NewShell(c.IpfsConnection)
	file := uuid.NewV4().String()
	err := sh.Get(hash, file)
	if err != nil {
		return errors.New("Failed to get the json file for the inflators " + hash + ": " + err.Error())
	}
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return errors.New("Failed to read the json file from the hash " + hash + " for the inflators: " + err.Error())
	}
	os.RemoveAll(file)
	inflators := []string{}
	err = json.Unmarshal(b, &inflators)
	if err != nil {
		return errors.New("The json file for the inflators has not the correct JSON format: " + err.Error())
	}
	c.Inflators = inflators
	return nil
}
