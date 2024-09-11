// Code generated by capnpc-go. DO NOT EDIT.

package data

import (
	capnp "capnproto.org/go/capnp/v3"
	text "capnproto.org/go/capnp/v3/encoding/text"
	fc "capnproto.org/go/capnp/v3/flowcontrol"
	schemas "capnproto.org/go/capnp/v3/schemas"
	server "capnproto.org/go/capnp/v3/server"
	context "context"
)

type Artifact capnp.Struct

// Artifact_TypeID is the unique identifier for the type Artifact.
const Artifact_TypeID = 0xff7ca5c7f859f959

func NewArtifact(s *capnp.Segment) (Artifact, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 10})
	return Artifact(st), err
}

func NewRootArtifact(s *capnp.Segment) (Artifact, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 10})
	return Artifact(st), err
}

func ReadRootArtifact(msg *capnp.Message) (Artifact, error) {
	root, err := msg.Root()
	return Artifact(root.Struct()), err
}

func (s Artifact) String() string {
	str, _ := text.Marshal(0xff7ca5c7f859f959, capnp.Struct(s))
	return str
}

func (s Artifact) EncodeAsPtr(seg *capnp.Segment) capnp.Ptr {
	return capnp.Struct(s).EncodeAsPtr(seg)
}

func (Artifact) DecodeFromPtr(p capnp.Ptr) Artifact {
	return Artifact(capnp.Struct{}.DecodeFromPtr(p))
}

func (s Artifact) ToPtr() capnp.Ptr {
	return capnp.Struct(s).ToPtr()
}
func (s Artifact) IsValid() bool {
	return capnp.Struct(s).IsValid()
}

func (s Artifact) Message() *capnp.Message {
	return capnp.Struct(s).Message()
}

func (s Artifact) Segment() *capnp.Segment {
	return capnp.Struct(s).Segment()
}
func (s Artifact) Id() (string, error) {
	p, err := capnp.Struct(s).Ptr(0)
	return p.Text(), err
}

func (s Artifact) HasId() bool {
	return capnp.Struct(s).HasPtr(0)
}

func (s Artifact) IdBytes() ([]byte, error) {
	p, err := capnp.Struct(s).Ptr(0)
	return p.TextBytes(), err
}

func (s Artifact) SetId(v string) error {
	return capnp.Struct(s).SetText(0, v)
}

func (s Artifact) Checksum() ([]byte, error) {
	p, err := capnp.Struct(s).Ptr(1)
	return []byte(p.Data()), err
}

func (s Artifact) HasChecksum() bool {
	return capnp.Struct(s).HasPtr(1)
}

func (s Artifact) SetChecksum(v []byte) error {
	return capnp.Struct(s).SetData(1, v)
}

func (s Artifact) Pubkey() ([]byte, error) {
	p, err := capnp.Struct(s).Ptr(2)
	return []byte(p.Data()), err
}

func (s Artifact) HasPubkey() bool {
	return capnp.Struct(s).HasPtr(2)
}

func (s Artifact) SetPubkey(v []byte) error {
	return capnp.Struct(s).SetData(2, v)
}

func (s Artifact) Version() (string, error) {
	p, err := capnp.Struct(s).Ptr(3)
	return p.Text(), err
}

func (s Artifact) HasVersion() bool {
	return capnp.Struct(s).HasPtr(3)
}

func (s Artifact) VersionBytes() ([]byte, error) {
	p, err := capnp.Struct(s).Ptr(3)
	return p.TextBytes(), err
}

func (s Artifact) SetVersion(v string) error {
	return capnp.Struct(s).SetText(3, v)
}

func (s Artifact) Type() (string, error) {
	p, err := capnp.Struct(s).Ptr(4)
	return p.Text(), err
}

func (s Artifact) HasType() bool {
	return capnp.Struct(s).HasPtr(4)
}

func (s Artifact) TypeBytes() ([]byte, error) {
	p, err := capnp.Struct(s).Ptr(4)
	return p.TextBytes(), err
}

func (s Artifact) SetType(v string) error {
	return capnp.Struct(s).SetText(4, v)
}

