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

package p2pserver

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	evtActor "github.com/ontio/ontology-eventbus/actor"
	comm "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/dht"
	dt "github.com/ontio/ontology/p2pserver/dht/types"
	"github.com/ontio/ontology/p2pserver/message/msg_pack"
	msgtypes "github.com/ontio/ontology/p2pserver/message/types"
	"github.com/ontio/ontology/p2pserver/message/utils"
	"github.com/ontio/ontology/p2pserver/net/netserver"
	p2pnet "github.com/ontio/ontology/p2pserver/net/protocol"
	"github.com/ontio/ontology/p2pserver/peer"
)

//P2PServer control all network activities
type P2PServer struct {
	network   p2pnet.P2P
	msgRouter *utils.MessageRouter
	pid       *evtActor.PID
	blockSync *BlockSyncMgr
	ledger    *ledger.Ledger
	ReconnectAddrs
	recentPeers    []dt.Node
	quitSyncRecent chan bool
	quitOnline     chan bool
	quitHeartBeat  chan bool
	dht            *dht.DHT
}

//ReconnectAddrs contain addr need to reconnect
type ReconnectAddrs struct {
	sync.RWMutex
	RetryAddrs map[string]int
}

//NewServer return a new p2pserver according to the pubkey
func NewServer() *P2PServer {
	id := dt.ConstructID(config.DefConfig.P2PNode.NetworkMgrCfg.DHT.IP,
		config.DefConfig.P2PNode.NetworkMgrCfg.DHT.UDPPort)
	n := netserver.NewNetServer(id)

	p := &P2PServer{
		network: n,
		ledger:  ledger.DefLedger,
	}

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, id)
	var nodeID dt.NodeID
	copy(nodeID[:], b[:])
	p.dht = dht.NewDHT(nodeID)
	p.network.SetFeedCh(p.dht.GetFeedCh())

	p.msgRouter = utils.NewMsgRouter(p.network)
	p.blockSync = NewBlockSyncMgr(p)
	p.recentPeers = make([]dt.Node, 0, common.RECENT_LIMIT)
	p.quitSyncRecent = make(chan bool)
	p.quitOnline = make(chan bool)
	p.quitHeartBeat = make(chan bool)
	return p
}

//GetConnectionCnt return the established connect count
func (this *P2PServer) GetConnectionCnt() uint32 {
	return this.network.GetConnectionCnt()
}

//Start create all services
func (this *P2PServer) Start() error {
	if this.network != nil {
		this.network.Start()
	} else {
		return errors.New("p2p network invalid")
	}
	if this.msgRouter != nil {
		this.msgRouter.Start()
	} else {
		return errors.New("p2p msg router invalid")
	}

	this.loadRecentPeers()
	this.dht.SetFallbackNodes(this.recentPeers)
	go this.syncUpRecentPeers()
	go this.keepOnlineService()
	go this.heartBeatService()
	go this.blockSync.Start()
	go this.dht.Start()
	go this.DisplayDHT()
	this.tryRecentPeers()
	return nil
}

func (this *P2PServer) DisplayDHT() {
	timer := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-timer.C:
			log.Info("DHT table is:")
			this.dht.DisplayRoutingTable()
			log.Info("Neighbor peers: ", this.GetConnectionCnt())
			peers := this.GetNeighborAddrs()
			for i, peer := range peers {
				var ip net.IP
				ip = peer.IpAddr[:]
				address := ip.To16().String() + ":" + strconv.Itoa(int(peer.Port))
				log.Infof("peer %d address is %s", i, address)
			}
		}
	}
}

//Stop halt all service by send signal to channels
func (this *P2PServer) Stop() {
	this.network.Halt()
	this.quitSyncRecent <- true
	this.quitOnline <- true
	this.quitHeartBeat <- true
	this.msgRouter.Stop()
	this.blockSync.Close()
	this.dht.Stop()
}

// GetNetWork returns the low level netserver
func (this *P2PServer) GetNetWork() p2pnet.P2P {
	return this.network
}

//GetPort return two network port
func (this *P2PServer) GetPort() (uint16, uint16) {
	return this.network.GetSyncPort(), this.network.GetConsPort()
}

//GetVersion return self version
func (this *P2PServer) GetVersion() uint32 {
	return this.network.GetVersion()
}

//GetNeighborAddrs return all nbr`s address
func (this *P2PServer) GetNeighborAddrs() []common.PeerAddr {
	return this.network.GetNeighborAddrs()
}

func (this *P2PServer) GetDHT() *dht.DHT {
	return this.dht
}

