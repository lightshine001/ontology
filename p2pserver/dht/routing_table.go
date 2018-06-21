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

package dht

import (
	"sync"

	"fmt"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/p2pserver/dht/types"
	"github.com/ontio/ontology/p2pserver/dht/utils"
	"io"
	"strconv"
)

const (
	DHT_BLACK_LIST_FILE = "./dht_black_list"
	DHT_WHITE_LIST_FILE = "./dht_white_list"
	DHT_ROUTING_TABLE   = "./dht_routing_table"
)

type bucket struct {
	entries []*types.Node
}

type routingTable struct {
	mu      sync.RWMutex
	id      types.NodeID
	buckets []*bucket
	feedCh  chan *types.FeedEvent

	whiteList []string
	blackList []string
}

func (this *routingTable) init(id types.NodeID, ch chan *types.FeedEvent) {
	this.buckets = make([]*bucket, types.BUCKET_NUM)
	for i := range this.buckets {
		this.buckets[i] = &bucket{
			entries: make([]*types.Node, 0, types.BUCKET_SIZE),
		}
	}

	this.id = id
	this.feedCh = ch

	// load white list and black list
	this.whiteList = utils.LoadList(DHT_WHITE_LIST_FILE)
	this.blackList = utils.LoadList(DHT_BLACK_LIST_FILE)
}

func (this *routingTable) locateBucket(id types.NodeID) (int, *bucket) {
	dist := logdist(this.id, id)
	if dist == 0 {
		return 0, this.buckets[0]
	}
	return dist - 1, this.buckets[dist-1]
}

func (this *routingTable) queryNode(id types.NodeID) *types.Node {
	this.mu.RLock()
	defer this.mu.RUnlock()
	_, bucket := this.locateBucket(id)
	for _, node := range bucket.entries {
		if (*node).ID == id {
			return node
		}
	}
	return nil
}

// add node to bucket, if bucket contains the node, move it to bucket head
func (this *routingTable) addNode(node *types.Node, bucketIndex int) bool {
	this.mu.Lock()
	defer this.mu.Unlock()

	nodeAddress := node.IP + ":" + strconv.Itoa(int(node.UDPPort))
	if !utils.IsContained(this.whiteList, nodeAddress) {
		return false
	}
	bucket := this.buckets[bucketIndex]
	for i, entry := range bucket.entries {
		if entry.ID == node.ID {
			copy(bucket.entries[1:], bucket.entries[:i])
			bucket.entries[0] = node
			return false
		}
	}

	// Todo: if the bucket is full, use LRU to replace
	if len(bucket.entries) >= types.BUCKET_SIZE {
		// bucket is full
		return false
	}

	bucket.entries = append(bucket.entries, node)

	copy(bucket.entries[1:], bucket.entries[:])
	bucket.entries[0] = node
	feed := &types.FeedEvent{
		EvtType: types.Add,
		Event:   node,
	}
	this.feedCh <- feed
	return true
}

func (this *routingTable) removeNode(id types.NodeID) {
	this.mu.Lock()
	defer this.mu.Unlock()
	_, bucket := this.locateBucket(id)

	entries := bucket.entries[:0]
	var node *types.Node
	for _, entry := range bucket.entries {
		if entry.ID != id {
			entries = append(entries, entry)
		} else {
			node = entry
		}
	}
	bucket.entries = entries

	if node != nil {
		feed := &types.FeedEvent{
			EvtType: types.Del,
			Event:   node,
		}
		this.feedCh <- feed
	}
}

func (this *routingTable) getClosestNodes(num int, targetID types.NodeID) []*types.Node {
	this.mu.RLock()
	defer this.mu.RUnlock()
	closestList := make([]*types.Node, 0, num)

	index, _ := this.locateBucket(targetID)
	buckets := []int{index}
	i := index - 1
	j := index + 1

	for len(buckets) < types.BUCKET_NUM {
		if j < types.BUCKET_NUM {
			buckets = append(buckets, j)
		}
		if i >= 0 {
			buckets = append(buckets, i)
		}
		i--
		j++
	}

	for index := range buckets {
		for _, entry := range this.buckets[index].entries {
			closestList = append(closestList, entry)
			if len(closestList) >= num {
				return closestList
			}
		}
	}
	return closestList
}

func (this *routingTable) getTotalNodeNumInBukcet(bucket int) int {
	this.mu.RLock()
	defer this.mu.RUnlock()
	b := this.buckets[bucket]
	if b == nil {
		return 0
	}

	return len(b.entries)
}

func (this *routingTable) getLastNodeInBucket(bucket int) *types.Node {
	this.mu.RLock()
	defer this.mu.RUnlock()
	b := this.buckets[bucket]
	if b == nil {
		return nil
	}

	return b.entries[len(b.entries)-1]
}

func (this *routingTable) getDistance(id1, id2 types.NodeID) int {
	dist := logdist(id1, id2)
	return dist
}

