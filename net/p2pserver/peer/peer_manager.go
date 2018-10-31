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
	"sync"
	//"github.com/ontio/ontology/p2pserver/common"
	//"github.com/ontio/ontology/p2pserver/message/types"
)

type PeerManager struct {
	mu               sync.Mutex
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
