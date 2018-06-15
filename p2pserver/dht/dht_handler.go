package dht

import (
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/dht/types"
	mt "github.com/ontio/ontology/p2pserver/message/types"
	"net"
)

// findNodeHandle handles a find node message from UDP network
func (this *DHT) findNodeHandle(from *net.UDPAddr, msg mt.Message) {
	findNode, ok := msg.(*mt.FindNode)
	if !ok {
		log.Error("find node handle detected error message type!")
		return
	}

	if node := this.routingTable.queryNode(findNode.FromID); node == nil {
		return
	}

<<<<<<< HEAD
<<<<<<< HEAD
	this.updateNode(findNode.FromID)
	this.findNodeReply(from, findNode.TargetID)
=======
	this.updateNode(findNode.P.FromID)
	this.findNodeReply(from, findNode.P.TargetID)
>>>>>>> fix a bug of ping handler, ensure the routing table of a pair of nodes contains each other
=======
	this.updateNode(findNode.FromID)
	this.findNodeReply(from, findNode.TargetID)
>>>>>>> clean dht network message
}

// neighborsHandle handles a neighbors message from UDP network
func (this *DHT) neighborsHandle(from *net.UDPAddr, msg mt.Message) {
	neighbors, ok := msg.(*mt.Neighbors)
	if !ok {
		log.Error("neighbors handle detected error message type!")
		return
	}
<<<<<<< HEAD
<<<<<<< HEAD
=======

>>>>>>> clean dht network message
=======
>>>>>>> fix dht handle bug
	if node := this.routingTable.queryNode(neighbors.FromID); node == nil {
		return
	}

	requestId := types.ConstructRequestId(neighbors.FromID,
		types.DHT_FIND_NODE_REQUEST)
	this.messagePool.DeleteRequest(requestId)

	pingReqIds := make([]types.RequestId, 0)
<<<<<<< HEAD
<<<<<<< HEAD

	for i := 0; i < len(neighbors.Nodes); i++ {
		node := &neighbors.Nodes[i]
		if node.ID == this.nodeID {
			continue
		}
=======
	for i := 0; i < len(neighbors.P.Nodes); i++ {
		node := &neighbors.P.Nodes[i]
>>>>>>> optimize lookup
=======
	for i := 0; i < len(neighbors.Nodes); i++ {
		node := &neighbors.Nodes[i]
>>>>>>> clean dht network message
		// ping this node
		addr, err := getNodeUDPAddr(node)
		if err != nil {
			continue
<<<<<<< HEAD
=======

>>>>>>> optimize lookup
		}
		reqId, isNewRequest := this.messagePool.AddRequest(node, types.DHT_PING_REQUEST, nil, true)
		if isNewRequest {
			this.ping(addr)
		}
		pingReqIds = append(pingReqIds, reqId)
	}
	this.messagePool.Wait(pingReqIds)
	liveNodes := make([]*types.Node, 0)
<<<<<<< HEAD
<<<<<<< HEAD
	for i := 0; i < len(neighbors.Nodes); i++ {
		node := &neighbors.Nodes[i]
=======
	for i := 0; i < len(neighbors.P.Nodes); i++ {
		node := &neighbors.P.Nodes[i]
>>>>>>> optimize lookup
=======
	for i := 0; i < len(neighbors.Nodes); i++ {
		node := &neighbors.Nodes[i]
>>>>>>> clean dht network message
		if queryResult := this.routingTable.queryNode(node.ID); queryResult != nil {
			liveNodes = append(liveNodes, node)
		}
	}
	this.messagePool.SetResults(liveNodes)

<<<<<<< HEAD
<<<<<<< HEAD
	this.updateNode(neighbors.FromID)
=======
	this.updateNode(neighbors.P.FromID)
>>>>>>> fix a bug of ping handler, ensure the routing table of a pair of nodes contains each other
=======
	this.updateNode(neighbors.FromID)
>>>>>>> clean dht network message
}

// pingHandle handles a ping message from UDP network
func (this *DHT) pingHandle(from *net.UDPAddr, msg mt.Message) {
	ping, ok := msg.(*mt.DHTPing)
	if !ok {
		log.Error("ping handle detected error message type!")
		return
	}
<<<<<<< HEAD
<<<<<<< HEAD
=======

>>>>>>> clean dht network message
=======
>>>>>>> fix dht handle bug
	if ping.Version != this.version {
		log.Errorf("pingHandle: version is incompatible. local %d remote %d",
			this.version, ping.Version)
		return
	}

	// if routing table doesn't contain the node, add it to routing table and wait request return
<<<<<<< HEAD
<<<<<<< HEAD
	if node := this.routingTable.queryNode(ping.FromID); node == nil {
=======
	if node := this.routingTable.queryNode(ping.P.FromID); node == nil {
>>>>>>> fix a bug of ping handler, ensure the routing table of a pair of nodes contains each other
=======
	if node := this.routingTable.queryNode(ping.FromID); node == nil {
>>>>>>> clean dht network message
		node := &types.Node{
			ID:      ping.FromID,
			IP:      from.IP.String(),
			UDPPort: uint16(from.Port),
			TCPPort: uint16(ping.SrcEndPoint.TCPPort),
		}
<<<<<<< HEAD
<<<<<<< HEAD
		this.addNode(node)
	} else {
		// update this node
		bucketIndex, _ := this.routingTable.locateBucket(ping.FromID)
		this.routingTable.addNode(node, bucketIndex)
	}
	this.pong(from)
=======
		requestId := this.addNode(node, true)
		if len(requestId) > 0 {
			this.messagePool.Wait([]types.RequestId{requestId})
		}
	}
	// query again, if routing table contain the node, pong to from node
	if node := this.routingTable.queryNode(ping.P.FromID); node != nil {
		this.pong(from)
	}
>>>>>>> fix a bug of ping handler, ensure the routing table of a pair of nodes contains each other
=======
		this.addNode(node)
	} else {
		// update this node
		bucketIndex, _ := this.routingTable.locateBucket(ping.FromID)
		this.routingTable.addNode(node, bucketIndex)
	}
	this.pong(from)
>>>>>>> optimize lookup
	this.DisplayRoutingTable()
}

// pongHandle handles a pong message from UDP network
func (this *DHT) pongHandle(from *net.UDPAddr, msg mt.Message) {
	pong, ok := msg.(*mt.DHTPong)
	if !ok {
		log.Error("pong handle detected error message type!")
		return
	}
<<<<<<< HEAD
<<<<<<< HEAD
=======

>>>>>>> clean dht network message
=======
>>>>>>> fix dht handle bug
	if pong.Version != this.version {
		log.Errorf("pongHandle: version is incompatible. local %d remote %d",
			this.version, pong.Version)
		return
	}

	requesetId := types.ConstructRequestId(pong.FromID, types.DHT_PING_REQUEST)
	node, ok := this.messagePool.GetRequestData(requesetId)
	if !ok {
		// request pool doesn't contain the node, ping timeout
		this.routingTable.removeNode(pong.FromID)
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