//Xmit called by other module to broadcast msg
func (this *P2PServer) Xmit(message interface{}) error {
	log.Debug()
	var msg msgtypes.Message
	var msgHash comm.Uint256
	isConsensus := false
	switch message.(type) {
	case *types.Transaction:
		log.Debug("TX transaction message")
		txn := message.(*types.Transaction)
		msg = msgpack.NewTxn(txn)
		msgHash = txn.Hash()
	case *types.Block:
		log.Debug("TX block message")
		block := message.(*types.Block)
		msg = msgpack.NewBlock(block)
		msgHash = block.Hash()
	case *msgtypes.ConsensusPayload:
		log.Debug("TX consensus message")
		consensusPayload := message.(*msgtypes.ConsensusPayload)
		msg = msgpack.NewConsensus(consensusPayload)
		isConsensus = true
		msgHash = consensusPayload.Hash()
	case comm.Uint256:
		log.Debug("TX block hash message")
		hash := message.(comm.Uint256)
		// construct inv message
		invPayload := msgpack.NewInvPayload(comm.BLOCK, []comm.Uint256{hash})
		msg = msgpack.NewInv(invPayload)
		msgHash = hash
	default:
		log.Warnf("Unknown Xmit message %v , type %v", message,
			reflect.TypeOf(message))
		return errors.New("Unknown Xmit message type")
	}
	this.network.Xmit(msg, msgHash, isConsensus)
	return nil
}

//Send tranfer buffer to peer
func (this *P2PServer) Send(p *peer.Peer, msg msgtypes.Message,
	isConsensus bool) error {
	if this.network.IsPeerEstablished(p) {
		return this.network.Send(p, msg, isConsensus)
	}
	log.Errorf("P2PServer send to a not ESTABLISH peer %d",
		p.GetID())
	return errors.New("send to a not ESTABLISH peer")
}

// GetID returns local node id
func (this *P2PServer) GetID() uint64 {
	return this.network.GetID()
}

// OnAddNode adds the peer id to the block sync mgr
func (this *P2PServer) OnAddNode(id uint64) {
	this.blockSync.OnAddNode(id)
}

// OnDelNode removes the peer id from the block sync mgr
func (this *P2PServer) OnDelNode(id uint64) {
	this.blockSync.OnDelNode(id)
}

// OnHeaderReceive adds the header list from network
func (this *P2PServer) OnHeaderReceive(headers []*types.Header) {
	this.blockSync.OnHeaderReceive(headers)
}

// OnBlockReceive adds the block from network
func (this *P2PServer) OnBlockReceive(block *types.Block) {
	this.blockSync.OnBlockReceive(block)
}

// Todo: remove it if no use
func (this *P2PServer) GetConnectionState() uint32 {
	return common.INIT
}

//GetTime return lastet contact time
func (this *P2PServer) GetTime() int64 {
	return this.network.GetTime()
}

// SetPID sets p2p actor
func (this *P2PServer) SetPID(pid *evtActor.PID) {
	this.pid = pid
	this.msgRouter.SetPID(pid)
}

// GetPID returns p2p actor
func (this *P2PServer) GetPID() *evtActor.PID {
	return this.pid
}

//blockSyncFinished compare all nbr peers and self height at beginning
func (this *P2PServer) blockSyncFinished() bool {
	peers := this.network.GetNeighbors()
	if len(peers) == 0 {
		return false
	}

	blockHeight := this.ledger.GetCurrentBlockHeight()

	for _, v := range peers {
		if blockHeight < uint32(v.GetHeight()) {
			return false
		}
	}
	return true
}

//WaitForSyncBlkFinish compare the height of self and remote peer in loop
func (this *P2PServer) WaitForSyncBlkFinish() {
	consensusType := strings.ToLower(config.DefConfig.Genesis.ConsensusType)
	if consensusType == "solo" {
		return
	}

	for {
		headerHeight := this.ledger.GetCurrentHeaderHeight()
		currentBlkHeight := this.ledger.GetCurrentBlockHeight()
		log.Info("WaitForSyncBlkFinish... current block height is ",
			currentBlkHeight, " ,current header height is ", headerHeight)

		if this.blockSyncFinished() {
			break
		}

		<-time.After(time.Second * (time.Duration(common.SYNC_BLK_WAIT)))
	}
}

//WaitForPeersStart check whether enough peer linked in loop
func (this *P2PServer) WaitForPeersStart() {
	periodTime := config.DEFAULT_GEN_BLOCK_TIME / common.UPDATE_RATE_PER_BLOCK
	for {
		log.Info("Wait for minimum connection...")
		if this.reachMinConnection() {
			break
		}

		<-time.After(time.Second * (time.Duration(periodTime)))
	}
}

