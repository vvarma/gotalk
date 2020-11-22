// Code generated by protoc-gen-go. DO NOT EDIT.
// source: pkg/gotalk/admin.proto

package gotalk

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type AdminMessage struct {
	Identity             *AdminMessage_Identity    `protobuf:"bytes,1,opt,name=identity,proto3" json:"identity,omitempty"`
	CurrentHost          *AdminMessage_CurrentHost `protobuf:"bytes,2,opt,name=current_host,json=currentHost,proto3" json:"current_host,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                  `json:"-"`
	XXX_unrecognized     []byte                    `json:"-"`
	XXX_sizecache        int32                     `json:"-"`
}

func (m *AdminMessage) Reset()         { *m = AdminMessage{} }
func (m *AdminMessage) String() string { return proto.CompactTextString(m) }
func (*AdminMessage) ProtoMessage()    {}
func (*AdminMessage) Descriptor() ([]byte, []int) {
	return fileDescriptor_b1bfdbf9c0991cfe, []int{0}
}

func (m *AdminMessage) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AdminMessage.Unmarshal(m, b)
}
func (m *AdminMessage) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AdminMessage.Marshal(b, m, deterministic)
}
func (m *AdminMessage) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AdminMessage.Merge(m, src)
}
func (m *AdminMessage) XXX_Size() int {
	return xxx_messageInfo_AdminMessage.Size(m)
}
func (m *AdminMessage) XXX_DiscardUnknown() {
	xxx_messageInfo_AdminMessage.DiscardUnknown(m)
}

var xxx_messageInfo_AdminMessage proto.InternalMessageInfo

func (m *AdminMessage) GetIdentity() *AdminMessage_Identity {
	if m != nil {
		return m.Identity
	}
	return nil
}

func (m *AdminMessage) GetCurrentHost() *AdminMessage_CurrentHost {
	if m != nil {
		return m.CurrentHost
	}
	return nil
}

type AdminMessage_Identity struct {
	Username             string   `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AdminMessage_Identity) Reset()         { *m = AdminMessage_Identity{} }
func (m *AdminMessage_Identity) String() string { return proto.CompactTextString(m) }
func (*AdminMessage_Identity) ProtoMessage()    {}
func (*AdminMessage_Identity) Descriptor() ([]byte, []int) {
	return fileDescriptor_b1bfdbf9c0991cfe, []int{0, 0}
}

func (m *AdminMessage_Identity) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AdminMessage_Identity.Unmarshal(m, b)
}
func (m *AdminMessage_Identity) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AdminMessage_Identity.Marshal(b, m, deterministic)
}
func (m *AdminMessage_Identity) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AdminMessage_Identity.Merge(m, src)
}
func (m *AdminMessage_Identity) XXX_Size() int {
	return xxx_messageInfo_AdminMessage_Identity.Size(m)
}
func (m *AdminMessage_Identity) XXX_DiscardUnknown() {
	xxx_messageInfo_AdminMessage_Identity.DiscardUnknown(m)
}

var xxx_messageInfo_AdminMessage_Identity proto.InternalMessageInfo

func (m *AdminMessage_Identity) GetUsername() string {
	if m != nil {
		return m.Username
	}
	return ""
}

type AdminMessage_CurrentHost struct {
	PeerId               string   `protobuf:"bytes,1,opt,name=peerId,proto3" json:"peerId,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AdminMessage_CurrentHost) Reset()         { *m = AdminMessage_CurrentHost{} }
func (m *AdminMessage_CurrentHost) String() string { return proto.CompactTextString(m) }
func (*AdminMessage_CurrentHost) ProtoMessage()    {}
func (*AdminMessage_CurrentHost) Descriptor() ([]byte, []int) {
	return fileDescriptor_b1bfdbf9c0991cfe, []int{0, 1}
}

func (m *AdminMessage_CurrentHost) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AdminMessage_CurrentHost.Unmarshal(m, b)
}
func (m *AdminMessage_CurrentHost) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AdminMessage_CurrentHost.Marshal(b, m, deterministic)
}
func (m *AdminMessage_CurrentHost) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AdminMessage_CurrentHost.Merge(m, src)
}
func (m *AdminMessage_CurrentHost) XXX_Size() int {
	return xxx_messageInfo_AdminMessage_CurrentHost.Size(m)
}
func (m *AdminMessage_CurrentHost) XXX_DiscardUnknown() {
	xxx_messageInfo_AdminMessage_CurrentHost.DiscardUnknown(m)
}

var xxx_messageInfo_AdminMessage_CurrentHost proto.InternalMessageInfo

func (m *AdminMessage_CurrentHost) GetPeerId() string {
	if m != nil {
		return m.PeerId
	}
	return ""
}

func init() {
	proto.RegisterType((*AdminMessage)(nil), "gotalk.AdminMessage")
	proto.RegisterType((*AdminMessage_Identity)(nil), "gotalk.AdminMessage.Identity")
	proto.RegisterType((*AdminMessage_CurrentHost)(nil), "gotalk.AdminMessage.CurrentHost")
}

func init() {
	proto.RegisterFile("pkg/gotalk/admin.proto", fileDescriptor_b1bfdbf9c0991cfe)
}

var fileDescriptor_b1bfdbf9c0991cfe = []byte{
	// 184 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x2b, 0xc8, 0x4e, 0xd7,
	0x4f, 0xcf, 0x2f, 0x49, 0xcc, 0xc9, 0xd6, 0x4f, 0x4c, 0xc9, 0xcd, 0xcc, 0xd3, 0x2b, 0x28, 0xca,
	0x2f, 0xc9, 0x17, 0x62, 0x83, 0x88, 0x29, 0xdd, 0x65, 0xe4, 0xe2, 0x71, 0x04, 0x89, 0xfb, 0xa6,
	0x16, 0x17, 0x27, 0xa6, 0xa7, 0x0a, 0x59, 0x72, 0x71, 0x64, 0xa6, 0xa4, 0xe6, 0x95, 0x64, 0x96,
	0x54, 0x4a, 0x30, 0x2a, 0x30, 0x6a, 0x70, 0x1b, 0xc9, 0xea, 0x41, 0xd4, 0xea, 0x21, 0xab, 0xd3,
	0xf3, 0x84, 0x2a, 0x0a, 0x82, 0x2b, 0x17, 0x72, 0xe6, 0xe2, 0x49, 0x2e, 0x2d, 0x2a, 0x4a, 0xcd,
	0x2b, 0x89, 0xcf, 0xc8, 0x2f, 0x2e, 0x91, 0x60, 0x02, 0x6b, 0x57, 0xc0, 0xaa, 0xdd, 0x19, 0xa2,
	0xd0, 0x23, 0xbf, 0xb8, 0x24, 0x88, 0x3b, 0x19, 0xc1, 0x91, 0x52, 0xe3, 0xe2, 0x80, 0x19, 0x2d,
	0x24, 0xc5, 0xc5, 0x51, 0x5a, 0x9c, 0x5a, 0x94, 0x97, 0x98, 0x9b, 0x0a, 0x76, 0x0b, 0x67, 0x10,
	0x9c, 0x2f, 0xa5, 0xca, 0xc5, 0x8d, 0x64, 0x86, 0x90, 0x18, 0x17, 0x5b, 0x41, 0x6a, 0x6a, 0x91,
	0x67, 0x0a, 0x54, 0x21, 0x94, 0x97, 0xc4, 0x06, 0xf6, 0xae, 0x31, 0x20, 0x00, 0x00, 0xff, 0xff,
	0x12, 0x73, 0x5c, 0xbe, 0x08, 0x01, 0x00, 0x00,
}