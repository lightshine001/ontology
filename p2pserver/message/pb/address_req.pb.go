// Code generated by protoc-gen-go. DO NOT EDIT.
// source: address_req.proto

package netpb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type AddrReq struct {
}

func (m *AddrReq) Reset()                    { *m = AddrReq{} }
func (m *AddrReq) String() string            { return proto.CompactTextString(m) }
func (*AddrReq) ProtoMessage()               {}
func (*AddrReq) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{0} }

func init() {
	proto.RegisterType((*AddrReq)(nil), "netpb.AddrReq")
}

func init() { proto.RegisterFile("address_req.proto", fileDescriptor1) }

var fileDescriptor1 = []byte{
	// 66 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x4c, 0x4c, 0x49, 0x29,
	0x4a, 0x2d, 0x2e, 0x8e, 0x2f, 0x4a, 0x2d, 0xd4, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0xcd,
	0x4b, 0x2d, 0x29, 0x48, 0x52, 0xe2, 0xe4, 0x62, 0x77, 0x4c, 0x49, 0x29, 0x0a, 0x4a, 0x2d, 0x4c,
	0x62, 0x03, 0x4b, 0x18, 0x03, 0x02, 0x00, 0x00, 0xff, 0xff, 0x68, 0x91, 0x99, 0x1e, 0x2d, 0x00,
	0x00, 0x00,
}