//reachMinConnection return whether net layer have enough link under different config
func (this *P2PServer) reachMinConnection() bool {
	consensusType := strings.ToLower(config.DefConfig.Genesis.ConsensusType)
	if consensusType == "" {
		consensusType = "dbft"
	}
	minCount := config.DBFT_MIN_NODE_NUM
	switch consensusType {
	case "dbft":
	case "solo":
		minCount = config.SOLO_MIN_NODE_NUM
	case "vbft":
		minCount = config.VBFT_MIN_NODE_NUM

	}
	return int(this.GetConnectionCnt())+1 >= minCount
}

//getNode returns the peer with the id
func (this *P2PServer) getNode(id uint64) *peer.Peer {
	return this.network.GetPeer(id)
}

//retryInactivePeer try to connect peer in INACTIVITY state
func (this *P2PServer) retryInactivePeer() {
	np := this.network.GetNp()
	np.Lock()
	var ip net.IP
	neighborPeers := make(map[uint64]*peer.Peer)
	for _, p := range np.List {
		addr, _ := p.GetAddr16()
		ip = addr[:]
		nodeAddr := ip.To16().String() + ":" +
			strconv.Itoa(int(p.GetSyncPort()))
		if p.GetSyncState() == common.INACTIVITY {
			log.Infof(" try reconnect %s", nodeAddr)
			//add addr to retry list
			this.addToRetryList(nodeAddr)
			p.CloseSync()
			p.CloseCons()
		} else {
			//add others to tmp node map
			this.removeFromRetryList(nodeAddr)
			neighborPeers[p.GetID()] = p
		}
	}

	np.List = neighborPeers
	np.Unlock()

	connCount := uint(this.network.GetOutConnRecordLen())
	if connCount >= config.DefConfig.P2PNode.MaxConnOutBound {
		log.Warnf("Connect: out connections(%d) reach the max limit(%d)", connCount,
			config.DefConfig.P2PNode.MaxConnOutBound)
		return
	}

	//try connect
	if len(this.RetryAddrs) > 0 {
		this.ReconnectAddrs.Lock()

		list := make(map[string]int)
		addrs := make([]string, 0, len(this.RetryAddrs))
		for addr, v := range this.RetryAddrs {
			v += 1
			addrs = append(addrs, addr)
			if v < common.MAX_RETRY_COUNT {
				list[addr] = v
			}
			if v >= common.MAX_RETRY_COUNT {
				this.network.RemoveFromConnectingList(addr)
				remotePeer := this.network.GetPeerFromAddr(addr)
				if remotePeer != nil {
					if remotePeer.SyncLink.GetAddr() == addr {
						this.network.RemovePeerSyncAddress(addr)
						this.network.RemovePeerConsAddress(addr)
					}
					if remotePeer.ConsLink.GetAddr() == addr {
						this.network.RemovePeerConsAddress(addr)
					}
					this.network.DelNbrNode(remotePeer.GetID())
				}
			}
		}

		this.RetryAddrs = list
		this.ReconnectAddrs.Unlock()
		for _, addr := range addrs {
			rand.Seed(time.Now().UnixNano())
			log.Info("Try to reconnect peer, peer addr is ", addr)
			<-time.After(time.Duration(rand.Intn(common.CONN_MAX_BACK)) * time.Millisecond)
			log.Info("Back off time`s up, start connect node")
			this.network.Connect(addr, false)
		}

	}
}

//keepOnline try connect lost peer
func (this *P2PServer) keepOnlineService() {
	t := time.NewTimer(time.Second * common.CONN_MONITOR)
	for {
		select {
		case <-t.C:
			this.retryInactivePeer()
			t.Stop()
			t.Reset(time.Second * common.CONN_MONITOR)
		case <-this.quitOnline:
			t.Stop()
			break
		}
	}
}

//heartBeat send ping to nbr peers and check the timeout
func (this *P2PServer) heartBeatService() {
	var periodTime uint
	periodTime = config.DEFAULT_GEN_BLOCK_TIME / common.UPDATE_RATE_PER_BLOCK
	t := time.NewTicker(time.Second * (time.Duration(periodTime)))

	for {
		select {
		case <-t.C:
			this.ping()
			this.timeout()
		case <-this.quitHeartBeat:
			t.Stop()
			break
		}
	}
}

//ping send pkg to get pong msg from others
func (this *P2PServer) ping() {
	peers := this.network.GetNeighbors()
	for _, p := range peers {
		if p.GetSyncState() == common.ESTABLISH {
			height := this.ledger.GetCurrentBlockHeight()
			ping := msgpack.NewPingMsg(uint64(height))
			go this.Send(p, ping, false)
		}
	}
}