func (s Artifact) Timestamp() uint64 {
	return capnp.Struct(s).Uint64(0)
}

func (s Artifact) SetTimestamp(v uint64) {
	capnp.Struct(s).SetUint64(0, v)
}

func (s Artifact) Origin() (string, error) {
	p, err := capnp.Struct(s).Ptr(5)
	return p.Text(), err
}

func (s Artifact) HasOrigin() bool {
	return capnp.Struct(s).HasPtr(5)
}

func (s Artifact) OriginBytes() ([]byte, error) {
	p, err := capnp.Struct(s).Ptr(5)
	return p.TextBytes(), err
}

func (s Artifact) SetOrigin(v string) error {
	return capnp.Struct(s).SetText(5, v)
}

func (s Artifact) Role() (string, error) {
	p, err := capnp.Struct(s).Ptr(6)
	return p.Text(), err
}

func (s Artifact) HasRole() bool {
	return capnp.Struct(s).HasPtr(6)
}

func (s Artifact) RoleBytes() ([]byte, error) {
	p, err := capnp.Struct(s).Ptr(6)
	return p.TextBytes(), err
}

func (s Artifact) SetRole(v string) error {
	return capnp.Struct(s).SetText(6, v)
}

func (s Artifact) Scope() (string, error) {
	p, err := capnp.Struct(s).Ptr(7)
	return p.Text(), err
}

func (s Artifact) HasScope() bool {
	return capnp.Struct(s).HasPtr(7)
}

func (s Artifact) ScopeBytes() ([]byte, error) {
	p, err := capnp.Struct(s).Ptr(7)
	return p.TextBytes(), err
}

func (s Artifact) SetScope(v string) error {
	return capnp.Struct(s).SetText(7, v)
}

func (s Artifact) Attributes() (Attribute_List, error) {
	p, err := capnp.Struct(s).Ptr(8)
	return Attribute_List(p.List()), err
}

func (s Artifact) HasAttributes() bool {
	return capnp.Struct(s).HasPtr(8)
}

func (s Artifact) SetAttributes(v Attribute_List) error {
	return capnp.Struct(s).SetPtr(8, v.ToPtr())
}

// NewAttributes sets the attributes field to a newly
// allocated Attribute_List, preferring placement in s's segment.
func (s Artifact) NewAttributes(n int32) (Attribute_List, error) {
	l, err := NewAttribute_List(capnp.Struct(s).Segment(), n)
	if err != nil {
		return Attribute_List{}, err
	}
	err = capnp.Struct(s).SetPtr(8, l.ToPtr())
	return l, err
}
func (s Artifact) Payload() ([]byte, error) {
	p, err := capnp.Struct(s).Ptr(9)
	return []byte(p.Data()), err
}

func (s Artifact) HasPayload() bool {
	return capnp.Struct(s).HasPtr(9)
}

func (s Artifact) SetPayload(v []byte) error {
	return capnp.Struct(s).SetData(9, v)
}

// Artifact_List is a list of Artifact.
type Artifact_List = capnp.StructList[Artifact]

// NewArtifact creates a new list of Artifact.
func NewArtifact_List(s *capnp.Segment, sz int32) (Artifact_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 10}, sz)
	return capnp.StructList[Artifact](l), err
}

// Artifact_Future is a wrapper for a Artifact promised by a client call.
type Artifact_Future struct{ *capnp.Future }

func (f Artifact_Future) Struct() (Artifact, error) {
	p, err := f.Future.Ptr()
	return Artifact(p.Struct()), err
}

type Attribute capnp.Struct

// Attribute_TypeID is the unique identifier for the type Attribute.
const Attribute_TypeID = 0xd1697cd3e7511b33

func NewAttribute(s *capnp.Segment) (Attribute, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return Attribute(st), err
}

func NewRootAttribute(s *capnp.Segment) (Attribute, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return Attribute(st), err
}

func ReadRootAttribute(msg *capnp.Message) (Attribute, error) {
	root, err := msg.Root()
	return Attribute(root.Struct()), err
}

func (s Attribute) String() string {
	str, _ := text.Marshal(0xd1697cd3e7511b33, capnp.Struct(s))
	return str
}

