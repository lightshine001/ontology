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
	"github.com/ontio/ontology/p2pserver/dht/types"
)

type FindNodePayload struct {
	FromID   types.NodeID
	TargetID types.NodeID
}

type FindNode struct {
	P FindNodePayload
}

//Serialize message payload
func (this FindNode) Serialization() ([]byte, error) {
	p := bytes.NewBuffer([]byte{})
	err := binary.Write(p, binary.LittleEndian, this.P)
	if err != nil {
		log.Error("failed to write DHT findnode payload failed")
		return nil, err
	}

	return p.Bytes(), nil
}

//Deserialize message payload
func (this *FindNode) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)

	err := binary.Read(buf, binary.LittleEndian, &this.P)
	if err != nil {
		log.Error("Parse DHT findnode message error", err)
		return errors.New("Parse DHT findnode payload message error")
	}

	return err
}
