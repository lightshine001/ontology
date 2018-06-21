package dht

import (
	"bytes"
	"encoding/hex"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/dht/types"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func getTestSeeds() []*types.Node {
	seeds := make([]*types.Node, 0, len(config.DefConfig.Genesis.DHT.Seeds))
	for i := 0; i < len(config.DefConfig.Genesis.DHT.Seeds); i++ {
		node := config.DefConfig.Genesis.DHT.Seeds[i]
		pubKey, err := hex.DecodeString(node.PubKey)
		k, err := keypair.DeserializePublicKey(pubKey)
		if err != nil {
			return nil
		}
		seed := &types.Node{
			IP:      node.IP,
			UDPPort: node.UDPPort,
			TCPPort: node.TCPPort,
		}
		seed.ID, _ = types.PubkeyID(k)
		seeds = append(seeds, seed)
	}
	return seeds
}

func getTestRoutingTable() *routingTable {
	seeds := getTestSeeds()
	table := new(routingTable)
	table.init(seeds[3].ID, nil)
	for i := 0; i < 2; i++ {
		bucketIndex, _ := table.locateBucket(seeds[i].ID)
		table.addNode(seeds[i], bucketIndex)
	}
	return table
}

func TestRoutingTableSerialize(t *testing.T) {
	log.InitLog(log.InfoLog, log.PATH, log.Stdout)
	defer os.RemoveAll("./Log")

	bf := new(bytes.Buffer)
	table := getTestRoutingTable()
	err := table.serialize(bf)
	assert.Nil(t, err)

	deserializeTable := new(routingTable)
	var emptyId types.NodeID
	deserializeTable.init(emptyId, nil)
	err = deserializeTable.deserialize(bf)
	assert.Nil(t, err)

	assert.Equal(t, table, deserializeTable)
}

func TestDHTLoadAndSave(t *testing.T) {
	log.InitLog(log.InfoLog, log.PATH, log.Stdout)
	defer func() {
		os.RemoveAll("./Log")
		os.Remove(DHT_ROUTING_TABLE)
	}()
	testTable := getTestRoutingTable()
	var emptyId types.NodeID
	dht := NewDHT(emptyId, getTestSeeds())
	dht.routingTable = testTable
	dht.saveToFile()
	// cathe file exist
	_, err := os.Stat(DHT_ROUTING_TABLE)
	assert.Nil(t, err)
	loadDht := NewDHT(emptyId, getTestSeeds())
	loadDht.routingTable.feedCh = nil
	loadDht.loadFromFile()
	assert.Equal(t, dht.routingTable, loadDht.routingTable)
}
