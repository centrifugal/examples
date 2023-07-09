// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v3.17.3
// source: proxystream.proto

package proxystreamproto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type CommunicateRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Centrifugo always sends this within the first message upon user subscription request.
	// It's always not set in the following StreamRequest messages from Centrifugo.
	SubscribeRequest *SubscribeRequest `protobuf:"bytes,1,opt,name=subscribe_request,json=subscribeRequest,proto3" json:"subscribe_request,omitempty"`
	// Publication may be set when client publishes to the on-demand stream. If you are using
	// bidirectional stream then Centrifugo assumes publications from client-side are allowed.
	Publication *Publication `protobuf:"bytes,2,opt,name=publication,proto3" json:"publication,omitempty"`
}

func (x *CommunicateRequest) Reset() {
	*x = CommunicateRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proxystream_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CommunicateRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CommunicateRequest) ProtoMessage() {}

func (x *CommunicateRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proxystream_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CommunicateRequest.ProtoReflect.Descriptor instead.
func (*CommunicateRequest) Descriptor() ([]byte, []int) {
	return file_proxystream_proto_rawDescGZIP(), []int{0}
}

func (x *CommunicateRequest) GetSubscribeRequest() *SubscribeRequest {
	if x != nil {
		return x.SubscribeRequest
	}
	return nil
}

func (x *CommunicateRequest) GetPublication() *Publication {
	if x != nil {
		return x.Publication
	}
	return nil
}

type Response struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// SubscribeResponse may optionally be set in the first message from backend to Centrifugo.
	SubscribeResponse *SubscribeResponse `protobuf:"bytes,1,opt,name=subscribe_response,json=subscribeResponse,proto3" json:"subscribe_response,omitempty"`
	// Publication goes to client. Can't be set in the first message from backend to Centrifugo.
	Publication *Publication `protobuf:"bytes,2,opt,name=publication,proto3" json:"publication,omitempty"`
}

func (x *Response) Reset() {
	*x = Response{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proxystream_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Response) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Response) ProtoMessage() {}

func (x *Response) ProtoReflect() protoreflect.Message {
	mi := &file_proxystream_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Response.ProtoReflect.Descriptor instead.
func (*Response) Descriptor() ([]byte, []int) {
	return file_proxystream_proto_rawDescGZIP(), []int{1}
}

func (x *Response) GetSubscribeResponse() *SubscribeResponse {
	if x != nil {
		return x.SubscribeResponse
	}
	return nil
}

func (x *Response) GetPublication() *Publication {
	if x != nil {
		return x.Publication
	}
	return nil
}

type SubscribeResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *SubscribeResponse) Reset() {
	*x = SubscribeResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proxystream_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SubscribeResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubscribeResponse) ProtoMessage() {}

func (x *SubscribeResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proxystream_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SubscribeResponse.ProtoReflect.Descriptor instead.
func (*SubscribeResponse) Descriptor() ([]byte, []int) {
	return file_proxystream_proto_rawDescGZIP(), []int{2}
}

// SubscribeRequest contains information about channel subscription.
type SubscribeRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Client    string `protobuf:"bytes,1,opt,name=client,proto3" json:"client,omitempty"`
	Transport string `protobuf:"bytes,2,opt,name=transport,proto3" json:"transport,omitempty"`
	Protocol  string `protobuf:"bytes,3,opt,name=protocol,proto3" json:"protocol,omitempty"`
	Encoding  string `protobuf:"bytes,4,opt,name=encoding,proto3" json:"encoding,omitempty"`
	Channel   string `protobuf:"bytes,10,opt,name=channel,proto3" json:"channel,omitempty"`
	Data      []byte `protobuf:"bytes,11,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *SubscribeRequest) Reset() {
	*x = SubscribeRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proxystream_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SubscribeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubscribeRequest) ProtoMessage() {}

func (x *SubscribeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proxystream_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SubscribeRequest.ProtoReflect.Descriptor instead.
func (*SubscribeRequest) Descriptor() ([]byte, []int) {
	return file_proxystream_proto_rawDescGZIP(), []int{3}
}

func (x *SubscribeRequest) GetClient() string {
	if x != nil {
		return x.Client
	}
	return ""
}

func (x *SubscribeRequest) GetTransport() string {
	if x != nil {
		return x.Transport
	}
	return ""
}

func (x *SubscribeRequest) GetProtocol() string {
	if x != nil {
		return x.Protocol
	}
	return ""
}

func (x *SubscribeRequest) GetEncoding() string {
	if x != nil {
		return x.Encoding
	}
	return ""
}

func (x *SubscribeRequest) GetChannel() string {
	if x != nil {
		return x.Channel
	}
	return ""
}

func (x *SubscribeRequest) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

// Publication is an event to be sent to a client.
type Publication struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data []byte            `protobuf:"bytes,4,opt,name=data,proto3" json:"data,omitempty"`
	Tags map[string]string `protobuf:"bytes,7,rep,name=tags,proto3" json:"tags,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *Publication) Reset() {
	*x = Publication{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proxystream_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Publication) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Publication) ProtoMessage() {}