//timeout trace whether some peer be long time no response
func (this *P2PServer) timeout() {
	peers := this.network.GetNeighbors()
	var periodTime uint
	periodTime = config.DEFAULT_GEN_BLOCK_TIME / common.UPDATE_RATE_PER_BLOCK
	for _, p := range peers {
		if p.GetSyncState() == common.ESTABLISH {
			t := p.GetContactTime()
			if t.Before(time.Now().Add(-1 * time.Second *
				time.Duration(periodTime) * common.KEEPALIVE_TIMEOUT)) {
				log.Warnf("keep alive timeout!!!lost remote peer %d - %s from %s", p.GetID(), p.SyncLink.GetAddr(), t.String())
				p.CloseSync()
				p.CloseCons()
			}
		}
	}
}

//addToRetryList add retry address to ReconnectAddrs
func (this *P2PServer) addToRetryList(addr string) {
	this.ReconnectAddrs.Lock()
	defer this.ReconnectAddrs.Unlock()
	if this.RetryAddrs == nil {
		this.RetryAddrs = make(map[string]int)
	}
	if _, ok := this.RetryAddrs[addr]; ok {
		delete(this.RetryAddrs, addr)
	}
	//alway set retry to 0
	this.RetryAddrs[addr] = 0
}

//removeFromRetryList remove connected address from ReconnectAddrs
func (this *P2PServer) removeFromRetryList(addr string) {
	this.ReconnectAddrs.Lock()
	defer this.ReconnectAddrs.Unlock()
	if len(this.RetryAddrs) > 0 {
		if _, ok := this.RetryAddrs[addr]; ok {
			delete(this.RetryAddrs, addr)
		}
	}
}

//loadRecentPeers loads latest remote peers
func (this *P2PServer) loadRecentPeers() {
	if comm.FileExisted(common.RECENT_FILE_NAME) {
		buf, err := ioutil.ReadFile(common.RECENT_FILE_NAME)
		if err != nil {
			log.Error("read %s fail:%s, connect recent peers cancel",
				common.RECENT_FILE_NAME, err.Error())
			return
		}

		err = json.Unmarshal(buf, &this.recentPeers)
		if err != nil {
			log.Error("parse recent peer file fail: ", err)
			return
		}
	}
}

//tryRecentPeers try connect recent contact peer when service start
func (this *P2PServer) tryRecentPeers() {
	for _, v := range this.recentPeers {
		addr := v.IP + ":" + strconv.Itoa(int(v.TCPPort))
		go this.network.Connect(addr, false)
	}

}

//syncUpRecentPeers sync up recent peers periodically
func (this *P2PServer) syncUpRecentPeers() {
	periodTime := common.RECENT_TIMEOUT
	t := time.NewTicker(time.Second * (time.Duration(periodTime)))
	for {
		select {
		case <-t.C:
			this.syncPeerAddr()
		case <-this.quitSyncRecent:
			t.Stop()
			break
		}
	}

}

//syncPeerAddr compare snapshot of recent peer with current link,then persist the list
func (this *P2PServer) syncPeerAddr() {
	changed := false
	for i := 0; i < len(this.recentPeers); i++ {
		addr := this.recentPeers[i].IP + ":" + strconv.Itoa(int(this.recentPeers[i].TCPPort))
		p := this.network.GetPeerFromAddr(addr)
		if p == nil || (p != nil && p.GetSyncState() != common.ESTABLISH) {
			this.recentPeers = append(this.recentPeers[:i], this.recentPeers[i+1:]...)
			changed = true
			i--
		}
	}
	left := common.RECENT_LIMIT - len(this.recentPeers)
	if left > 0 {
		np := this.network.GetNp()
		np.Lock()
		found := false
		for _, p := range np.List {
			addrIp, err := common.ParseIPAddr(p.GetAddr())
			if err != nil {
				log.Info(err)
				continue
			}
			port := p.GetSyncPort()
			found = false
			for i := 0; i < len(this.recentPeers); i++ {
				if this.recentPeers[i].IP == addrIp &&
					this.recentPeers[i].TCPPort == port {
					found = true
					break
				}
			}

			if !found {
				node := dt.Node{
					IP:      addrIp,
					UDPPort: p.GetUDPPort(),
					TCPPort: port,
				}

				id := dt.ConstructID(addrIp, p.GetUDPPort())
				b := make([]byte, 8)
				binary.LittleEndian.PutUint64(b, id)
				copy(node.ID[:], b[:])

				this.recentPeers = append(this.recentPeers, node)
				left--
				changed = true
				if left == 0 {
					break
				}
			}
		}
		np.Unlock()
	}
	if changed {
		buf, err := json.Marshal(this.recentPeers)
		if err != nil {
			log.Error("package recent peer fail: ", err)
			return
		}
		err = ioutil.WriteFile(common.RECENT_FILE_NAME, buf, os.ModePerm)
		if err != nil {
			log.Error("write recent peer fail: ", err)
		}
	}
}
