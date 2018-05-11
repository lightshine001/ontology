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
	"github.com/ontio/ontology/common/log"
<<<<<<< HEAD
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/p2pserver/dht/types"
	"github.com/ontio/ontology/p2pserver/common"
)

=======
	"github.com/ontio/ontology/p2pserver/dht/types"
)

type DHTPongPayload struct {
	Version  uint16
	FromID   types.NodeID
	SrcAddr  [16]byte
	SrcPort  uint16
	DestAddr [16]byte
	DestPort uint16
}
>>>>>>> add msg pack for ping/pong, findnode/neighbors

type DHTPong struct {
	Version      uint16
	FromID       types.NodeID
	SrcEndPoint  EndPoint
	DestEndPoint EndPoint
}

func (this *DHTPong) CmdType() string {
	return common.DHT_PONG
}

//Serialize message
func (this DHTPong) Serialization() ([]byte, error) {
	p := bytes.NewBuffer([]byte{})
	err := serialization.WriteUint16(p, this.Version)
	if err != nil {
		log.Errorf("failed to serialize version %v. version %x",
			err, this.Version)
		return nil, err
	}

	err = serialization.WriteVarBytes(p, this.FromID[:])
	if err != nil {
		log.Errorf("failed to serialize node id %v. ID %x",
			err, this.FromID)
		return nil, err
	}

	err = serialization.WriteVarBytes(p, this.SrcEndPoint.Addr[:])
	if err != nil {
		log.Errorf("failed to serialize src addr %v. addr %s",
			err, this.SrcEndPoint.Addr)
		return nil, err
	}

	err = serialization.WriteUint16(p, this.SrcEndPoint.UDPPort)
	if err != nil {
		log.Errorf("failed to serialize src udp port %v. UDPPort %d",
			err, this.SrcEndPoint.UDPPort)
		return nil, err
	}

	err = serialization.WriteUint16(p, this.SrcEndPoint.TCPPort)
	if err != nil {
		log.Errorf("failed to serialize src tcp port %v. TCPPort %d",
			err, this.SrcEndPoint.TCPPort)
		return nil, err
	}

	err = serialization.WriteVarBytes(p, this.DestEndPoint.Addr[:])
	if err != nil {
		log.Errorf("failed to serialize dest addr %v. addr %s",
			err, this.SrcEndPoint.Addr)
		return nil, err
	}

	err = serialization.WriteUint16(p, this.DestEndPoint.UDPPort)
	if err != nil {
		log.Errorf("failed to serialize dest udp port %v. UDPPort %d",
			err, this.SrcEndPoint.UDPPort)
		return nil, err
	}

	err = serialization.WriteUint16(p, this.DestEndPoint.TCPPort)
	if err != nil {
		log.Errorf("failed to serialize dest tcp port %v. TCPPort %d",
			err, this.SrcEndPoint.TCPPort)
		return nil, err
	}

	return p.Bytes(), nil
}

//Deserialize message
func (this *DHTPong) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	var err error
	this.Version, err = serialization.ReadUint16(buf)
	if err != nil {
		log.Errorf("failed to deserialize pong version %v", err)
		return err
	}

	id, err := serialization.ReadVarBytes(buf)
	if err != nil {
		log.Errorf("failed to deserialize pong  id %v", err)
		return err
	}
	copy(this.FromID[:], id)

	addr, err := serialization.ReadVarBytes(buf)
	if err != nil {
		log.Errorf("failed to deserialize node ip %v", err)
		return err
	}
	copy(this.SrcEndPoint.Addr[:], addr)

	this.SrcEndPoint.UDPPort, err = serialization.ReadUint16(buf)
	if err != nil {
		log.Errorf("failed to deserialize pong src udp port %v", err)
		return err
	}

	this.SrcEndPoint.TCPPort, err = serialization.ReadUint16(buf)
	if err != nil {
		log.Errorf("failed to deserialize pong src tcp port %v", err)
		return err
	}

	addr, err = serialization.ReadVarBytes(buf)
	if err != nil {
		log.Errorf("failed to deserialize pong dest  address  %v", err)
		return err
	}
	copy(this.DestEndPoint.Addr[:], addr)

	this.DestEndPoint.UDPPort, err = serialization.ReadUint16(buf)
	if err != nil {
		log.Errorf("failed to deserialize pong dest udp port %v", err)
		return err
	}

	this.DestEndPoint.TCPPort, err = serialization.ReadUint16(buf)
	if err != nil {
		log.Errorf("failed to deserialize ping dest tcp port %v", err)
		return err
	}

	return err
}
