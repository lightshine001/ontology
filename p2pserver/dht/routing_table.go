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
	"crypto/sha256"
	//"fmt"
	"sync"

	"github.com/ontio/ontology/common"
)

type bucket struct {
	entries []*Node
}

type routingTable struct {
	mu      sync.Mutex
	id      NodeID
	buckets []*bucket
}

func (this *routingTable) init(id NodeID) {
	this.buckets = make([]*bucket, BUCKET_NUM)
	this.id = id
}

func (this *routingTable) locateBucket(id NodeID) (int, *bucket) {
	id1 := sha256.Sum256(this.id[:])
	id2 := sha256.Sum256(id[:])
	dist := logdist(id1, id2)
	return dist, this.buckets[dist-1]
}

func (this *routingTable) AddNode(node *Node) bool {
	this.mu.Lock()
	defer this.mu.Unlock()

	_, bucket := this.locateBucket(node.ID)

	for i, entry := range bucket.entries {
		if entry.ID == node.ID {
			copy(bucket.entries[1:], bucket.entries[:i])
			bucket.entries[0] = node
			return true
		}
	}

	// Todo: if the bucket is full, use LRU to replace
	if len(bucket.entries) >= BUCKET_SIZE {
		// bucket is full
		return false
	}

	copy(bucket.entries[1:], bucket.entries[:])
	bucket.entries[0] = node
	return true
}

func (this *routingTable) RemoveNode(id NodeID) {
	this.mu.Lock()
	defer this.mu.Unlock()
	_, bucket := this.locateBucket(id)

	for i, entry := range bucket.entries {
		if entry.ID == id {
			copy(bucket.entries[:i], bucket.entries[i+1:])
			return
		}
	}
}

func (this *routingTable) GetClosestNodes(num int, targetID NodeID) []*Node {
	this.mu.Lock()
	defer this.mu.Unlock()
	closestList := make([]*Node, 0, num)

	index, _ := this.locateBucket(targetID)
	buckets := []int{index}
	i := index - 1
	j := index + 1

	for len(buckets) < BUCKET_NUM {
		if j < BUCKET_NUM {
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

func (this *routingTable) GetTotalNodeNumInBukcet(bucket int) int {
	this.mu.Lock()
	defer this.mu.Unlock()
	b := this.buckets[bucket]
	if b == nil {
		return 0
	}

	return len(b.entries)
}

func (this *routingTable) GetDistance(id1 NodeID, id2 NodeID) int {
	sha1 := sha256.Sum256(id1[:])
	sha2 := sha256.Sum256(id2[:])
	dist := logdist(sha1, sha2)
	return dist
}

func (this *routingTable) totalNodes() int {
	this.mu.Lock()
	defer this.mu.Unlock()
	var num int
	for _, bucket := range this.buckets {
		num += len(bucket.entries)
	}
	return num
}

func (this *routingTable) isNodeInBucket(id NodeID, bucket int) bool {
	this.mu.Lock()
	defer this.mu.Unlock()

	b := this.buckets[bucket]
	if b == nil {
		return false
	}

	for _, entry := range b.entries {
		if entry.ID == id {
			return true
		}
	}
	return false
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
func logdist(a, b common.Uint256) int {
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
