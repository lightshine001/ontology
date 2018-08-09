/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package msgpack

import (
	"bytes"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	ct "github.com/ontio/ontology/core/types"
	msgCommon "github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/dht/types"
	mt "github.com/ontio/ontology/p2pserver/message/types"
	p2pnet "github.com/ontio/ontology/p2pserver/net/protocol"

	"github.com/ontio/ontology/p2pserver/message/pb"
	"net"
	"time"
)

///block package
func ConstructBlockMsg(bk *ct.Block) *mt.NetMessage {
	log.Debug()
	p := bytes.NewBuffer(nil)
	err := bk.Serialize(p)
	if err != nil {
		return nil
	}
	blkMsg := &netpb.Block{
		BlockData: p.Bytes(),
	}
	return mt.NewNetMessage(msgCommon.BLOCK_TYPE, blkMsg)

}

//blk hdr package
func ConstructHeadersMsg(headers []*ct.Header) *mt.NetMessage {
	headersMsg := &netpb.BlkHeader{
		BlkHdr: make([][]byte, 0, len(headers)),
	}
	for _, header := range headers {
		p := bytes.NewBuffer(nil)
		err := header.Serialize(p)
		if err != nil {
			return nil
		}
		headersMsg.BlkHdr = append(headersMsg.BlkHdr, p.Bytes())
	}
	return mt.NewNetMessage(msgCommon.HEADERS_TYPE, headersMsg)
}

//blk hdr req package
func ConstructHeadersReqMsg(curHdrHash common.Uint256) *mt.NetMessage {
	headersReq := &netpb.HeadersReq{
		Len:     1,
		HashEnd: curHdrHash.ToArray(),
	}

	return mt.NewNetMessage(msgCommon.GET_HEADERS_TYPE, headersReq)
}

////Consensus info package
func ConsrtuctConsensusMsg(cp *netpb.ConsensusPayload) *mt.NetMessage {
	log.Debug()
	cons := &netpb.Consensus{
		Cons: cp,
		Hop:  msgCommon.MAX_HOP,
	}

	return mt.NewNetMessage(msgCommon.CONSENSUS_TYPE, cons)
}

//InvPayload
func ConstructInvPayload(invType common.InventoryType, msg []common.Uint256) *netpb.InvPayload {
	invPayload := &netpb.InvPayload{}
	invPayload.InvType = netpb.InventoryType(invType)
	invPayload.Blk = make([][]byte, 0, len(msg))
	for _, v := range msg {
		invPayload.Blk = append(invPayload.Blk, v.ToArray())
	}

	return invPayload
}

//Inv request package
func ConstructInv(invPayload *netpb.InvPayload) *mt.NetMessage {
	inv := &netpb.Inv{
		P:   invPayload,
		Hop: msgCommon.MAX_HOP,
	}
	return mt.NewNetMessage(msgCommon.INV_TYPE, inv)
}

//NotFound package
func ConstructNotFound(hash common.Uint256) *mt.NetMessage {
	log.Debug()
	notFound := &netpb.NotFound{
		Hash: hash.ToArray(),
	}

	return mt.NewNetMessage(msgCommon.NOT_FOUND_TYPE, notFound)
}

//ping msg package
func ConstructPingMsg(height uint64) *mt.NetMessage {
	log.Debug()
	ping := &netpb.Ping{
		Height: height,
	}

	return mt.NewNetMessage(msgCommon.PING_TYPE, ping)
}

//pong msg package
func ConstructPongMsg(height uint64) *mt.NetMessage {
	log.Debug()
	pong := &netpb.Pong{
		Height: height,
	}

	return mt.NewNetMessage(msgCommon.PONG_TYPE, pong)
}

//Transaction package
func ConstructTxn(txn *ct.Transaction) *mt.NetMessage {
	log.Debug()
	p := bytes.NewBuffer(nil)
	err := txn.Serialize(p)
	if err != nil {
		return nil
	}
	txnMsg := &netpb.Trn{
		Transaction: p.Bytes(),
		Hop:         msgCommon.MAX_HOP,
	}

	return mt.NewNetMessage(msgCommon.TX_TYPE, txnMsg)
}

//version ack package
func ConstructVerAck(isConsensus bool) *mt.NetMessage {
	verAck := &netpb.VerAck{
		IsConsensus: isConsensus,
	}

	return mt.NewNetMessage(msgCommon.VERACK_TYPE, verAck)
}

