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

package peer

import (
	//"errors"
	"bufio"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fatih/set"
	libnet "github.com/libp2p/go-libp2p-net"
	libPeer "github.com/libp2p/go-libp2p-peer"
	ma "github.com/multiformats/go-multiaddr"
	comm "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/net/p2pserver/common"
	"github.com/ontio/ontology/net/p2pserver/message/types"
)

// PeerCom provides the basic information of a peer
type PeerCom struct {
	id           libPeer.ID
	addr         ma.Multiaddr
	version      uint32
	services     uint64
	relay        bool
	httpInfoPort uint16
	port         uint16
	height       uint64
}

// SetID sets a peer's id
func (this *PeerCom) SetID(id libPeer.ID) {
	this.id = id
}

// GetID returns a peer's id
func (this *PeerCom) GetID() libPeer.ID {
	return this.id
}

func (this *PeerCom) SetAddr(addr ma.Multiaddr) {
	this.addr = addr
}

func (this *PeerCom) GetAddr() ma.Multiaddr {
	return this.addr
}

// SetVersion sets a peer's version
func (this *PeerCom) SetVersion(version uint32) {
	this.version = version
}

// GetVersion returns a peer's version
func (this *PeerCom) GetVersion() uint32 {
	return this.version
}

// SetServices sets a peer's services
func (this *PeerCom) SetServices(services uint64) {
	this.services = services
}

// GetServices returns a peer's services
func (this *PeerCom) GetServices() uint64 {
	return this.services
}

// SerRelay sets a peer's relay
func (this *PeerCom) SetRelay(relay bool) {
	this.relay = relay
}

// GetRelay returns a peer's relay
func (this *PeerCom) GetRelay() bool {
	return this.relay
}

// SetSyncPort sets a peer's port
func (this *PeerCom) SetPort(port uint16) {
	this.port = port
}

// GetSyncPort returns a peer's sync port
func (this *PeerCom) GetPort() uint16 {
	return this.port
}

// SetHttpInfoPort sets a peer's http info port
func (this *PeerCom) SetHttpInfoPort(port uint16) {
	this.httpInfoPort = port
}

// GetHttpInfoPort returns a peer's http info port
func (this *PeerCom) GetHttpInfoPort() uint16 {
	return this.httpInfoPort
}

// SetHeight sets a peer's height
func (this *PeerCom) SetHeight(height uint64) {
	this.height = height
}

// GetHeight returns a peer's height
func (this *PeerCom) GetHeight() uint64 {
	return this.height
}

//Peer represent the node in p2p
type Peer struct {
	base                    PeerCom
	mu                      sync.Mutex
	cap                     [32]byte
	stream                  libnet.Stream
	state                   uint32
	latestReadAt            int64
	latestWriteAt           int64
	recvCh                  chan *types.MsgPayload
	highPriorityMessageCh   chan types.Message
	normalPriorityMessageCh chan types.Message
	lowPriorityMessageCh    chan types.Message
	msgCount                map[string]uint64
	reqRecord               map[string]int64
	knownHash               set.Interface
	quitWriteCh             chan struct{}
}

func NewPeer(stream libnet.Stream, recvCh chan *types.MsgPayload) *Peer {
	p := &Peer{
		state:                   common.INIT,
		knownHash:               set.New(set.ThreadSafe),
		stream:                  stream,
		recvCh:                  recvCh,
		highPriorityMessageCh:   make(chan types.Message, 10000),
		normalPriorityMessageCh: make(chan types.Message, 10000),
		lowPriorityMessageCh:    make(chan types.Message, 100000),
		msgCount:                make(map[string]uint64),
		reqRecord:               make(map[string]int64),
		quitWriteCh:             make(chan struct{}),
	}
	p.base.id = stream.Conn().RemotePeer()
	p.base.addr = stream.Conn().RemoteMultiaddr()
	runtime.SetFinalizer(p, rmPeer)
	return p
}

//rmPeer print a debug log when peer be finalized by system
func rmPeer(p *Peer) {
	log.Debugf("[p2p]Remove unused peer: %d", p.GetID())
}

//DumpInfo print all information of peer
func (this *Peer) DumpInfo() {
	log.Debug("[p2p]Node info:")
	log.Debug("[p2p]\t State = ", this.state)
	log.Debug("[p2p]\t id = ", this.GetID().Pretty())
	log.Debug("[p2p]\t addr = ", this.GetAddr())
	log.Debug("[p2p]\t cap = ", this.cap)
	log.Debug("[p2p]\t version = ", this.GetVersion())
	log.Debug("[p2p]\t services = ", this.GetServices())
	log.Debug("[p2p]\t port = ", this.GetPort())
	log.Debug("[p2p]\t relay = ", this.GetRelay())
	log.Debug("[p2p]\t height = ", this.GetHeight())
}

