package dht

import (
	"fmt"
	"github.com/ontio/ontology/p2pserver/dht/types"
	mt "github.com/ontio/ontology/p2pserver/message/types"
	"net"
)

func (this *DHT) FindNodeHandler(from *net.UDPAddr, findNodeMsgData []byte) error {
	findNodeMsg := new(mt.FindNode)
	findNodeMsg.Deserialization(findNodeMsgData)
	this.updateFromNode(findNodeMsg.P.FromID)
	return this.FindNodeReply(from, findNodeMsg.P.TargetID)
}

func (this *DHT) NeighborsHandler(from *net.UDPAddr, neighborsMsgData []byte) {
	neighborsMsg := &mt.Neighbors{}
	neighborsMsg.Deserialization(neighborsMsgData)
	neighbors := neighborsMsg.P.Nodes
	results := make([]*types.Node, 0, len(neighbors))
	for i := 0; i < len(neighbors); i++ {
		results = append(results, &neighbors[i])
	}
	this.findNodeQueue.SetResult(results, neighborsMsg.P.FromID)
	this.updateFromNode(neighborsMsg.P.FromID)
}

func (this *DHT) PingHandler(fromAddr *net.UDPAddr, pingMsgData []byte) {
	pingMsg := new(mt.DHTPing)
	pingMsg.Deserialization(pingMsgData)
	// response
	this.Pong(fromAddr)
	this.updateFromNode(pingMsg.P.FromID)
	fromNode := this.routingTable.queryNode(pingMsg.P.FromID)
	if fromNode == nil {
		node := &types.Node{
			ID:      pingMsg.P.FromID,
			IP:      fromAddr.IP.String(),
			UDPPort: uint16(fromAddr.Port),
		}
		this.AddNode(node)
	}
}

func (this *DHT) PongHandler(fromAddr *net.UDPAddr, pongMsgData []byte) {
	pongMsg := new(mt.DHTPong)
	pongMsg.Deserialization(pongMsgData)
	fromNodeId := pongMsg.P.FromID
	fromNode, ok := this.pingNodeQueue.GetRequestNode(fromNodeId)
	if !ok {
		// ping node queue doesn't contain the node, ping timeout
		fmt.Println("PongHandler: from id ", fromNodeId)
		this.routingTable.RemoveNode(fromNodeId)
	} else {
		// add to routing table
		this.AddNode(fromNode)
		// remove node from ping node queue
		this.pingNodeQueue.DeleteNode(fromNodeId)
		this.pingNodeQueue.AppendRsp(fromNode)
	}
}

// update the node to bucket when receive message from the node
func (this *DHT) updateFromNode(fromNodeId types.NodeID) {
	fromNode := this.routingTable.queryNode(fromNodeId)
	if fromNode != nil {
		// add node to bucket
		this.AddNode(fromNode)
	}
}