//Version package
func ConstructVersion(n p2pnet.P2P, isCons bool, height uint32) *mt.NetMessage {
	versionPayload := &netpb.VersionPayload{
		Version:      n.GetVersion(),
		Services:     n.GetServices(),
		SyncPort:     uint32(n.GetSyncPort()),
		ConsPort:     uint32(n.GetConsPort()),
		UDPPort:      uint32(n.GetUDPPort()),
		Nonce:        n.GetID(),
		IsConsensus:  isCons,
		HttpInfoPort: uint32(n.GetHttpInfoPort()),
		StartHeight:  uint64(height),
		TimeStamp:    time.Now().UnixNano(),
	}

	version := &netpb.Version{
		P: versionPayload,
	}

	if n.GetRelay() {
		version.P.Relay = 1
	} else {
		version.P.Relay = 0
	}
	if config.DefConfig.P2PNode.HttpInfoPort > 0 {
		version.P.Cap[msgCommon.HTTP_INFO_FLAG] = 0x01
	} else {
		version.P.Cap[msgCommon.HTTP_INFO_FLAG] = 0x00
	}
	return mt.NewNetMessage(msgCommon.VERSION_TYPE, version)
}

//transaction request package
func ConstructTxnDataReq(hash common.Uint256) *mt.NetMessage {
	dataReq := &netpb.DataReq{
		DataType: netpb.InventoryType_TRANSACTION,
		Hash:     hash.ToArray(),
	}

	return mt.NewNetMessage(msgCommon.GET_DATA_TYPE, dataReq)
}

//block request package
func ConstructBlkDataReq(hash common.Uint256) *mt.NetMessage {
	dataReq := &netpb.DataReq{
		DataType: netpb.InventoryType_BLOCK,
		Hash:     hash.ToArray(),
	}

	return mt.NewNetMessage(msgCommon.GET_DATA_TYPE, dataReq)
}

//consensus request package
func ConstructConsensusDataReq(hash common.Uint256) *mt.NetMessage {
	dataReq := &netpb.DataReq{
		DataType: netpb.InventoryType_CONSENSUS,
		Hash:     hash.ToArray(),
	}

	return mt.NewNetMessage(msgCommon.GET_DATA_TYPE, dataReq)
}

//DHT ping message packet
func ConstructDHTPing(nodeID types.NodeID, udpPort, tcpPort uint16, ip net.IP, destAddr *net.UDPAddr, version uint16) *mt.NetMessage {
	ping := &netpb.DHTPing{}
	ping.Version = uint32(version)
	ping.FromID = nodeID.Bytes()

	srcEndPoint := &netpb.EndPoint{
		UDPPort: uint32(udpPort),
		TCPPort: uint32(tcpPort),
	}
	srcEndPoint.Addr = ip[:16]
	ping.SrcEndPoint = srcEndPoint

	destEndPoint := &netpb.EndPoint{
		UDPPort: uint32(destAddr.Port),
	}
	destIP := destAddr.IP.To16()
	destEndPoint.Addr = destIP[:16]
	ping.DestEndPoint = destEndPoint

	return mt.NewNetMessage(msgCommon.DHT_PING, ping)
}

//DHT pong message packet
func ConstructDHTPong(nodeID types.NodeID, udpPort, tcpPort uint16, ip net.IP, destAddr *net.UDPAddr, version uint16) *mt.NetMessage {
	pong := &netpb.DHTPong{}
	pong.Version = uint32(version)
	pong.FromID = nodeID.Bytes()

	srcEndPoint := &netpb.EndPoint{
		UDPPort: uint32(udpPort),
		TCPPort: uint32(tcpPort),
	}
	srcEndPoint.Addr = ip[:16]
	pong.SrcEndPoint = srcEndPoint

	destEndPoint := &netpb.EndPoint{
		UDPPort: uint32(destAddr.Port),
	}
	destIP := destAddr.IP.To16()
	destEndPoint.Addr = destIP[:16]
	pong.DestEndPoint = destEndPoint

	return mt.NewNetMessage(msgCommon.DHT_PONG, pong)
}

//DHT findNode message packet
func ConstructFindNode(nodeID types.NodeID, targetID types.NodeID) *mt.NetMessage {
	findNode := &netpb.FindNode{
		FromID:   nodeID.Bytes(),
		TargetID: targetID.Bytes(),
	}

	return mt.NewNetMessage(msgCommon.DHT_FIND_NODE, findNode)
}

//DHT neighbors message packet
func ConstructNeighbors(nodeID types.NodeID, nodes []*types.Node) *mt.NetMessage {
	neighbors := &netpb.Neighbors{
		FromID: nodeID.Bytes(),
		Nodes:  make([]*netpb.Node, 0, len(nodes)),
	}

	for _, node := range nodes {
		nodepb := &netpb.Node{
			ID:      node.ID.Bytes(),
			IP:      node.IP,
			UDPPort: uint32(node.UDPPort),
			TCPPort: uint32(node.TCPPort),
		}
		neighbors.Nodes = append(neighbors.Nodes, nodepb)
	}

	return mt.NewNetMessage(msgCommon.DHT_NEIGHBORS, neighbors)
}
