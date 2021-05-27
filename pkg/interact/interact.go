// Copyright 2021 bilibili-base
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package interact

import (
	"encoding/json"

	"github.com/golang/protobuf/proto"
)

// Protocol defines the protocol of request
type Protocol string

// defines a set of known protocols
const (
	ProtocolHTTP Protocol = "HTTP"
	ProtocolGRPC Protocol = "GRPC"
)

// Message defines a generic message interface
type Message interface {
	proto.Message
	json.Marshaler
	json.Unmarshaler
	proto.Marshaler
	Bytes() []byte
}

// Request defines the request structure
type Request struct {
	Protocol Protocol          `json:"protocol"`
	Method   string            `json:"method"`
	Host     string            `json:"host"`
	Path     string            `json:"path"`
	Header   map[string]string `json:"header"`
	Body     Message           `json:"body"`
}

// Response defines the response structure
type Response struct {
	Code    uint32            `json:"code"`
	Header  map[string]string `json:"header"`
	Body    Message           `json:"body"`
	Trailer map[string]string `json:"trailer"`
}

// NewDefaultResponse is used to create default response
func NewDefaultResponse(request *Request) *Response {
	var code uint32
	switch request.Protocol {
	case ProtocolGRPC:
		code = 0
	case ProtocolHTTP:
		code = 1
	}
	return &Response{
		Code:    code,
		Header:  map[string]string{},
		Trailer: map[string]string{},
		Body:    NewBytesMessage(nil),
	}
}

// BytesMessage is the simple implement of Message
type BytesMessage struct {
	data []byte
}

// NewBytesMessage is used to init BytesMessage
func NewBytesMessage(data []byte) Message {
	return &BytesMessage{
		data: data,
	}
}

// Reset implements the proto.Message interface
func (b *BytesMessage) Reset() {}

// String implements the proto.Message interface
func (b *BytesMessage) String() string {
	return string(b.data)
}

// ProtoMessage implements the proto.Message interface
func (b *BytesMessage) ProtoMessage() {}

// Marshal implements the proto.Marshaler interface
func (b *BytesMessage) Marshal() ([]byte, error) {
	return b.data, nil
}

// UnmarshalJSON implements the json.UnmarshalJSON interface
func (b *BytesMessage) UnmarshalJSON(data []byte) error {
	b.data = data
	return nil
}

// MarshalJSON implements the json.Marshaler interface
func (b *BytesMessage) MarshalJSON() ([]byte, error) {
	if len(b.data) == 0 {
		return []byte(`null`), nil
	}
	return b.data, nil
}

// Bytes is used to return native data
func (b *BytesMessage) Bytes() []byte {
	return b.data
}
