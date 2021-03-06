// Code generated by protoc-gen-go.
// source: source/data.proto
// DO NOT EDIT!

package wire

import proto "code.google.com/p/goprotobuf/proto"
import json "encoding/json"
import math "math"

// Reference proto, json, and math imports to suppress error if they are not otherwise used.
var _ = proto.Marshal
var _ = &json.SyntaxError{}
var _ = math.Inf

type Data struct {
	Hash             []byte  `protobuf:"bytes,1,req,name=hash" json:"hash,omitempty"`
	Length           *uint64 `protobuf:"varint,2,req,name=length" json:"length,omitempty"`
	Key              []byte  `protobuf:"bytes,3,req,name=key" json:"key,omitempty"`
	Type             *string `protobuf:"bytes,4,req,name=type" json:"type,omitempty"`
	Name             *string `protobuf:"bytes,5,opt,name=name" json:"name,omitempty"`
	File             *string `protobuf:"bytes,6,opt,name=file" json:"file,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *Data) Reset()         { *m = Data{} }
func (m *Data) String() string { return proto.CompactTextString(m) }
func (*Data) ProtoMessage()    {}

func (m *Data) GetHash() []byte {
	if m != nil {
		return m.Hash
	}
	return nil
}

func (m *Data) GetLength() uint64 {
	if m != nil && m.Length != nil {
		return *m.Length
	}
	return 0
}

func (m *Data) GetKey() []byte {
	if m != nil {
		return m.Key
	}
	return nil
}

func (m *Data) GetType() string {
	if m != nil && m.Type != nil {
		return *m.Type
	}
	return ""
}

func (m *Data) GetName() string {
	if m != nil && m.Name != nil {
		return *m.Name
	}
	return ""
}

func (m *Data) GetFile() string {
	if m != nil && m.File != nil {
		return *m.File
	}
	return ""
}

type Mail struct {
	Components       []*Mail_Component `protobuf:"bytes,1,rep,name=components" json:"components,omitempty"`
	Name             *string           `protobuf:"bytes,2,opt,name=name" json:"name,omitempty"`
	XXX_unrecognized []byte            `json:"-"`
}

func (m *Mail) Reset()         { *m = Mail{} }
func (m *Mail) String() string { return proto.CompactTextString(m) }
func (*Mail) ProtoMessage()    {}

func (m *Mail) GetComponents() []*Mail_Component {
	if m != nil {
		return m.Components
	}
	return nil
}

func (m *Mail) GetName() string {
	if m != nil && m.Name != nil {
		return *m.Name
	}
	return ""
}

type Mail_Component struct {
	Type             *string `protobuf:"bytes,1,req,name=type" json:"type,omitempty"`
	Data             []byte  `protobuf:"bytes,2,req,name=data" json:"data,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *Mail_Component) Reset()         { *m = Mail_Component{} }
func (m *Mail_Component) String() string { return proto.CompactTextString(m) }
func (*Mail_Component) ProtoMessage()    {}

func (m *Mail_Component) GetType() string {
	if m != nil && m.Type != nil {
		return *m.Type
	}
	return ""
}

func (m *Mail_Component) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

type Error struct {
	Code             *uint32 `protobuf:"varint,1,req,name=code" json:"code,omitempty"`
	Description      *string `protobuf:"bytes,2,opt,name=description" json:"description,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *Error) Reset()         { *m = Error{} }
func (m *Error) String() string { return proto.CompactTextString(m) }
func (*Error) ProtoMessage()    {}

func (m *Error) GetCode() uint32 {
	if m != nil && m.Code != nil {
		return *m.Code
	}
	return 0
}

func (m *Error) GetDescription() string {
	if m != nil && m.Description != nil {
		return *m.Description
	}
	return ""
}

func init() {
}