//GetVersion return peer`s version
func (this *Peer) GetVersion() uint32 {
	return this.base.GetVersion()
}

//GetHeight return peer`s block height
func (this *Peer) GetHeight() uint64 {
	return this.base.GetHeight()
}

//SetHeight set height to peer
func (this *Peer) SetHeight(height uint64) {
	this.base.SetHeight(height)
}

//GetSyncState return sync state
func (this *Peer) GetState() uint32 {
	return this.state
}

//SetSyncState set sync state to peer
func (this *Peer) SetState(state uint32) {
	atomic.StoreUint32(&(this.state), state)
}

//GetSyncPort return peer`s sync port
func (this *Peer) GetPort() uint16 {
	return this.base.GetPort()
}

//SetConsPort set peer`s consensus port
func (this *Peer) SetPort(port uint16) {
	this.base.SetPort(port)
}

//CloseSync halt sync connection
func (this *Peer) close(reason error) {
	// Add lock & close flag to prevent multi call.
	this.mu.Lock()
	defer this.mu.Unlock()

	if this.state == common.INACTIVITY {
		return
	}
	this.state = common.INACTIVITY
	log.Infof("Closing stream %s due to err %v", this.String(), reason)
	// quit.
	this.quitWriteCh <- struct{}{}

	// close stream.
	if this.stream != nil {
		this.stream.Close()
	}
}

//GetID return peer`s id
func (this *Peer) GetID() libPeer.ID {
	return this.base.GetID()
}

//GetRelay return peer`s relay state
func (this *Peer) GetRelay() bool {
	return this.base.GetRelay()
}

//GetServices return peer`s service state
func (this *Peer) GetServices() uint64 {
	return this.base.GetServices()
}

//GetAddr return peer`s sync link address
func (this *Peer) GetAddr() ma.Multiaddr {
	return this.base.GetAddr()
}

func (this *Peer) SetLatestReadAt(time int64) {
	this.latestReadAt = time
}

func (this *Peer) GetLatestReadAt() int64 {
	return this.latestReadAt
}

func (this *Peer) SetLatestWriteAt(time int64) {
	this.latestWriteAt = time
}

func (this *Peer) GetLatestWriteAt() int64 {
	return this.latestWriteAt
}

//SetHttpInfoState set peer`s httpinfo state
func (this *Peer) SetHttpInfoState(httpInfo bool) {
	if httpInfo {
		this.cap[common.HTTP_INFO_FLAG] = 0x01
	} else {
		this.cap[common.HTTP_INFO_FLAG] = 0x00
	}
}

//GetHttpInfoState return peer`s httpinfo state
func (this *Peer) GetHttpInfoState() bool {
	return this.cap[common.HTTP_INFO_FLAG] == 1
}

//GetHttpInfoPort return peer`s httpinfo port
func (this *Peer) GetHttpInfoPort() uint16 {
	return this.base.GetHttpInfoPort()
}

//SetHttpInfoPort set peer`s httpinfo port
func (this *Peer) SetHttpInfoPort(port uint16) {
	this.base.SetHttpInfoPort(port)
}

func (this *Peer) SendMessage(msg types.Message, priority int) error {
	switch priority {
	case common.MessagePriorityHigh:
		this.highPriorityMessageCh <- msg
	case common.MessagePriorityNormal:
		select {
		case this.normalPriorityMessageCh <- msg:
		default:
			return fmt.Errorf("Received too many normal priority message")
		}
	default:
		select {
		case this.lowPriorityMessageCh <- msg:
		default:
			fmt.Errorf("Received too many low priority message.")
		}

	}
	return nil
}

func (this *Peer) String() string {
	addrStr := ""
	if this.base.addr != nil {
		addrStr = this.base.GetAddr().String()
	}

	return fmt.Sprintf("Peer Stream: %s,%s", this.base.id.Pretty(), addrStr)
}

func (this *Peer) UpdateInfo(t int64, version uint32,
	services uint64, port uint16, relay uint8, height uint64) {
	this.SetLatestReadAt(t)
	this.base.SetVersion(version)
	this.base.SetServices(services)
	this.base.SetPort(port)
	if relay == 0 {
		this.base.SetRelay(false)
	} else {
		this.base.SetRelay(true)
	}
	this.SetHeight(height)
}