func (s Attribute) EncodeAsPtr(seg *capnp.Segment) capnp.Ptr {
	return capnp.Struct(s).EncodeAsPtr(seg)
}

func (Attribute) DecodeFromPtr(p capnp.Ptr) Attribute {
	return Attribute(capnp.Struct{}.DecodeFromPtr(p))
}

func (s Attribute) ToPtr() capnp.Ptr {
	return capnp.Struct(s).ToPtr()
}
func (s Attribute) IsValid() bool {
	return capnp.Struct(s).IsValid()
}

func (s Attribute) Message() *capnp.Message {
	return capnp.Struct(s).Message()
}

func (s Attribute) Segment() *capnp.Segment {
	return capnp.Struct(s).Segment()
}
func (s Attribute) Key() (string, error) {
	p, err := capnp.Struct(s).Ptr(0)
	return p.Text(), err
}

func (s Attribute) HasKey() bool {
	return capnp.Struct(s).HasPtr(0)
}

func (s Attribute) KeyBytes() ([]byte, error) {
	p, err := capnp.Struct(s).Ptr(0)
	return p.TextBytes(), err
}

func (s Attribute) SetKey(v string) error {
	return capnp.Struct(s).SetText(0, v)
}

func (s Attribute) Value() (string, error) {
	p, err := capnp.Struct(s).Ptr(1)
	return p.Text(), err
}

func (s Attribute) HasValue() bool {
	return capnp.Struct(s).HasPtr(1)
}

func (s Attribute) ValueBytes() ([]byte, error) {
	p, err := capnp.Struct(s).Ptr(1)
	return p.TextBytes(), err
}

func (s Attribute) SetValue(v string) error {
	return capnp.Struct(s).SetText(1, v)
}

// Attribute_List is a list of Attribute.
type Attribute_List = capnp.StructList[Attribute]

// NewAttribute creates a new list of Attribute.
func NewAttribute_List(s *capnp.Segment, sz int32) (Attribute_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2}, sz)
	return capnp.StructList[Attribute](l), err
}

// Attribute_Future is a wrapper for a Attribute promised by a client call.
type Attribute_Future struct{ *capnp.Future }

func (f Attribute_Future) Struct() (Attribute, error) {
	p, err := f.Future.Ptr()
	return Attribute(p.Struct()), err
}

type ModelService capnp.Client

// ModelService_TypeID is the unique identifier for the type ModelService.
const ModelService_TypeID = 0xee73f44b4fdab4e9

func (c ModelService) Query(ctx context.Context, params func(ModelService_query_Params) error) (ModelService_query_Results_Future, capnp.ReleaseFunc) {

	s := capnp.Send{
		Method: capnp.Method{
			InterfaceID:   0xee73f44b4fdab4e9,
			MethodID:      0,
			InterfaceName: "artifact.capnp:ModelService",
			MethodName:    "query",
		},
	}
	if params != nil {
		s.ArgsSize = capnp.ObjectSize{DataSize: 0, PointerCount: 1}
		s.PlaceArgs = func(s capnp.Struct) error { return params(ModelService_query_Params(s)) }
	}

	ans, release := capnp.Client(c).SendCall(ctx, s)
	return ModelService_query_Results_Future{Future: ans.Future()}, release

}

func (c ModelService) WaitStreaming() error {
	return capnp.Client(c).WaitStreaming()
}

// String returns a string that identifies this capability for debugging
// purposes.  Its format should not be depended on: in particular, it
// should not be used to compare clients.  Use IsSame to compare clients
// for equality.
func (c ModelService) String() string {
	return "ModelService(" + capnp.Client(c).String() + ")"
}

// AddRef creates a new Client that refers to the same capability as c.
// If c is nil or has resolved to null, then AddRef returns nil.
func (c ModelService) AddRef() ModelService {
	return ModelService(capnp.Client(c).AddRef())
}

// Release releases a capability reference.  If this is the last
// reference to the capability, then the underlying resources associated
// with the capability will be released.
//
// Release will panic if c has already been released, but not if c is
// nil or resolved to null.
func (c ModelService) Release() {
	capnp.Client(c).Release()
}