func (x *Publication) ProtoReflect() protoreflect.Message {
	mi := &file_proxystream_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Publication.ProtoReflect.Descriptor instead.
func (*Publication) Descriptor() ([]byte, []int) {
	return file_proxystream_proto_rawDescGZIP(), []int{4}
}

func (x *Publication) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *Publication) GetTags() map[string]string {
	if x != nil {
		return x.Tags
	}
	return nil
}

var File_proxystream_proto protoreflect.FileDescriptor

var file_proxystream_proto_rawDesc = []byte{
	0x0a, 0x11, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x22, 0x63, 0x65, 0x6e, 0x74, 0x72, 0x69, 0x66, 0x75, 0x67, 0x61, 0x6c,
	0x2e, 0x63, 0x65, 0x6e, 0x74, 0x72, 0x69, 0x66, 0x75, 0x67, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x78,
	0x79, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x22, 0xca, 0x01, 0x0a, 0x12, 0x43, 0x6f, 0x6d, 0x6d,
	0x75, 0x6e, 0x69, 0x63, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x61,
	0x0a, 0x11, 0x73, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x62, 0x65, 0x5f, 0x72, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x34, 0x2e, 0x63, 0x65, 0x6e, 0x74,
	0x72, 0x69, 0x66, 0x75, 0x67, 0x61, 0x6c, 0x2e, 0x63, 0x65, 0x6e, 0x74, 0x72, 0x69, 0x66, 0x75,
	0x67, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x2e, 0x53,
	0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x62, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x52,
	0x10, 0x73, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x62, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x51, 0x0a, 0x0b, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2f, 0x2e, 0x63, 0x65, 0x6e, 0x74, 0x72, 0x69, 0x66,
	0x75, 0x67, 0x61, 0x6c, 0x2e, 0x63, 0x65, 0x6e, 0x74, 0x72, 0x69, 0x66, 0x75, 0x67, 0x6f, 0x2e,
	0x70, 0x72, 0x6f, 0x78, 0x79, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x2e, 0x50, 0x75, 0x62, 0x6c,
	0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x0b, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x22, 0xc3, 0x01, 0x0a, 0x08, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x64, 0x0a, 0x12, 0x73, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x62, 0x65, 0x5f, 0x72,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x35, 0x2e,
	0x63, 0x65, 0x6e, 0x74, 0x72, 0x69, 0x66, 0x75, 0x67, 0x61, 0x6c, 0x2e, 0x63, 0x65, 0x6e, 0x74,
	0x72, 0x69, 0x66, 0x75, 0x67, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x73, 0x74, 0x72, 0x65,
	0x61, 0x6d, 0x2e, 0x53, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x62, 0x65, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x52, 0x11, 0x73, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x62, 0x65, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x51, 0x0a, 0x0b, 0x70, 0x75, 0x62, 0x6c, 0x69,
	0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2f, 0x2e, 0x63,
	0x65, 0x6e, 0x74, 0x72, 0x69, 0x66, 0x75, 0x67, 0x61, 0x6c, 0x2e, 0x63, 0x65, 0x6e, 0x74, 0x72,
	0x69, 0x66, 0x75, 0x67, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x73, 0x74, 0x72, 0x65, 0x61,
	0x6d, 0x2e, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x0b, 0x70,
	0x75, 0x62, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x13, 0x0a, 0x11, 0x53, 0x75,
	0x62, 0x73, 0x63, 0x72, 0x69, 0x62, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0xae, 0x01, 0x0a, 0x10, 0x53, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x62, 0x65, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x12, 0x1c, 0x0a, 0x09,
	0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x09, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x12, 0x1a, 0x0a, 0x08, 0x65, 0x6e, 0x63, 0x6f, 0x64, 0x69,
	0x6e, 0x67, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x65, 0x6e, 0x63, 0x6f, 0x64, 0x69,
	0x6e, 0x67, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x18, 0x0a, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x12, 0x12, 0x0a, 0x04,
	0x64, 0x61, 0x74, 0x61, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61,
	0x22, 0xc7, 0x01, 0x0a, 0x0b, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04,
	0x64, 0x61, 0x74, 0x61, 0x12, 0x4d, 0x0a, 0x04, 0x74, 0x61, 0x67, 0x73, 0x18, 0x07, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x39, 0x2e, 0x63, 0x65, 0x6e, 0x74, 0x72, 0x69, 0x66, 0x75, 0x67, 0x61, 0x6c,
	0x2e, 0x63, 0x65, 0x6e, 0x74, 0x72, 0x69, 0x66, 0x75, 0x67, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x78,
	0x79, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x2e, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x2e, 0x54, 0x61, 0x67, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x04, 0x74,
	0x61, 0x67, 0x73, 0x1a, 0x37, 0x0a, 0x09, 0x54, 0x61, 0x67, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b,
	0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x4a, 0x04, 0x08, 0x01,
	0x10, 0x02, 0x4a, 0x04, 0x08, 0x02, 0x10, 0x03, 0x4a, 0x04, 0x08, 0x03, 0x10, 0x04, 0x4a, 0x04,
	0x08, 0x05, 0x10, 0x06, 0x4a, 0x04, 0x08, 0x06, 0x10, 0x07, 0x32, 0x81, 0x02, 0x0a, 0x15, 0x43,
	0x65, 0x6e, 0x74, 0x72, 0x69, 0x66, 0x75, 0x67, 0x6f, 0x50, 0x72, 0x6f, 0x78, 0x79, 0x53, 0x74,
	0x72, 0x65, 0x61, 0x6d, 0x12, 0x6f, 0x0a, 0x07, 0x43, 0x6f, 0x6e, 0x73, 0x75, 0x6d, 0x65, 0x12,
	0x34, 0x2e, 0x63, 0x65, 0x6e, 0x74, 0x72, 0x69, 0x66, 0x75, 0x67, 0x61, 0x6c, 0x2e, 0x63, 0x65,
	0x6e, 0x74, 0x72, 0x69, 0x66, 0x75, 0x67, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x73, 0x74,
	0x72, 0x65, 0x61, 0x6d, 0x2e, 0x53, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x62, 0x65, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2c, 0x2e, 0x63, 0x65, 0x6e, 0x74, 0x72, 0x69, 0x66, 0x75,
	0x67, 0x61, 0x6c, 0x2e, 0x63, 0x65, 0x6e, 0x74, 0x72, 0x69, 0x66, 0x75, 0x67, 0x6f, 0x2e, 0x70,
	0x72, 0x6f, 0x78, 0x79, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x2e, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x30, 0x01, 0x12, 0x77, 0x0a, 0x0b, 0x43, 0x6f, 0x6d, 0x6d, 0x75, 0x6e, 0x69,
	0x63, 0x61, 0x74, 0x65, 0x12, 0x36, 0x2e, 0x63, 0x65, 0x6e, 0x74, 0x72, 0x69, 0x66, 0x75, 0x67,
	0x61, 0x6c, 0x2e, 0x63, 0x65, 0x6e, 0x74, 0x72, 0x69, 0x66, 0x75, 0x67, 0x6f, 0x2e, 0x70, 0x72,
	0x6f, 0x78, 0x79, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x2e, 0x43, 0x6f, 0x6d, 0x6d, 0x75, 0x6e,
	0x69, 0x63, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2c, 0x2e, 0x63,
	0x65, 0x6e, 0x74, 0x72, 0x69, 0x66, 0x75, 0x67, 0x61, 0x6c, 0x2e, 0x63, 0x65, 0x6e, 0x74, 0x72,
	0x69, 0x66, 0x75, 0x67, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x73, 0x74, 0x72, 0x65, 0x61,
	0x6d, 0x2e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x28, 0x01, 0x30, 0x01, 0x42, 0x15,
	0x5a, 0x13, 0x2e, 0x2f, 0x3b, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proxystream_proto_rawDescOnce sync.Once
	file_proxystream_proto_rawDescData = file_proxystream_proto_rawDesc
)

