package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ontio/ontology/common/config"
	"io/ioutil"
	"os"
)

type DHTConfig struct {
	DHTSeeds   []config.DHTNode
	DHTUDPPort uint `json:"DHTUDPPort"`
	NodePort   uint `json:"NodePort"`
}

func GetDHTConfig() *DHTConfig {
	dhtConfig := DHTConfig{}
	file, e := ioutil.ReadFile("./config.json")
	if e != nil {
		fmt.Println("read dht config failed!")
		os.Exit(1)
	}

	// Remove the UTF-8 Byte Order Mark
	file = bytes.TrimPrefix(file, []byte("\xef\xbb\xbf"))
	e = json.Unmarshal(file, &dhtConfig)
	if e != nil {
		fmt.Println("unmarshal dht config failed!")
		os.Exit(1)
	}
	return &dhtConfig
}
