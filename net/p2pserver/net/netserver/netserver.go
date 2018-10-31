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

package netserver

import (
	"errors"
	//"sync"
	//"time"
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p"
	libcrypto "github.com/libp2p/go-libp2p-crypto"
	libhost "github.com/libp2p/go-libp2p-host"
	libnet "github.com/libp2p/go-libp2p-net"
	libpeer "github.com/libp2p/go-libp2p-peer"
	libpeerstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	//"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/net/p2pserver/common"
	//"github.com/ontio/ontology/p2pserver/message/msg_pack"
	"github.com/ontio/ontology/net/p2pserver/message/types"
	//"github.com/ontio/ontology/p2pserver/net/protocol"
	"github.com/ontio/ontology/net/p2pserver/peer"
)

const (
	ONTOLOGY_STREAM_V1 = "/ontology/1.0.0"
)

type NetworkManager struct {
	base       peer.PeerCom
	host       libhost.Host
	RecvCh     chan *types.MsgPayload
	stopCh     chan struct{}
	peerMgr    *peer.PeerManager
	networkKey libcrypto.PrivKey
}

func NewNetworkManager(port uint16) *NetworkManager {
	n := &NetworkManager{
		RecvCh: make(chan *types.MsgPayload, 100000),
		stopCh: make(chan struct{}),
	}

	err := n.init(port)
	if err != nil {
		log.Errorf("failed to initialize network manager, err %v", err)
		return nil
	}
	return n
}

//init initializes attribute of network server
func (this *NetworkManager) init(port uint16) error {
	this.base.SetVersion(common.PROTOCOL_VERSION)

	if config.DefConfig.Consensus.EnableConsensus {
		this.base.SetServices(uint64(common.VERIFY_NODE))
	} else {
		this.base.SetServices(uint64(common.SERVICE_NODE))
	}

	if config.DefConfig.P2PNode.NodePort == 0 {
		log.Error("[p2p]link port invalid")
		return errors.New("[p2p]invalid link port")
	}
	this.base.SetPort(port)
	this.base.SetRelay(true)

	var err error
	this.networkKey, err = initP2PNetworkKey()
	if err != nil {
		return err
	}
	id, err := libpeer.IDFromPublicKey(this.networkKey.GetPublic())
	if err != nil {
		return err
	}

	this.base.SetID(id)

	log.Infof("[p2p]init peer ID to %v", this.base.GetID().Pretty())

	sourceMultiAddr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", this.base.GetPort()))
	if err != nil {
		log.Errorf("failed to new multi addr err %v", err)
		return err
	}

	host, err := libp2p.New(context.Background(),
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(this.networkKey))
	if err != nil {
		log.Errorf("failed to create network manager, err %v", err)
		return nil
	}
	this.host = host

	this.host.SetStreamHandler(ONTOLOGY_STREAM_V1, this.handleStream)
	return nil
}

func (this *NetworkManager) addAddrToPeerstore(h libhost.Host, addr string) libpeer.ID {
	maddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		log.Errorf("failed to new multiaddr for %s, err %v", addr, err)
	}
	info, err := libpeerstore.InfoFromP2pAddr(maddr)
	if err != nil {
		log.Errorf("failed to get info from addr, err %v", err)
	}
	h.Peerstore().AddAddrs(info.ID, info.Addrs, libpeerstore.PermanentAddrTTL)
	return info.ID
}

func (this *NetworkManager) handleStream(s libnet.Stream) {
	log.Infof("Got a new stream")
}

func (this *NetworkManager) Connect(addr string) error {
	maddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		log.Errorf("failed to new multiaddr for %s", addr)
		return err
	}
	info, err := libpeerstore.InfoFromP2pAddr(maddr)
	if err != nil {
		log.Errorf("failed to get info from p2paddr, err %v", err)
		return err
	}
	this.host.Peerstore().AddAddrs(info.ID, info.Addrs, libpeerstore.PermanentAddrTTL)
	s, err := this.host.NewStream(context.Background(), info.ID, ONTOLOGY_STREAM_V1)
	if err != nil {
		log.Errorf("failed to new stream to target %v, addr %s, err %v", info.ID, addr, err)
		return err
	}
	log.Info(s)
	return nil
}
