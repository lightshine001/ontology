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
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/golang/protobuf/proto"
	com "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/message/pb"
)

var LastInvHash com.Uint256

type NetMessage struct {
	messageName string
	content     proto.Message
}

func NewNetMessage(messageName string, message proto.Message) *NetMessage {
	netMessage := &NetMessage{
		messageName: messageName,
		content:     message,
	}
	return netMessage
}

func (this *NetMessage) MessageName() string {
	return this.messageName
}

func (this *NetMessage) Content() proto.Message {
	return this.content
}

type Message interface {
	Serialization() ([]byte, error)
	Deserialization([]byte) error
	CmdType() string
}

//MsgPayload in link channel
type MsgPayload struct {
	Id          uint64      //peer ID
	Addr        string      //link address
	PayloadSize uint32      //payload size
	Payload     *NetMessage //msg payload
}

type messageHeader struct {
	Magic    uint32
	CMD      [common.MSG_CMD_LEN]byte // The message type
	Length   uint32
	Checksum [common.CHECKSUM_LEN]byte
}

func readMessageHeader(reader io.Reader) (messageHeader, error) {
	msgh := messageHeader{}
	err := binary.Read(reader, binary.LittleEndian, &msgh)
	return msgh, err
}

func writeMessageHeader(writer io.Writer, msgh messageHeader) error {
	return binary.Write(writer, binary.LittleEndian, msgh)
}

func newMessageHeader(cmd string, length uint32, checksum [common.CHECKSUM_LEN]byte) messageHeader {
	msgh := messageHeader{}
	msgh.Magic = config.DefConfig.P2PNode.NetworkMagic
	copy(msgh.CMD[:], cmd)
	msgh.Checksum = checksum
	msgh.Length = length
	return msgh
}

func WriteMessage(writer io.Writer, msg *NetMessage) error {
	data, err := proto.Marshal(msg.Content())
	if err != nil {
		return err
	}

	checksum := CheckSum(data)

	hdr := newMessageHeader(msg.MessageName(), uint32(len(data)), checksum)

	err = writeMessageHeader(writer, hdr)
	if err != nil {
		return err
	}

	_, err = writer.Write(data)
	return err
}

func ReadMessage(reader io.Reader) (*NetMessage, error) {
	hdr, err := readMessageHeader(reader)
	if err != nil {
		return nil, err
	}

	magic := config.DefConfig.P2PNode.NetworkMagic
	if hdr.Magic != magic {
		return nil, fmt.Errorf("unmatched magic number %d, expected %d", hdr.Magic, magic)
	}

	if int(hdr.Length) > common.MAX_PAYLOAD_LEN {
		return nil, fmt.Errorf("msg payload length:%d exceed max payload size: %d",
			hdr.Length, common.MAX_PAYLOAD_LEN)
	}

	buf := make([]byte, hdr.Length)
	_, err = io.ReadFull(reader, buf)
	if err != nil {
		return nil, err
	}

	checksum := CheckSum(buf)
	if checksum != hdr.Checksum {
		return nil, fmt.Errorf("message checksum mismatch: %x != %x ", hdr.Checksum, checksum)
	}

	cmdType := string(bytes.TrimRight(hdr.CMD[:], string(0)))

	msg, err := MakeEmptyMessage(cmdType)
	if err != nil {
		return nil, err
	}

	if err := proto.Unmarshal(buf, msg); err != nil {
		return nil, err
	}

	netMessage := NewNetMessage(cmdType, msg)
	return netMessage, nil
}

func MakeEmptyMessage(cmdType string) (proto.Message, error) {
	switch cmdType {
	case common.PING_TYPE:
		return &netpb.Ping{}, nil
	case common.VERSION_TYPE:
		return &netpb.Version{}, nil
	case common.VERACK_TYPE:
		return &netpb.VerAck{}, nil
	case common.ADDR_TYPE:
		return &netpb.Addrs{}, nil
	case common.GetADDR_TYPE:
		return &netpb.AddrReq{}, nil
	case common.PONG_TYPE:
		return &netpb.Pong{}, nil
	case common.GET_HEADERS_TYPE:
		return &netpb.HeadersReq{}, nil
	case common.HEADERS_TYPE:
		return &netpb.BlkHeader{}, nil
	case common.INV_TYPE:
		return &netpb.Inv{}, nil
	case common.GET_DATA_TYPE:
		return &netpb.DataReq{}, nil
	case common.BLOCK_TYPE:
		return &netpb.Block{}, nil
	case common.TX_TYPE:
		return &netpb.Trn{}, nil
	case common.CONSENSUS_TYPE:
		return &netpb.Consensus{}, nil
	case common.NOT_FOUND_TYPE:
		return &netpb.NotFound{}, nil
	case common.DISCONNECT_TYPE:
		return &netpb.Disconnected{}, nil
	case common.GET_BLOCKS_TYPE:
		return &netpb.BlocksReq{}, nil
	case common.DHT_PING:
		return &netpb.DHTPing{}, nil
	case common.DHT_PONG:
		return &netpb.DHTPong{}, nil
	case common.DHT_FIND_NODE:
		return &netpb.FindNode{}, nil
	case common.DHT_NEIGHBORS:
		return &netpb.Neighbors{}, nil
	default:
		return nil, errors.New("unsupported cmd type:" + cmdType)
	}

}

//caculate checksum value
func CheckSum(p []byte) [common.CHECKSUM_LEN]byte {
	var checksum [common.CHECKSUM_LEN]byte
	t := sha256.Sum256(p)
	s := sha256.Sum256(t[:])

	copy(checksum[:], s[:])

	return checksum
}