func file_proxystream_proto_rawDescGZIP() []byte {
	file_proxystream_proto_rawDescOnce.Do(func() {
		file_proxystream_proto_rawDescData = protoimpl.X.CompressGZIP(file_proxystream_proto_rawDescData)
	})
	return file_proxystream_proto_rawDescData
}

var file_proxystream_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_proxystream_proto_goTypes = []interface{}{
	(*CommunicateRequest)(nil), // 0: centrifugal.centrifugo.proxystream.CommunicateRequest
	(*Response)(nil),           // 1: centrifugal.centrifugo.proxystream.Response
	(*SubscribeResponse)(nil),  // 2: centrifugal.centrifugo.proxystream.SubscribeResponse
	(*SubscribeRequest)(nil),   // 3: centrifugal.centrifugo.proxystream.SubscribeRequest
	(*Publication)(nil),        // 4: centrifugal.centrifugo.proxystream.Publication
	nil,                        // 5: centrifugal.centrifugo.proxystream.Publication.TagsEntry
}
var file_proxystream_proto_depIdxs = []int32{
	3, // 0: centrifugal.centrifugo.proxystream.CommunicateRequest.subscribe_request:type_name -> centrifugal.centrifugo.proxystream.SubscribeRequest
	4, // 1: centrifugal.centrifugo.proxystream.CommunicateRequest.publication:type_name -> centrifugal.centrifugo.proxystream.Publication
	2, // 2: centrifugal.centrifugo.proxystream.Response.subscribe_response:type_name -> centrifugal.centrifugo.proxystream.SubscribeResponse
	4, // 3: centrifugal.centrifugo.proxystream.Response.publication:type_name -> centrifugal.centrifugo.proxystream.Publication
	5, // 4: centrifugal.centrifugo.proxystream.Publication.tags:type_name -> centrifugal.centrifugo.proxystream.Publication.TagsEntry
	3, // 5: centrifugal.centrifugo.proxystream.CentrifugoProxyStream.Consume:input_type -> centrifugal.centrifugo.proxystream.SubscribeRequest
	0, // 6: centrifugal.centrifugo.proxystream.CentrifugoProxyStream.Communicate:input_type -> centrifugal.centrifugo.proxystream.CommunicateRequest
	1, // 7: centrifugal.centrifugo.proxystream.CentrifugoProxyStream.Consume:output_type -> centrifugal.centrifugo.proxystream.Response
	1, // 8: centrifugal.centrifugo.proxystream.CentrifugoProxyStream.Communicate:output_type -> centrifugal.centrifugo.proxystream.Response
	7, // [7:9] is the sub-list for method output_type
	5, // [5:7] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_proxystream_proto_init() }
func file_proxystream_proto_init() {
	if File_proxystream_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proxystream_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CommunicateRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proxystream_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Response); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proxystream_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SubscribeResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proxystream_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SubscribeRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proxystream_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Publication); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proxystream_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proxystream_proto_goTypes,
		DependencyIndexes: file_proxystream_proto_depIdxs,
		MessageInfos:      file_proxystream_proto_msgTypes,
	}.Build()
	File_proxystream_proto = out.File
	file_proxystream_proto_rawDesc = nil
	file_proxystream_proto_goTypes = nil
	file_proxystream_proto_depIdxs = nil
}