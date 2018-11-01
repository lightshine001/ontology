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
	//"fmt"
	"github.com/ontio/ontology/common/log"
	"sync"

	libpeer "github.com/libp2p/go-libp2p-peer"
	"github.com/ontio/ontology/common"
	pCom "github.com/ontio/ontology/net/p2pserver/common"
	"github.com/ontio/ontology/net/p2pserver/message/types"
)

type PeerManager struct {
	mu               sync.RWMutex
	quitCh           chan struct{}
	allPeers         *sync.Map
	activePeersCount uint32
	maxPeerNum       uint32
	reservedPeerNum  uint32
}

func NewPeerManager() *PeerManager {
	pm := &PeerManager{
		quitCh:           make(chan struct{}),
		allPeers:         new(sync.Map),
		activePeersCount: 0,
		maxPeerNum:       32,
		reservedPeerNum:  16,
	}
	return pm
}

func (this *PeerManager) Count() uint32 {
	this.mu.RLock()
	defer this.mu.RUnlock()
	return this.activePeersCount
}

func (this *PeerManager) BroadcastMessage(msg types.Message, hash common.Uint256, priority int) {
	this.allPeers.Range(func(key, value interface{}) bool {
		peer := value.(*Peer)
		if peer.GetState() == pCom.ESTABLISH && peer.GetRelay() == true &&
			peer.IsHashContained(hash) {
			if msg.CmdType() == pCom.CONSENSUS_TYPE &&
				peer.GetServices() != pCom.VERIFY_NODE {
				return false
			}
			peer.MarkHashAsSeen(hash)
			err := peer.SendMessage(msg, priority)
			if err != nil {
				log.Infof("failed to send msg %s to peer %s, err %v",
					msg.CmdType(), peer.GetAddr().String(), err)
			}

		}
		return true
	})
}

func (this *PeerManager) AddPeer(p *Peer) {
	this.mu.Lock()
	defer this.mu.Unlock()

	if this.activePeersCount >= this.maxPeerNum {
		if p.stream != nil {
			p.stream.Close()
		}
		return
	}

	if v, ok := this.allPeers.Load(p.GetID().Pretty()); ok {
		old, _ := v.(*Peer)
		log.Infof("Removing old peer %s", p.GetID().Pretty())
		this.activePeersCount--
		this.allPeers.Delete(old.GetID().Pretty())

		if old.stream != nil {
			old.stream.Close()
		}
	}
	log.Infof("Added a new peer %s", p.String())
	this.activePeersCount++
	this.allPeers.Store(p.GetID().Pretty(), p)
	p.StartLoop()
}

func (this *PeerManager) RemovePeer(p *Peer) {
	this.mu.Lock()
	defer this.mu.Unlock()

	v, ok := this.allPeers.Load(p.GetID().Pretty())
	if !ok {
		return
	}

	exist, _ := v.(*Peer)
	if exist != p {
		return
	}

	log.Infof("Removing a peer %s", p.GetID().Pretty())
	this.activePeersCount--
	this.allPeers.Delete(p.GetID().Pretty())
}

func (this *PeerManager) FindByPeerID(peerID string) *Peer {
	v, _ := this.allPeers.Load(peerID)
	if v == nil {
		return nil
	}
	return v.(*Peer)
}

func (this *PeerManager) Find(pid libpeer.ID) *Peer {
	return this.FindByPeerID(pid.Pretty())
}

func (this *PeerManager) CloseStream(peerID string, reason error) {
	peer := this.FindByPeerID(peerID)
	if peer != nil {
		peer.close(reason)
	}
}

func (this *PeerManager) PeerEstablished(peerID string) bool {
	peer := this.FindByPeerID(peerID)
	if peer == nil {
		return false
	}
	if peer.GetState() != pCom.ESTABLISH {
		return false
	}
	return true
}

func (this *PeerManager) GetPeerAddrs() []pCom.PeerAddr {
	var addrs []pCom.PeerAddr
	this.allPeers.Range(func(key, value interface{}) bool {
		peer := value.(*Peer)
		if peer.GetState() != pCom.ESTABLISH {
			return false
		}
		var addr pCom.PeerAddr
		addr.Multiaddr = peer.GetAddr().String()
		addr.Time = peer.GetLatestReadAt()
		addr.Services = peer.GetServices()
		addr.ID = peer.GetID().Pretty()
		addrs = append(addrs, addr)
		return true
	})
	return addrs
}

func (this *PeerManager) GetPeerHeights() map[string]uint64 {
	hm := make(map[string]uint64)
	this.allPeers.Range(func(key, value interface{}) bool {
		peer := value.(*Peer)
		if peer.GetState() != pCom.ESTABLISH {
			return false
		}
		hm[peer.GetID().Pretty()] = peer.GetHeight()
		return true
	})
	return hm
}

func (this *PeerManager) GetPeers() []*Peer {
	peers := []*Peer{}
	this.allPeers.Range(func(key, value interface{}) bool {
		peer := value.(*Peer)
		if peer.GetState() != pCom.ESTABLISH {
			return false
		}
		peers = append(peers, peer)
		return true
	})
	return peers
}