// Resolve blocks until the capability is fully resolved or the Context
// expires.
func (c ModelService) Resolve(ctx context.Context) error {
	return capnp.Client(c).Resolve(ctx)
}

func (c ModelService) EncodeAsPtr(seg *capnp.Segment) capnp.Ptr {
	return capnp.Client(c).EncodeAsPtr(seg)
}

func (ModelService) DecodeFromPtr(p capnp.Ptr) ModelService {
	return ModelService(capnp.Client{}.DecodeFromPtr(p))
}

// IsValid reports whether c is a valid reference to a capability.
// A reference is invalid if it is nil, has resolved to null, or has
// been released.
func (c ModelService) IsValid() bool {
	return capnp.Client(c).IsValid()
}

// IsSame reports whether c and other refer to a capability created by the
// same call to NewClient.  This can return false negatives if c or other
// are not fully resolved: use Resolve if this is an issue.  If either
// c or other are released, then IsSame panics.
func (c ModelService) IsSame(other ModelService) bool {
	return capnp.Client(c).IsSame(capnp.Client(other))
}

// Update the flowcontrol.FlowLimiter used to manage flow control for
// this client. This affects all future calls, but not calls already
// waiting to send. Passing nil sets the value to flowcontrol.NopLimiter,
// which is also the default.
func (c ModelService) SetFlowLimiter(lim fc.FlowLimiter) {
	capnp.Client(c).SetFlowLimiter(lim)
}

// Get the current flowcontrol.FlowLimiter used to manage flow control
// for this client.
func (c ModelService) GetFlowLimiter() fc.FlowLimiter {
	return capnp.Client(c).GetFlowLimiter()
}

// A ModelService_Server is a ModelService with a local implementation.
type ModelService_Server interface {
	Query(context.Context, ModelService_query) error
}

// ModelService_NewServer creates a new Server from an implementation of ModelService_Server.
func ModelService_NewServer(s ModelService_Server) *server.Server {
	c, _ := s.(server.Shutdowner)
	return server.New(ModelService_Methods(nil, s), s, c)
}

// ModelService_ServerToClient creates a new Client from an implementation of ModelService_Server.
// The caller is responsible for calling Release on the returned Client.
func ModelService_ServerToClient(s ModelService_Server) ModelService {
	return ModelService(capnp.NewClient(ModelService_NewServer(s)))
}

// ModelService_Methods appends Methods to a slice that invoke the methods on s.
// This can be used to create a more complicated Server.
func ModelService_Methods(methods []server.Method, s ModelService_Server) []server.Method {
	if cap(methods) == 0 {
		methods = make([]server.Method, 0, 1)
	}

	methods = append(methods, server.Method{
		Method: capnp.Method{
			InterfaceID:   0xee73f44b4fdab4e9,
			MethodID:      0,
			InterfaceName: "artifact.capnp:ModelService",
			MethodName:    "query",
		},
		Impl: func(ctx context.Context, call *server.Call) error {
			return s.Query(ctx, ModelService_query{call})
		},
	})

	return methods
}

// ModelService_query holds the state for a server call to ModelService.query.
// See server.Call for documentation.
type ModelService_query struct {
	*server.Call
}

// Args returns the call's arguments.
func (c ModelService_query) Args() ModelService_query_Params {
	return ModelService_query_Params(c.Call.Args())
}