func (this *routingTable) totalNodes() int {
	this.mu.RLock()
	defer this.mu.RUnlock()
	var num int
	for _, bucket := range this.buckets {
		num += len(bucket.entries)
	}
	return num
}

func (this *routingTable) isNodeInBucket(id types.NodeID, bucket int) (*types.Node, bool) {
	this.mu.RLock()
	defer this.mu.RUnlock()

	b := this.buckets[bucket]
	if b == nil {
		return nil, false
	}

	for _, entry := range b.entries {
		if entry.ID == id {
			return entry, true
		}
	}
	return nil, false
}

func (this *routingTable) serialize(w io.Writer) error {
	if _, err := w.Write(this.id[:]); err != nil {
		return fmt.Errorf("serialize dht error: serialize ID err, %s", err)
	}
	notEmptyNum := 0
	for _, buck := range this.buckets {
		if buck != nil && len(buck.entries) > 0 {
			notEmptyNum++
		}
	}
	if err := serialization.WriteVarUint(w, uint64(notEmptyNum)); err != nil {
		return fmt.Errorf("serialize dht error: serialize not empty bucket num err, %s", err)
	}
	for index, buck := range this.buckets {
		if buck != nil && len(buck.entries) > 0 {
			if err := serialization.WriteVarUint(w, uint64(index)); err != nil {
				return fmt.Errorf("serialize dht error: serialize bucket index err, %s", err)
			}
			if err := serialization.WriteVarUint(w, uint64(len(buck.entries))); err != nil {
				return fmt.Errorf("serialize dht error: serialize node length err, %s", err)
			}
			for _, node := range buck.entries {
				if _, err := w.Write(node.ID[:]); err != nil {
					return fmt.Errorf("serialize dht error: serialize node ID err, %s, bucket index: %d", err, index)
				}
				if err := serialization.WriteString(w, node.IP); err != nil {
					return fmt.Errorf("serialize dht error: serialize node IP err, %s, bucket index: %d",
						err, index)
				}
				if err := serialization.WriteUint16(w, node.UDPPort); err != nil {
					return fmt.Errorf("serialize dht error: serialize node udp port err, %s, bucket index: %d",
						err, index)
				}
				if err := serialization.WriteUint16(w, node.TCPPort); err != nil {
					return fmt.Errorf("serialize dht error: serialize node tcp port err, %s, bucket index: %d",
						err, index)
				}
			}
		}
	}
	return nil
}

func (this *routingTable) deserialize(r io.Reader) error {
	if _, err := r.Read(this.id[:]); err != nil {
		return fmt.Errorf("serialize dht error: serialize ID err, %s", err)
	}
	notEmptyNum, err := serialization.ReadVarUint(r, 0)
	if err != nil {
		return fmt.Errorf("deserialize dht error: serialize not empty bucket num err, %s", err)
	}
	for i := 0; i < int(notEmptyNum); i++ {
		bucketIndex, err := serialization.ReadVarUint(r, 0)
		if err != nil {
			return fmt.Errorf("deserialize dht error: deserialize bucket index err, %s", err)
		}
		nodeNum, err := serialization.ReadVarUint(r, 0)
		if err != nil {
			return fmt.Errorf("deserialize dht error: deserialize node length err, %s", err)
		}
		if nodeNum > 0 {
			for j := 0; j < int(nodeNum); j++ {
				node := new(types.Node)
				if _, err := io.ReadFull(r, node.ID[:]); err != nil {
					return fmt.Errorf("deserialize dht error: deserialize node ID err, %s, bucket index: %d",
						err, bucketIndex)
				}
				if node.IP, err = serialization.ReadString(r); err != nil {
					return fmt.Errorf("deserialize dht error: deserialize node IP err, %s, bucket index: %d",
						err, bucketIndex)
				}
				if node.UDPPort, err = serialization.ReadUint16(r); err != nil {
					return fmt.Errorf("deserialize dht error: deserialize node udp port err, %s, bucket index: %d",
						err, bucketIndex)
				}
				if node.TCPPort, err = serialization.ReadUint16(r); err != nil {
					return fmt.Errorf("deserialize dht error: deserialize node tcp port err, %s, bucket index: %d",
						err, bucketIndex)
				}
				this.addNode(node, int(bucketIndex))
			}
		}
	}
	return nil
}

// table of leading zero counts for bytes [0..255]
var lzcount = [256]int{
	8, 7, 6, 6, 5, 5, 5, 5,
	4, 4, 4, 4, 4, 4, 4, 4,
	3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3,
	2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
}

// logdist returns the logarithmic distance between a and b, log2(a ^ b).
func logdist(a, b types.NodeID) int {
	lz := 0
	for i := range a {
		x := a[i] ^ b[i]
		if x == 0 {
			lz += 8
		} else {
			lz += lzcount[x]
			break
		}
	}
	return len(a)*8 - lz
}
