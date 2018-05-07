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

package types

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/dht"
)

type NeighborsPayload struct {
	FromID dht.NodeID
	Nodes  []dht.Node
}

type Neighbors struct {
	Hdr MsgHdr
	P   NeighborsPayload
}

//Check whether header is correct
func (this Neighbors) Verify(buf []byte) error {
	err := this.Hdr.Verify(buf)
	return err
}

//Serialize message payload
func (this Neighbors) Serialization() ([]byte, error) {
	p := bytes.NewBuffer([]byte{})
	err := binary.Write(p, binary.LittleEndian, this.P)
	if err != nil {
		log.Error("failed to write DHT neighbors payload failed")
		return nil, err
	}

	checkSumBuf := CheckSum(p.Bytes())
	this.Hdr.Init("neighbors", checkSumBuf, uint32(len(p.Bytes())))

	hdrBuf, err := this.Hdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	data := append(buf.Bytes(), p.Bytes()...)
	return data, nil
}

//Deserialize message payload
func (this *Neighbors) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(this.Hdr))
	if err != nil {
		return err
	}

	err = binary.Read(buf, binary.LittleEndian, &this.P)
	if err != nil {
		log.Error("Parse DHT neighbors message error", err)
		return errors.New("Parse DHT neighbors payload message error")
	}

	return err
}