func (this *Peer) StartLoop() {
	go this.readLoop()
	go this.writeLoop()
}

func (this *Peer) IsConnected() bool {
	return this.stream != nil
}

func (this *Peer) needSendMsg(msg types.Message) bool {
	if msg.CmdType() != common.GET_DATA_TYPE {
		return true
	}

	var dataReq = msg.(*types.DataReq)
	reqID := fmt.Sprintf("%x%s", dataReq.DataType, dataReq.Hash.ToHexString())
	now := time.Now().Unix()
	if t, ok := this.reqRecord[reqID]; ok {
		if int(now-t) < common.REQ_INTERVAL {
			return false
		}
	}
	return true
}

func (this *Peer) addReqRecord(msg types.Message) {
	if msg.CmdType() != common.GET_DATA_TYPE {
		return
	}
	now := time.Now().Unix()
	if len(this.reqRecord) >= common.MAX_REQ_RECORD_SIZE-1 {
		for id := range this.reqRecord {
			t := this.reqRecord[id]
			if int(now-t) > common.REQ_INTERVAL {
				delete(this.reqRecord, id)
			}
		}
	}
	var dataReq = msg.(*types.DataReq)
	reqID := fmt.Sprintf("%x%s", dataReq.DataType, dataReq.Hash.ToHexString())
	this.reqRecord[reqID] = now
}

func (this *Peer) readLoop() {
	if !this.IsConnected() {
		log.Errorf("peer %s is not connected", this.GetID().Pretty())
		return
	}

	reader := bufio.NewReaderSize(this.stream, common.MAX_BUF_LEN)

	for {
		msg, payloadSize, err := types.ReadMessage(reader)
		if err != nil {
			log.Infof("[p2p]error read from %s :%s", this.GetAddr(), err.Error())
			break
		}

		t := time.Now().Unix()
		this.SetLatestReadAt(t)

		if !this.needSendMsg(msg) {
			log.Infof("skip handle msgType:%s from:%s", msg.CmdType(), this.GetID().Pretty())
			continue
		}
		this.addReqRecord(msg)
		this.msgCount[msg.CmdType()]++

		this.recvCh <- &types.MsgPayload{
			Id:          this.GetID(),
			Addr:        this.base.GetAddr(),
			PayloadSize: payloadSize,
			Payload:     msg,
		}

	}

	//this.disconnectNotify()
}

func (this *Peer) writeLoop() {
	for {
		select {
		case <-this.quitWriteCh:
			log.Infof("Quit stream write loop")
			return
		}

		select {
		case message, ok := <-this.highPriorityMessageCh:
			if ok {
				this.tx(message)
				continue
			}
		default:
		}

		select {
		case message, ok := <-this.normalPriorityMessageCh:
			if ok {
				this.tx(message)
				continue
			}
		default:
		}
		select {
		case message, ok := <-this.lowPriorityMessageCh:
			if ok {
				this.tx(message)
			}
		default:
		}
	}

}

func (this *Peer) tx(msg types.Message) {
	if !this.IsConnected() {
		log.Errorf("peer %s is not connected", this.GetID().Pretty())
		return
	}

	sink := comm.NewZeroCopySink(nil)
	err := types.WriteMessage(sink, msg)
	if err != nil {
		log.Errorf("[p2p]error serialize messge ", err.Error())
		return
	}

	payload := sink.Bytes()
	nByteCnt := len(payload)
	log.Tracef("[p2p]TX buf length: %d\n", nByteCnt)

	nCount := nByteCnt / common.PER_SEND_LEN
	if nCount == 0 {
		nCount = 1
	}
	deadline := time.Now().Add(time.Duration(nCount*common.WRITE_DEADLINE) * time.Second)
	if err := this.stream.SetWriteDeadline(deadline); err != nil {
		log.Errorf("[p2p]failed to set write deadline, err %v", err)
		return
	}
	_, err = this.stream.Write(payload)
	if err != nil {
		log.Infof("[p2p]error sending messge to %s :%s", this.base.GetAddr().String(), err.Error())
		//this.disconnectNotify()
		return
	}
	this.SetLatestWriteAt(time.Now().Unix())
}

func (this *Peer) MarkHashAsSeen(hash comm.Uint256) {
	if this.knownHash.Size() >= common.MAX_CACHE_SIZE {
		this.knownHash.Pop()
	}
	this.knownHash.Add(hash)
}

func (this *Peer) IsHashContained(hash comm.Uint256) bool {
	return this.knownHash.Has(hash)
}
