package main

import (
	"encoding/hex"
	"fmt"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/dht"
	"github.com/ontio/ontology/p2pserver/dht/types"
)

// test DHT
func main() {
	log.InitLog(1, log.PATH, log.Stdout)
	var acct *account.Account

	log.Info("0. Open the account")
	client, err := account.Open("./wallet.dat")
	if client == nil || err != nil{
		log.Fatal("Can't get local account.")
		return
	}
	acct, err = client.GetDefaultAccount([]byte("wangcheng"))
	if acct == nil || err != nil{
		log.Fatal("can not get default account")
		return
	}
	log.Debug("The Node's PublicKey ", acct.PublicKey)
	dhtConfig := types.GetDHTConfig()
	config.DefConfig.P2PNode.DHTSeeds = dhtConfig.DHTSeeds
	config.DefConfig.P2PNode.DHTUDPPort = dhtConfig.DHTUDPPort
	config.DefConfig.P2PNode.NodePort = dhtConfig.NodePort
	seeds := make([]*types.Node, 0, len(config.DefConfig.P2PNode.DHTSeeds))
	for i := 0; i < len(config.DefConfig.P2PNode.DHTSeeds); i++ {
		node := config.DefConfig.P2PNode.DHTSeeds[i]
		pubKey, err := hex.DecodeString(node.PubKey)
		k, err := keypair.DeserializePublicKey(pubKey)
		if err != nil {
			return
		}
		seed := &types.Node{
			IP:      node.IP,
			UDPPort: node.UDPPort,
			TCPPort: node.TCPPort,
		}
		seed.ID, _ = types.PubkeyID(k)
		seeds = append(seeds, seed)
	}

	// start seed node
	//seedIndex, _ := strconv.Atoi(os.Args[1])
	////seedIndex := 3
	//startSeedNode(seedIndex, seeds)

	//start common node
	nodeID, _ := types.PubkeyID(acct.PublicKey)
	//commonNode := &types.Node{
	//	ID:      nodeID,
	//	IP:      "127.0.0.1",
	//	UDPPort: 10010,
	//	TCPPort: 10011,
	//}
	testDht := dht.NewDHT(nodeID, seeds)
	testDht.Start()
	//stopCh := make(chan int)
	//<-stopCh
}

func startSeedNode(seedIndex int, seeds []*types.Node) {
	otherSeeds := make([]*types.Node, 3)
	seedNode := seeds[seedIndex]
	copy(otherSeeds[:seedIndex], seeds[:seedIndex])
	copy(otherSeeds[seedIndex:], seeds[seedIndex+1:])
	seedDht := dht.NewDHT(seedNode.ID, otherSeeds)
	go seedDht.Start()
	fmt.Println("node ", seedNode.ID, "start")
}
