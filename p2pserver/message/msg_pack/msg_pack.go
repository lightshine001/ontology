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
	"time"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	ct "github.com/ontio/ontology/core/types"
	msgCommon "github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/dht/types"
	mt "github.com/ontio/ontology/p2pserver/message/types"
	p2pnet "github.com/ontio/ontology/p2pserver/net/protocol"
	"net"
)

///block package
func NewBlock(bk *ct.Block) mt.Message {
	log.Debug()
	var blk mt.Block
	blk.Blk = *bk

	return &blk
}

//blk hdr package
func NewHeaders(headers []*ct.Header) mt.Message {
	var blkHdr mt.BlkHeader
	blkHdr.BlkHdr = headers

	return &blkHdr
}

//blk hdr req package
func NewHeadersReq(curHdrHash common.Uint256) mt.Message {
	var h mt.HeadersReq
	h.Len = 1
	h.HashEnd = curHdrHash

	return &h
}

////Consensus info package
func NewConsensus(cp *mt.ConsensusPayload) mt.Message {
	log.Debug()
	var cons mt.Consensus
	cons.Cons = *cp
	cons.Hop = msgCommon.MAX_HOP

	return &cons
}

//InvPayload
func NewInvPayload(invType common.InventoryType, msg []common.Uint256) *mt.InvPayload {
	return &mt.InvPayload{
		InvType: invType,
		Blk:     msg,
	}
}

//Inv request package
func NewInv(invPayload *mt.InvPayload) mt.Message {
	var inv mt.Inv
	inv.P.Blk = invPayload.Blk
	inv.P.InvType = invPayload.InvType
	inv.Hop = msgCommon.MAX_HOP

	return &inv
}

//NotFound package
func NewNotFound(hash common.Uint256) mt.Message {
	log.Debug()
	var notFound mt.NotFound
	notFound.Hash = hash

	return &notFound
}

//ping msg package
func NewPingMsg(height uint64) *mt.Ping {
	log.Debug()
	var ping mt.Ping
	ping.Height = uint64(height)

	return &ping
}

//pong msg package
func NewPongMsg(height uint64) *mt.Pong {
	log.Debug()
	var pong mt.Pong
	pong.Height = uint64(height)

	return &pong
}

//Transaction package
func NewTxn(txn *ct.Transaction) mt.Message {
	log.Debug()
	var trn mt.Trn
	trn.Txn = txn
	trn.Hop = msgCommon.MAX_HOP

	return &trn
}

//version ack package
func NewVerAck(isConsensus bool) mt.Message {
	var verAck mt.VerACK
	verAck.IsConsensus = isConsensus

	return &verAck
}

//Version package
func NewVersion(n p2pnet.P2P, isCons bool, height uint32) mt.Message {
	var version mt.Version
	version.P = mt.VersionPayload{
		Version:      n.GetVersion(),
		Services:     n.GetServices(),
		SyncPort:     n.GetSyncPort(),
		ConsPort:     n.GetConsPort(),
		UDPPort:      n.GetUDPPort(),
		Nonce:        n.GetID(),
		IsConsensus:  isCons,
		HttpInfoPort: n.GetHttpInfoPort(),
		StartHeight:  uint64(height),
		TimeStamp:    time.Now().UnixNano(),
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
	return &version
}

//transaction request package
func NewTxnDataReq(hash common.Uint256) mt.Message {
	var dataReq mt.DataReq
	dataReq.DataType = common.TRANSACTION
	dataReq.Hash = hash

	return &dataReq
}

//block request package
func NewBlkDataReq(hash common.Uint256) mt.Message {
	var dataReq mt.DataReq
	dataReq.DataType = common.BLOCK
	dataReq.Hash = hash

	return &dataReq
}

//consensus request package
func NewConsensusDataReq(hash common.Uint256) mt.Message {
	var dataReq mt.DataReq
	dataReq.DataType = common.CONSENSUS
	dataReq.Hash = hash

	return &dataReq
}

//DHT ping message packet
func NewDHTPing(nodeID types.NodeID, udpPort, tcpPort uint16, ip net.IP, destAddr *net.UDPAddr, version uint16) mt.Message {
	ping := new(mt.DHTPing)
	ping.Version = version
	copy(ping.FromID[:], nodeID[:])

	ping.SrcEndPoint.UDPPort = udpPort
	ping.SrcEndPoint.TCPPort = tcpPort

	copy(ping.SrcEndPoint.Addr[:], ip[:16])

	ping.DestEndPoint.UDPPort = uint16(destAddr.Port)

	destIP := destAddr.IP.To16()
	copy(ping.DestEndPoint.Addr[:], destIP[:16])

	return ping
}

//DHT pong message packet
func NewDHTPong(nodeID types.NodeID, udpPort, tcpPort uint16, ip net.IP, destAddr *net.UDPAddr, version uint16) mt.Message {
	pong := new(mt.DHTPong)
	pong.Version = version
	copy(pong.FromID[:], nodeID[:])
	pong.SrcEndPoint.UDPPort = udpPort
	pong.SrcEndPoint.TCPPort = tcpPort

	copy(pong.SrcEndPoint.Addr[:], ip[:16])

	pong.DestEndPoint.UDPPort = uint16(destAddr.Port)
	destIP := destAddr.IP.To16()
	copy(pong.DestEndPoint.Addr[:], destIP[:16])

	return pong
}

//DHT findNode message packet
func NewFindNode(nodeID types.NodeID, targetID types.NodeID) mt.Message {
	findNode := &mt.FindNode{
		FromID:   nodeID,
		TargetID: targetID,
	}

	return findNode
}

//DHT neighbors message packet
func NewNeighbors(nodeID types.NodeID, nodes []*types.Node) mt.Message {
	neighbors := &mt.Neighbors{
		FromID: nodeID,
		Nodes:  make([]types.Node, 0, len(nodes)),
	}
	for _, node := range nodes {
		neighbors.Nodes = append(neighbors.Nodes, *node)
	}

	return neighbors
}
