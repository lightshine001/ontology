// Code generated by protoc-gen-go. DO NOT EDIT.
// source: transaction.proto

package netpb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Trn struct {
	Transaction          []byte   `protobuf:"bytes,1,opt,name=Transaction,proto3" json:"Transaction,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Trn) Reset()         { *m = Trn{} }
func (m *Trn) String() string { return proto.CompactTextString(m) }
func (*Trn) ProtoMessage()    {}
func (*Trn) Descriptor() ([]byte, []int) {
	return fileDescriptor_transaction_797c255c61a76713, []int{0}
}
func (m *Trn) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Trn.Unmarshal(m, b)
}
func (m *Trn) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Trn.Marshal(b, m, deterministic)
}
func (dst *Trn) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Trn.Merge(dst, src)
}
func (m *Trn) XXX_Size() int {
	return xxx_messageInfo_Trn.Size(m)
}
func (m *Trn) XXX_DiscardUnknown() {
	xxx_messageInfo_Trn.DiscardUnknown(m)
}

var xxx_messageInfo_Trn proto.InternalMessageInfo

func (m *Trn) GetTransaction() []byte {
	if m != nil {
		return m.Transaction
	}
	return nil
}

func init() {
	proto.RegisterType((*Trn)(nil), "netpb.Trn")
}

func init() { proto.RegisterFile("transaction.proto", fileDescriptor_transaction_797c255c61a76713) }

var fileDescriptor_transaction_797c255c61a76713 = []byte{
	// 78 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x2c, 0x29, 0x4a, 0xcc,
	0x2b, 0x4e, 0x4c, 0x2e, 0xc9, 0xcc, 0xcf, 0xd3, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0xcd,
	0x4b, 0x2d, 0x29, 0x48, 0x52, 0x52, 0xe7, 0x62, 0x0e, 0x29, 0xca, 0x13, 0x52, 0xe0, 0xe2, 0x0e,
	0x41, 0x28, 0x91, 0x60, 0x54, 0x60, 0xd4, 0xe0, 0x09, 0x42, 0x16, 0x4a, 0x62, 0x03, 0x6b, 0x33,
	0x06, 0x04, 0x00, 0x00, 0xff, 0xff, 0xef, 0xf2, 0xa0, 0x32, 0x4b, 0x00, 0x00, 0x00,
}
