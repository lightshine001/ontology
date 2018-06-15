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
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/p2pserver/dht/types"
)

type DHTPongPayload struct {
	Version      uint16
	FromID       types.NodeID
	SrcEndPoint  EndPoint
	DestEndPoint EndPoint
}

type DHTPong struct {
	P DHTPongPayload
}

//Serialize message payload
func (this DHTPong) Serialization() ([]byte, error) {
	p := bytes.NewBuffer([]byte{})
	payload := this.P
	err := serialization.WriteUint16(p, payload.Version)
	if err != nil {
		log.Errorf("failed to serialize version %v. version %x",
			err, payload.Version)
		return nil, err
	}

	err = serialization.WriteVarBytes(p, payload.FromID[:])
	if err != nil {
		log.Errorf("failed to serialize node id %v. ID %x",
			err, payload.FromID)
		return nil, err
	}

	err = serialization.WriteVarBytes(p, payload.SrcEndPoint.Addr[:])
	if err != nil {
		log.Errorf("failed to serialize src addr %v. addr %s",
			err, payload.SrcEndPoint.Addr)
		return nil, err
	}

	err = serialization.WriteUint16(p, payload.SrcEndPoint.UDPPort)
	if err != nil {
		log.Errorf("failed to serialize src udp port %v. UDPPort %d",
			err, payload.SrcEndPoint.UDPPort)
		return nil, err
	}

	err = serialization.WriteUint16(p, payload.SrcEndPoint.TCPPort)
	if err != nil {
		log.Errorf("failed to serialize src tcp port %v. TCPPort %d",
			err, payload.SrcEndPoint.TCPPort)
		return nil, err
	}

	err = serialization.WriteVarBytes(p, payload.DestEndPoint.Addr[:])
	if err != nil {
		log.Errorf("failed to serialize dest addr %v. addr %s",
			err, payload.SrcEndPoint.Addr)
		return nil, err
	}

	err = serialization.WriteUint16(p, payload.DestEndPoint.UDPPort)
	if err != nil {
		log.Errorf("failed to serialize dest udp port %v. UDPPort %d",
			err, payload.SrcEndPoint.UDPPort)
		return nil, err
	}

	err = serialization.WriteUint16(p, payload.DestEndPoint.TCPPort)
	if err != nil {
		log.Errorf("failed to serialize dest tcp port %v. TCPPort %d",
			err, payload.SrcEndPoint.TCPPort)
		return nil, err
	}

	return p.Bytes(), nil
}

//Deserialize message payload
func (this *DHTPong) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	var err error
	this.P.Version, err = serialization.ReadUint16(buf)
	if err != nil {
		log.Errorf("failed to deserialize pong version %v", err)
		return err
	}

	id, err := serialization.ReadVarBytes(buf)
	if err != nil {
		log.Errorf("failed to deserialize pong  id %v", err)
		return err
	}
	copy(this.P.FromID[:], id)

	addr, err := serialization.ReadVarBytes(buf)
	if err != nil {
		log.Errorf("failed to deserialize node ip %v", err)
		return err
	}
	copy(this.P.SrcEndPoint.Addr[:], addr)

	this.P.SrcEndPoint.UDPPort, err = serialization.ReadUint16(buf)
	if err != nil {
		log.Errorf("failed to deserialize pong src udp port %v", err)
		return err
	}

	this.P.SrcEndPoint.TCPPort, err = serialization.ReadUint16(buf)
	if err != nil {
		log.Errorf("failed to deserialize pong src tcp port %v", err)
		return err
	}

	addr, err = serialization.ReadVarBytes(buf)
	if err != nil {
		log.Errorf("failed to deserialize pong dest  address  %v", err)
		return err
	}
	copy(this.P.DestEndPoint.Addr[:], addr)

	this.P.DestEndPoint.UDPPort, err = serialization.ReadUint16(buf)
	if err != nil {
		log.Errorf("failed to deserialize pong dest udp port %v", err)
		return err
	}

	this.P.DestEndPoint.TCPPort, err = serialization.ReadUint16(buf)
	if err != nil {
		log.Errorf("failed to deserialize ping dest tcp port %v", err)
		return err
	}

	return err
}