// AllocResults allocates the results struct.
func (c ModelService_query) AllocResults() (ModelService_query_Results, error) {
	r, err := c.Call.AllocResults(capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return ModelService_query_Results(r), err
}

// ModelService_List is a list of ModelService.
type ModelService_List = capnp.CapList[ModelService]

// NewModelService_List creates a new list of ModelService.
func NewModelService_List(s *capnp.Segment, sz int32) (ModelService_List, error) {
	l, err := capnp.NewPointerList(s, sz)
	return capnp.CapList[ModelService](l), err
}

type ModelService_query_Params capnp.Struct

// ModelService_query_Params_TypeID is the unique identifier for the type ModelService_query_Params.
const ModelService_query_Params_TypeID = 0xf0631ef284a5d4bf

func NewModelService_query_Params(s *capnp.Segment) (ModelService_query_Params, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return ModelService_query_Params(st), err
}

func NewRootModelService_query_Params(s *capnp.Segment) (ModelService_query_Params, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return ModelService_query_Params(st), err
}

func ReadRootModelService_query_Params(msg *capnp.Message) (ModelService_query_Params, error) {
	root, err := msg.Root()
	return ModelService_query_Params(root.Struct()), err
}

func (s ModelService_query_Params) String() string {
	str, _ := text.Marshal(0xf0631ef284a5d4bf, capnp.Struct(s))
	return str
}

func (s ModelService_query_Params) EncodeAsPtr(seg *capnp.Segment) capnp.Ptr {
	return capnp.Struct(s).EncodeAsPtr(seg)
}

func (ModelService_query_Params) DecodeFromPtr(p capnp.Ptr) ModelService_query_Params {
	return ModelService_query_Params(capnp.Struct{}.DecodeFromPtr(p))
}

func (s ModelService_query_Params) ToPtr() capnp.Ptr {
	return capnp.Struct(s).ToPtr()
}
func (s ModelService_query_Params) IsValid() bool {
	return capnp.Struct(s).IsValid()
}

func (s ModelService_query_Params) Message() *capnp.Message {
	return capnp.Struct(s).Message()
}

func (s ModelService_query_Params) Segment() *capnp.Segment {
	return capnp.Struct(s).Segment()
}
func (s ModelService_query_Params) Request() (Artifact, error) {
	p, err := capnp.Struct(s).Ptr(0)
	return Artifact(p.Struct()), err
}

func (s ModelService_query_Params) HasRequest() bool {
	return capnp.Struct(s).HasPtr(0)
}

func (s ModelService_query_Params) SetRequest(v Artifact) error {
	return capnp.Struct(s).SetPtr(0, capnp.Struct(v).ToPtr())
}

// NewRequest sets the request field to a newly
// allocated Artifact struct, preferring placement in s's segment.
func (s ModelService_query_Params) NewRequest() (Artifact, error) {
	ss, err := NewArtifact(capnp.Struct(s).Segment())
	if err != nil {
		return Artifact{}, err
	}
	err = capnp.Struct(s).SetPtr(0, capnp.Struct(ss).ToPtr())
	return ss, err
}

// ModelService_query_Params_List is a list of ModelService_query_Params.
type ModelService_query_Params_List = capnp.StructList[ModelService_query_Params]

// NewModelService_query_Params creates a new list of ModelService_query_Params.
func NewModelService_query_Params_List(s *capnp.Segment, sz int32) (ModelService_query_Params_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return capnp.StructList[ModelService_query_Params](l), err
}

// ModelService_query_Params_Future is a wrapper for a ModelService_query_Params promised by a client call.
type ModelService_query_Params_Future struct{ *capnp.Future }

func (f ModelService_query_Params_Future) Struct() (ModelService_query_Params, error) {
	p, err := f.Future.Ptr()
	return ModelService_query_Params(p.Struct()), err
}
func (p ModelService_query_Params_Future) Request() Artifact_Future {
	return Artifact_Future{Future: p.Future.Field(0, nil)}
}

type ModelService_query_Results capnp.Struct

// ModelService_query_Results_TypeID is the unique identifier for the type ModelService_query_Results.
const ModelService_query_Results_TypeID = 0x8981d3c40ae36ecc

func NewModelService_query_Results(s *capnp.Segment) (ModelService_query_Results, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return ModelService_query_Results(st), err
}

func NewRootModelService_query_Results(s *capnp.Segment) (ModelService_query_Results, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return ModelService_query_Results(st), err
}

func ReadRootModelService_query_Results(msg *capnp.Message) (ModelService_query_Results, error) {
	root, err := msg.Root()
	return ModelService_query_Results(root.Struct()), err
}

func (s ModelService_query_Results) String() string {
	str, _ := text.Marshal(0x8981d3c40ae36ecc, capnp.Struct(s))
	return str
}

func (s ModelService_query_Results) EncodeAsPtr(seg *capnp.Segment) capnp.Ptr {
	return capnp.Struct(s).EncodeAsPtr(seg)
}

func (ModelService_query_Results) DecodeFromPtr(p capnp.Ptr) ModelService_query_Results {
	return ModelService_query_Results(capnp.Struct{}.DecodeFromPtr(p))
}

func (s ModelService_query_Results) ToPtr() capnp.Ptr {
	return capnp.Struct(s).ToPtr()
}
func (s ModelService_query_Results) IsValid() bool {
	return capnp.Struct(s).IsValid()
}

func (s ModelService_query_Results) Message() *capnp.Message {
	return capnp.Struct(s).Message()
}

func (s ModelService_query_Results) Segment() *capnp.Segment {
	return capnp.Struct(s).Segment()
}
func (s ModelService_query_Results) Response() (Artifact, error) {
	p, err := capnp.Struct(s).Ptr(0)
	return Artifact(p.Struct()), err
}

func (s ModelService_query_Results) HasResponse() bool {
	return capnp.Struct(s).HasPtr(0)
}

func (s ModelService_query_Results) SetResponse(v Artifact) error {
	return capnp.Struct(s).SetPtr(0, capnp.Struct(v).ToPtr())
}

// NewResponse sets the response field to a newly
// allocated Artifact struct, preferring placement in s's segment.
func (s ModelService_query_Results) NewResponse() (Artifact, error) {
	ss, err := NewArtifact(capnp.Struct(s).Segment())
	if err != nil {
		return Artifact{}, err
	}
	err = capnp.Struct(s).SetPtr(0, capnp.Struct(ss).ToPtr())
	return ss, err
}

// ModelService_query_Results_List is a list of ModelService_query_Results.
type ModelService_query_Results_List = capnp.StructList[ModelService_query_Results]

// NewModelService_query_Results creates a new list of ModelService_query_Results.
func NewModelService_query_Results_List(s *capnp.Segment, sz int32) (ModelService_query_Results_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return capnp.StructList[ModelService_query_Results](l), err
}

// ModelService_query_Results_Future is a wrapper for a ModelService_query_Results promised by a client call.
type ModelService_query_Results_Future struct{ *capnp.Future }

func (f ModelService_query_Results_Future) Struct() (ModelService_query_Results, error) {
	p, err := f.Future.Ptr()
	return ModelService_query_Results(p.Struct()), err
}
func (p ModelService_query_Results_Future) Response() Artifact_Future {
	return Artifact_Future{Future: p.Future.Field(0, nil)}
}

const schema_e363a5839bf866c4 = "x\xda\x8c\x93\xbdk4U\x18\xc5\xcf\xb93\xb3_d" +
	"\xdd\x0cw\x04QqU\"\xbc\xaf$/\xf9\x12L\x88" +
	"d\x0d\x8a1\x1a\xdc\xc9h\x11\xbb\xc9\xec\x8d\x0e\xd9\xdd" +
	"\x99\xccGd!\x10\x826\xfe\x0d\xd6[Z\x08\xf6\x82" +
	"\x04\x04\xb1\x91XY\x9aBlD\xc4\"V#w\xd9" +
	"/\xb0I7\xcf\x8f\xe7y\xce\xbd\xe7\x9eY\xbdd\xcb" +
	"\\\xab7\x0d\x08w\xc9*\x15?\xf5\xefj7\xb7\xd7" +
	"_\xc2~\x9e\x80\xc52\xb0\x91\xf3\x88\xa0\xbc\xe6.X" +
	"l<\xe7\xfe~{\x19\xfe\x0c\xbb\xc1\xe2\xe6\xf4\xfe\xab" +
	"\xcf\x87\xc1\x1d,Q\x06\xe4\x90?\xcao\xf4\x88\xfc\x9a" +
	"\x9f\x81\xc5\x1f\xdf\xfe\xfa\xc1{\xff\xa4\x7f\xc2n\x18\xb3" +
	"^PV\xc5o\xf2\xe9\xd1\x88-\xde\x91[\xfa\xab\xf8" +
	"\xee\x97\xe1\x17\x7f\xbf\x10\xfc5\xaf\xfc\x928\xd0\xca+" +
	"B+\x1f\xff{|\xff\xc3\xf0\xb2\x80\xdb\xe0\xbctM" +
	"\xef9\x14\xdf\xcb\x8f\xf4\x9e\x0dWD\x02+\x85\x9fd" +
	"\xe1\xa9\x1fd\xe6\x93\xc0\x8f\xfb\xf1\xf6a\xd4Q]O" +
	"%\x17a\xa0\x9e\x9c\xe7*\x19,\x1d\xa94\xeffL" +
	"]\xd30\x01\x93\x80]?\x00\xdc\x05\x83\xee3\x82E" +
	"\xa2\xd28\xea\xa7\x0a\x00\x17g\xf2 \x17\xc1\xa9\x80\x18" +
	"\x0b\xbc\x99eIx\x92g\x0ah\x93ne\xba\xf3\xf1" +
	"\xcb\x80\xbbd\xd0]\x15\xb4I\x87\x1a\xae\xac\x03\xee#" +
	"\x83\xee\xa6`\xf9L\x0d\xb8\x00\xc1\x05\xb0y\xe1ws" +
	"5\xa9\xfe'2\xbdE9\x0c\x94\x961\x0dk\xce;" +
	"N\x9e\xcf\xb6\xd7!l\xab\xdc\x1c\xdd\xb4\xc56\xf9\x10" +
	"G\xda~\xe2\xf7R`\xde\x91=\xc0\xad\x18t\x1d\xc1" +
	"\xabD\x9d\xe7*\xcd\x1e\xe2\xc6\xb8\x1ey\xb1<\xd9&" +
	"_\xe1\xb3\x80\xf7\"\x0dz\xcb\x9c\xd9!\x1f\xf3\x00\xf0" +
	"\x1ei\xbe\xa9\xb9\x10\x0e\x05 \xd7\xb8\x0dx\xcb\x9a\xbf" +
	"\xae\xb9a84\x00\xf9\x1a\xf7\x00oU\xf3\x1d\xcdM" +
	"\xd3\xa1\x09\xc8-\xbe\x0ax\x9b\x9a\xb7(H\xcb\xa1\x05" +
	"\xc87x\x04x;\x1a\xef\xeb\xf6\x92\xe5\xb0\x04\xc8\xb7" +
	"G\xeb[\x9a\xbf\xafy\xb9\xe4\x8c\xe2\xfb\xeeh\xcd[" +
	"\x9a\xb75\xaf\x94\x1dVt\xca\xb8\x0ex\xfb\x9a\x7f\xa8" +
	"y\xb5\xe2\xb0\x0aH\x97\x1f\x03^[\xf3\xae\xe6\xb5\xaa" +
	"\xc3\x1a \xc3\xd11;\x9a\xc7\x144\xc2\xce\xf4e\x83" +
	"OUp\x96\xe6=\x1d\xae:\x04\xeb\xe0n\x9c\x9f\xe8" +
	"(\x8c\xcb\xab\x0b\x95\xa4a\xd4\x9f\x8c4\xb2A<K" +
	"F\x16\xf6T\x9a\xf9=0f\x15\x82Up7J\xc2" +
	"O\xc2Y\x7f\x12u\xa7\xfd\xcd4\x88\xe6\xa6\xfdqX" +
	"a\xa8\x94O\x81m\x83\\\x9c\xfd\xdb\xa0\x86W\xb1?" +
	"\xe8F~gr\xa0\xff\x02\x00\x00\xff\xff\xee\x83\xed\x86"

func RegisterSchema(reg *schemas.Registry) {
	reg.Register(&schemas.Schema{
		String: schema_e363a5839bf866c4,
		Nodes: []uint64{
			0x8981d3c40ae36ecc,
			0xd1697cd3e7511b33,
			0xee73f44b4fdab4e9,
			0xf0631ef284a5d4bf,
			0xff7ca5c7f859f959,
		},
		Compressed: true,
	})
}
