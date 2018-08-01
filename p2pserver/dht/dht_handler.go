package dht

import (
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/dht/types"
	mt "github.com/ontio/ontology/p2pserver/message/types"
	"net"
)

// findNodeHandle handles a find node message from UDP network
func (this *DHT) findNodeHandle(from *net.UDPAddr, packet []byte) {
	var findNode mt.FindNode
	if err := findNode.Deserialization(packet); err != nil {
		log.Error(err)
		return
	}

	if node := this.routingTable.queryNode(findNode.P.FromID); node == nil {
		return
	}

	this.updateNode(findNode.P.FromID)
	this.findNodeReply(from, findNode.P.TargetID)
}

// neighborsHandle handles a neighbors message from UDP network
func (this *DHT) neighborsHandle(from *net.UDPAddr, packet []byte) {
	var neighbors mt.Neighbors
	if err := neighbors.Deserialization(packet); err != nil {
		log.Error(err)
		return
	}

	if node := this.routingTable.queryNode(neighbors.P.FromID); node == nil {
		return
	}

	requestId := types.ConstructRequestId(neighbors.P.FromID,
		types.DHT_FIND_NODE_REQUEST)
	this.messagePool.DeleteRequest(requestId)

	pingReqIds := make([]types.RequestId, 0)
	for i := 0; i < len(neighbors.P.Nodes); i++ {
		node := &neighbors.P.Nodes[i]
		// ping this node
		addr, err := getNodeUDPAddr(node)
		if err != nil {
			continue

		}
		reqId, isNewRequest := this.messagePool.AddRequest(node, types.DHT_PING_REQUEST, nil, true)
		if isNewRequest {
			this.ping(addr)
		}
		pingReqIds = append(pingReqIds, reqId)
	}
	this.messagePool.Wait(pingReqIds)
	liveNodes := make([]*types.Node, 0)
	for i := 0; i < len(neighbors.P.Nodes); i++ {
		node := &neighbors.P.Nodes[i]
		if queryResult := this.routingTable.queryNode(node.ID); queryResult != nil {
			liveNodes = append(liveNodes, node)
		}
	}
	this.messagePool.SetResults(liveNodes)

	this.updateNode(neighbors.P.FromID)
}

// pingHandle handles a ping message from UDP network
func (this *DHT) pingHandle(from *net.UDPAddr, packet []byte) {
	var ping mt.DHTPing
	if err := ping.Deserialization(packet); err != nil {
		log.Error(err)
		return
	}

	if ping.P.Version != this.version {
		log.Errorf("pingHandle: version is incompatible. local %d remote %d",
			this.version, ping.P.Version)
		return
	}

	// if routing table doesn't contain the node, add it to routing table and wait request return
	if node := this.routingTable.queryNode(ping.P.FromID); node == nil {
		node := &types.Node{
			ID:      ping.P.FromID,
			IP:      from.IP.String(),
			UDPPort: uint16(from.Port),
			TCPPort: uint16(ping.P.SrcEndPoint.TCPPort),
		}
		this.addNode(node)
	} else {
		// update this node
		bucketIndex, _ := this.routingTable.locateBucket(ping.P.FromID)
		this.routingTable.addNode(node, bucketIndex)
	}
	this.pong(from)
	this.DisplayRoutingTable()
}

// pongHandle handles a pong message from UDP network
func (this *DHT) pongHandle(from *net.UDPAddr, packet []byte) {
	var pong mt.DHTPong
	if err := pong.Deserialization(packet); err != nil {
		log.Error(err)
		return
	}

	if pong.P.Version != this.version {
		log.Errorf("pongHandle: version is incompatible. local %d remote %d",
			this.version, pong.P.Version)
		return
	}

	requesetId := types.ConstructRequestId(pong.P.FromID, types.DHT_PING_REQUEST)
	node, ok := this.messagePool.GetRequestData(requesetId)
	if !ok {
		// request pool doesn't contain the node, ping timeout
		this.routingTable.removeNode(pong.P.FromID)
		return
	}

	// add to routing table
	this.addNode(node)
	// remove node from request pool
	this.messagePool.DeleteRequest(requesetId)
	log.Info("receive pong of ", requesetId)
}

// update the node to bucket when receive message from the node
func (this *DHT) updateNode(fromId types.NodeID) {
	node := this.routingTable.queryNode(fromId)
	if node != nil {
		// add node to bucket
		bucketIndex, _ := this.routingTable.locateBucket(fromId)
		this.routingTable.addNode(node, bucketIndex)
	}
}
