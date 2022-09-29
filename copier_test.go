/*
 * Copyright (c) 2022 by David Wartell. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package copier

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"math"
	"net"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestMarshal(t *testing.T) {
	var jsonDebug []byte
	var err error

	testSrcStruct, testDstStruct, testCopyBackStruct := testMarshal(t)

	jsonDebug, err = json.MarshalIndent(testSrcStruct, "", "    ")
	require.Nil(t, err)
	t.Log(string(jsonDebug))

	jsonDebug, err = json.MarshalIndent(testDstStruct, "", "    ")
	require.Nil(t, err)
	t.Log(string(jsonDebug))

	jsonDebug, err = json.MarshalIndent(testCopyBackStruct, "", "    ")
	require.Nil(t, err)
	t.Log(string(jsonDebug))

	assert.Equal(t, testSrcStruct, testCopyBackStruct)
}

func BenchmarkMarshal(b *testing.B) {
	for n := 0; n < b.N; n++ {
		testMarshal(b)
	}
}

func testMarshal(t require.TestingT) (testSrcStruct *TestMarshalStructA, testDstStruct *TestMarshalStructAProtoBuf, testCopyBackStruct *TestMarshalStructA) {
	var err error

	testSrcStruct, err = makeMockStruct()
	require.Nil(t, err)
	testDstStruct = &TestMarshalStructAProtoBuf{}

	err = Copy(testDstStruct, testSrcStruct)
	require.Nil(t, err)

	// copy back and compare
	testCopyBackStruct = &TestMarshalStructA{}
	err = Copy(testCopyBackStruct, testDstStruct)
	require.Nil(t, err)

	return
}

func makeMockStruct() (testStruct *TestMarshalStructA, err error) {
	var hwAddr net.HardwareAddr
	hwAddr, err = net.ParseMAC("6e:e4:fe:20:9a:4d")
	if err != nil {
		return
	}
	mac := MarshalMacMakeMacFromHardwareAddr(hwAddr)

	stringPtrStringWrapper := "StringPtrStringWrapper"
	stringPtrString := "stringPtrString"
	stringPtrTestMarshal := "6e:e4:fe:20:9a:4d"
	stringPtrTextMarshalPtr := "6e:e4:fe:20:9a:4d"
	timePtrString := time.Now().UTC()
	intPtr := math.MinInt
	intPtrPtr := &intPtr
	int32PtrWrapper := int32(math.MinInt32)
	boolPtr := true
	boolPtrPtr := &boolPtr
	uintPtr := uint(math.MaxUint)
	uintPtrPtr := &uintPtr
	uint32Ptr := uint32(math.MaxUint32)
	float32Ptr := float32(math.MaxFloat32)
	float32PtrPtr := &float32Ptr
	intSlicePtr := []int{1, 2, 3, 4, 5}
	intSlicePtrPtr := &intSlicePtr
	strPtrPtr := "StringPtrPtrStringPtrPtr"
	ptrToStrPtrPtr := &strPtrPtr
	strPtrPtrPtr := "StringPtrPtrPtrStringPtrPtrPtr"
	ptrToStrPtrPtrPtr := &strPtrPtrPtr
	ptrToPtrToStrPtrPtrPtr := &ptrToStrPtrPtrPtr
	testStructA := TestStructA{
		Uint:         math.MaxUint,
		StringString: "StringString",
	}
	testStructAPtr := &testStructA
	durationPtr := time.Second * time.Duration(500)

	//goland:noinspection SpellCheckingInspection
	testStruct = &TestMarshalStructA{
		StringAliasString:         TestMarshalId("TEST_STRING_ALIAS"),
		StringString:              "StringString",
		StringStringWrapper:       "StringStringWrapper",
		StringPtrStringWrapper:    &stringPtrStringWrapper,
		StringNilPtrStringWrapper: nil,
		StringPtrString:           &stringPtrString,
		StringNilPtrString:        nil,
		NilPtrStringString:        "",
		StringWrapperStringPtr: &wrapperspb.StringValue{
			Value: "StringWrapperStringPtr",
		},
		StringWrapperString: &wrapperspb.StringValue{
			Value: "StringWrapperString",
		},
		StringStringAlias: "StringStringAlias",

		TextMarshalerString:           mac,
		TextMarshalerPtrString:        &mac,
		TextMarshalerPtrStringWrapper: &mac,
		StringTestMarshal:             "6e:e4:fe:20:9a:4d",
		StringPtrTestMarshal:          &stringPtrTestMarshal,
		StringTextMarshalPtr:          "6e:e4:fe:20:9a:4d",
		StringPtrTextMarshalPtr:       &stringPtrTextMarshalPtr,
		StringPtrPtrStringPtrPtr:      &ptrToStrPtrPtr,

		TimeString:              time.Now().UTC(),
		StringTime:              time.Now().UTC().Format(time.RFC3339Nano),
		TimePtrString:           &timePtrString,
		TimeProtoTimestamp:      time.Now().UTC(),
		TimePtrProtoTimestamp:   &timePtrString,
		ProtoTimeStampToTime:    timestamppb.New(time.Now().UTC()),
		ProtoTimeStampToTimePtr: timestamppb.New(time.Now().UTC()),

		DurationProtoDuration:      time.Second * time.Duration(500),
		DurationPtrProtoDuration:   &durationPtr,
		ProtoDurationToDuration:    durationpb.New(durationPtr),
		ProtoDurationToDurationPtr: durationpb.New(durationPtr),

		Bool:            true,
		BoolPtrBool:     &boolPtr,
		BoolBoolPtr:     true,
		BoolBoolWrapper: true,
		BoolWrapperBool: &wrapperspb.BoolValue{
			Value: true,
		},
		BoolPtrBoolWrapper: &boolPtr,
		BoolPtrPtr:         &boolPtrPtr,

		Int:              math.MinInt,
		Int8:             math.MinInt8,
		Int16:            math.MinInt16,
		Int32:            math.MinInt32,
		Int64:            math.MinInt64,
		IntPtrInt:        &intPtr,
		IntIntPtr:        math.MinInt,
		Int8Int32Wrapper: math.MinInt8,
		Int32WrapperInt8: &wrapperspb.Int32Value{
			Value: math.MinInt8,
		},
		Int16Int64Wrapper: math.MinInt16,
		Int64WrapperInt16: &wrapperspb.Int64Value{
			Value: math.MinInt16,
		},
		Int32PtrWrapper: &int32PtrWrapper,
		Int32WrapperInt32Ptr: &wrapperspb.Int32Value{
			Value: math.MinInt32,
		},
		IntZeroIntPtr:           0,
		IntPtrNilToInt32Wrapper: nil,
		Int32WrapperNilToIntPtr: nil,
		IntPtrPtr:               &intPtrPtr,

		Uint:               math.MaxUint,
		Uint8:              math.MaxUint8,
		Uint16:             math.MaxUint16,
		Uint32:             math.MaxUint32,
		Uint64:             math.MaxUint64,
		UintPtrUint:        &uintPtr,
		UintUintPtr:        math.MaxUint,
		Uint8Uint32Wrapper: math.MaxUint8,
		Uint32WrapperUint8: &wrapperspb.UInt32Value{
			Value: math.MaxUint8,
		},
		Uint16Uint64Wrapper: math.MaxUint16,
		Uint64WrapperUint16: &wrapperspb.UInt64Value{
			Value: math.MaxUint16,
		},
		Uint32PtrWrapper: &uint32Ptr,
		Uint32WrapperUint32Ptr: &wrapperspb.UInt32Value{
			Value: math.MaxUint32,
		},
		UintPtrPtr: &uintPtrPtr,

		Float32:             math.MaxFloat32,
		Float64:             math.MaxFloat64,
		Float32PtrFloat32:   &float32Ptr,
		Float32Float32Ptr:   math.MaxFloat32,
		Float32FloatWrapper: math.MaxFloat32,
		FloatWrapperFloat32: &wrapperspb.FloatValue{
			Value: math.MaxFloat32,
		},
		Float32PtrWrapper: &float32Ptr,
		FloatWrapperFloat32Ptr: &wrapperspb.FloatValue{
			Value: math.MaxFloat32,
		},
		Float32PtrPtr: &float32PtrPtr,

		ByteByte: byte(127), // byte is alias for uint8 nothing special needed

		IntSlice:                       []int{1, 2, 3, 4, 5},
		IntSlicePtr:                    &intSlicePtr,
		IntArray:                       [5]int{1, 2, 3, 4, 5},
		StringPtrPtrPtrStringPtrPtrPtr: &ptrToPtrToStrPtrPtrPtr,
		IntSlicePtrPtr:                 &intSlicePtrPtr,
		ByteSlice:                      []byte{120, 121, 123, 124, 125},
		ByteSliceBytesWrapper:          []byte{120, 121, 123, 124, 125},
		BytesWrapperBytesWrapper: &wrapperspb.BytesValue{
			Value: []byte{120, 121, 123, 124, 125},
		},
		BytesWrapperByteSlice: &wrapperspb.BytesValue{
			Value: []byte{120, 121, 123, 124, 125},
		},

		StringMapStringMap: map[string]string{
			"Key-StringMapStringMap": "Value-StringMapStringMap",
		},

		ArrayMapArrayMap: map[string][]string{
			"Key-StringMapStringMap": {"ArrayValue-StringMapStringMap"},
		},

		StructField:              testStructA,
		StructPtrStruct:          &testStructA,
		PtrPtrStructPtrPtrStruct: &testStructAPtr,

		TestDocWrapperA: &TestDocWrapper{
			Type: TestMarshalId("A"),
			Doc: &TestInterfaceImplA{
				Int: math.MinInt,
			},
		},
		TestDocWrapperB: &TestDocWrapper{
			Type: TestMarshalId("B"),
			Doc: &TestInterfaceImplB{
				Float: math.MaxFloat32,
			},
		},
		TestDocWrapperProtoA: &TestDocWrapperProto{
			Type: "A",
			Doc: &TestDocWrapperProtoImplA{
				Int: math.MaxInt,
			},
		},
		TestDocWrapperProtoB: &TestDocWrapperProto{
			Type: "B",
			Doc: &TestDocWrapperProtoImplB{
				Float: math.MaxFloat32,
			},
		},

		TestHttpHeaderToProto: HttpHeader{
			"Strict-Transport-Security": {"max-age=15552000; includeSubDomains", "test array"},
			"X-Download-Options":        {"noopen"},
			"Content-Security-Policy":   {"default-src 'self'"},
			"Date":                      {"Mon, 04 Apr 2022 02:37:08 GMT"},
			"X-Xss-Protection":          {"1; mode=block"},
		},
		TestHttpHeaderProtoToHttpHeader: map[string]*ProtoHttpHeaderElement{
			"Strict-Transport-Security": {Values: []string{"max-age=15552000; includeSubDomains", "test array"}},
			"X-Download-Options":        {Values: []string{"max-age=15552000; includeSubDomains", "test array"}},
			"Content-Security-Policy":   {Values: []string{"default-src 'self'"}},
			"Date":                      {Values: []string{"Mon, 04 Apr 2022 02:37:08 GMT"}},
			"X-Xss-Protection":          {Values: []string{"1; mode=block"}},
		},
	}
	return
}

type TestMarshalId string

func (t TestMarshalId) String() string {
	return string(t)
}

type TestMarshalStructA struct {
	StringAliasString              TestMarshalId
	StringStringAlias              string
	StringString                   string
	StringStringWrapper            string
	StringPtrStringWrapper         *string
	StringNilPtrStringWrapper      *string
	StringPtrString                *string
	StringNilPtrString             *string
	NilPtrStringString             string
	StringWrapperStringPtr         *wrapperspb.StringValue
	StringWrapperString            *wrapperspb.StringValue
	StringPtrPtrStringPtrPtr       **string
	StringPtrPtrPtrStringPtrPtrPtr ***string

	TextMarshalerString           TestMarshalMac
	TextMarshalerPtrString        *TestMarshalMac
	TextMarshalerPtrStringWrapper *TestMarshalMac
	StringTestMarshal             string
	StringPtrTestMarshal          *string
	StringTextMarshalPtr          string
	StringPtrTextMarshalPtr       *string

	TimeString              time.Time
	StringTime              string
	TimePtrString           *time.Time
	TimeProtoTimestamp      time.Time
	TimePtrProtoTimestamp   *time.Time
	ProtoTimeStampToTime    *timestamppb.Timestamp
	ProtoTimeStampToTimePtr *timestamppb.Timestamp

	DurationProtoDuration      time.Duration
	DurationPtrProtoDuration   *time.Duration
	ProtoDurationToDuration    *durationpb.Duration
	ProtoDurationToDurationPtr *durationpb.Duration

	Bool               bool
	BoolPtrBool        *bool
	BoolBoolPtr        bool
	BoolBoolWrapper    bool
	BoolWrapperBool    *wrapperspb.BoolValue
	BoolPtrBoolWrapper *bool
	BoolPtrPtr         **bool

	Int                     int
	Int8                    int8
	Int16                   int16
	Int32                   int32
	Int64                   int64
	IntPtrInt               *int
	IntIntPtr               int
	Int8Int32Wrapper        int8
	Int32WrapperInt8        *wrapperspb.Int32Value
	Int16Int64Wrapper       int16
	Int64WrapperInt16       *wrapperspb.Int64Value
	Int32PtrWrapper         *int32
	Int32WrapperInt32Ptr    *wrapperspb.Int32Value
	IntZeroIntPtr           int
	IntPtrNilToInt32Wrapper *int
	Int32WrapperNilToIntPtr *wrapperspb.Int32Value
	IntPtrPtr               **int

	Uint                   uint
	Uint8                  uint8
	Uint16                 uint16
	Uint32                 uint32
	Uint64                 uint64
	UintPtrUint            *uint
	UintUintPtr            uint
	Uint8Uint32Wrapper     uint8
	Uint32WrapperUint8     *wrapperspb.UInt32Value
	Uint16Uint64Wrapper    uint16
	Uint64WrapperUint16    *wrapperspb.UInt64Value
	Uint32PtrWrapper       *uint32
	Uint32WrapperUint32Ptr *wrapperspb.UInt32Value
	UintPtrPtr             **uint

	Float32                float32
	Float64                float64
	Float32PtrFloat32      *float32
	Float32Float32Ptr      float32
	Float32FloatWrapper    float64
	FloatWrapperFloat32    *wrapperspb.FloatValue
	Float32PtrWrapper      *float32
	FloatWrapperFloat32Ptr *wrapperspb.FloatValue
	Float32PtrPtr          **float32

	ByteByte byte

	IntSlice                 []int
	IntSlicePtr              *[]int
	IntArray                 [5]int
	IntSlicePtrPtr           **[]int
	ByteSlice                []byte
	ByteSliceBytesWrapper    []byte
	BytesWrapperBytesWrapper *wrapperspb.BytesValue
	BytesWrapperByteSlice    *wrapperspb.BytesValue

	StringMapStringMap map[string]string
	ArrayMapArrayMap   map[string][]string

	StructField              TestStructA
	StructPtrStruct          *TestStructA
	PtrPtrStructPtrPtrStruct **TestStructA

	TestDocWrapperA      *TestDocWrapper
	TestDocWrapperB      *TestDocWrapper
	TestDocWrapperProtoA *TestDocWrapperProto
	TestDocWrapperProtoB *TestDocWrapperProto

	TestHttpHeaderToProto           HttpHeader
	TestHttpHeaderProtoToHttpHeader map[string]*ProtoHttpHeaderElement
}

type TestMarshalStructAProtoBuf struct {
	StringAliasString              string
	StringStringAlias              TestMarshalId
	StringString                   string
	StringStringWrapper            *wrapperspb.StringValue
	StringPtrStringWrapper         *wrapperspb.StringValue
	StringNilPtrStringWrapper      *wrapperspb.StringValue
	StringPtrString                string
	StringNilPtrString             string
	NilPtrStringString             *string
	StringWrapperStringPtr         *string
	StringWrapperString            string
	StringPtrPtrStringPtrPtr       **string
	StringPtrPtrPtrStringPtrPtrPtr ***string

	TextMarshalerString           string
	TextMarshalerPtrString        string
	TextMarshalerPtrStringWrapper *wrapperspb.StringValue
	StringTestMarshal             TestMarshalMac
	StringPtrTestMarshal          TestMarshalMac
	StringTextMarshalPtr          *TestMarshalMac
	StringPtrTextMarshalPtr       *TestMarshalMac

	TimeString              string
	StringTime              time.Time
	TimePtrString           string
	TimeProtoTimestamp      *timestamppb.Timestamp
	TimePtrProtoTimestamp   *timestamppb.Timestamp
	ProtoTimeStampToTime    time.Time
	ProtoTimeStampToTimePtr *time.Time

	DurationProtoDuration      *durationpb.Duration
	DurationPtrProtoDuration   *durationpb.Duration
	ProtoDurationToDuration    time.Duration
	ProtoDurationToDurationPtr *time.Duration

	Bool               bool
	BoolPtrBool        bool
	BoolBoolPtr        *bool
	BoolBoolWrapper    *wrapperspb.BoolValue
	BoolWrapperBool    bool
	BoolPtrBoolWrapper *wrapperspb.BoolValue
	BoolPtrPtr         **bool

	Int                     int
	Int8                    int8
	Int16                   int16
	Int32                   int32
	Int64                   int64
	IntPtrInt               int
	IntIntPtr               *int
	IntPtrPtr               **int
	Int8Int32Wrapper        *wrapperspb.Int32Value
	Int32WrapperInt8        int8
	Int16Int64Wrapper       *wrapperspb.Int64Value
	Int64WrapperInt16       int16
	Int32PtrWrapper         *wrapperspb.Int64Value
	Int32WrapperInt32Ptr    *int32
	IntZeroIntPtr           *int
	IntPtrNilToInt32Wrapper *wrapperspb.Int32Value
	Int32WrapperNilToIntPtr *int

	Uint                   uint
	Uint8                  uint8
	Uint16                 uint16
	Uint32                 uint32
	Uint64                 uint64
	UintPtrUint            uint
	UintUintPtr            *uint
	Uint8Uint32Wrapper     *wrapperspb.UInt32Value
	Uint32WrapperUint8     uint8
	Uint16Uint64Wrapper    *wrapperspb.UInt64Value
	Uint64WrapperUint16    uint16
	Uint32PtrWrapper       *wrapperspb.UInt32Value
	Uint32WrapperUint32Ptr *uint32
	UintPtrPtr             **uint

	Float32                float32
	Float64                float64
	Float32PtrFloat32      float32
	Float32Float32Ptr      *float32
	Float32FloatWrapper    *wrapperspb.FloatValue
	FloatWrapperFloat32    float64
	Float32PtrWrapper      *wrapperspb.FloatValue
	FloatWrapperFloat32Ptr *float32
	Float32PtrPtr          **float32

	ByteByte byte

	IntSlice                 []int
	IntSlicePtr              *[]int
	IntArray                 [5]int
	IntSlicePtrPtr           **[]int
	ByteSlice                []byte
	ByteSliceBytesWrapper    *wrapperspb.BytesValue
	BytesWrapperBytesWrapper *wrapperspb.BytesValue
	BytesWrapperByteSlice    []byte

	StringMapStringMap map[string]string
	ArrayMapArrayMap   map[string][]string

	StructField              TestStructB
	StructPtrStruct          TestStructB
	PtrPtrStructPtrPtrStruct **TestStructB

	TestDocWrapperA      *TestDocWrapperProto
	TestDocWrapperB      *TestDocWrapperProto
	TestDocWrapperProtoA *TestDocWrapper
	TestDocWrapperProtoB *TestDocWrapper

	TestHttpHeaderToProto           map[string]*ProtoHttpHeaderElement
	TestHttpHeaderProtoToHttpHeader HttpHeader
}

type TestDocWrapper struct {
	Type TestMarshalId
	Doc  TestInterface
}

func (d TestDocWrapper) CopierMarshal() (Any, error) {
	switch d.Type {
	case "A":
		var implA TestInterfaceImplA
		impl, ok := d.Doc.(*TestInterfaceImplA)
		if !ok {
			return nil, errors.Errorf("found type A and expected type %s and found %s", reflect.TypeOf(&implA).String(), reflect.TypeOf(d.Doc).String())
		}

		var protImpl TestDocWrapperProtoImplA
		err := Copy(&protImpl, impl)
		if err != nil {
			return nil, err
		}
		return &TestDocWrapperProto{
			Type: d.Type.String(),
			Doc:  &protImpl,
		}, nil
	case "B":
		var implB TestInterfaceImplB
		impl, ok := d.Doc.(*TestInterfaceImplB)
		if !ok {
			return nil, errors.Errorf("found type A and expected type %s and found %s", reflect.TypeOf(&implB).String(), reflect.TypeOf(d.Doc).String())
		}

		var protImpl TestDocWrapperProtoImplB
		err := Copy(&protImpl, impl)
		if err != nil {
			return nil, err
		}
		return &TestDocWrapperProto{
			Type: d.Type.String(),
			Doc:  &protImpl,
		}, nil
	}
	return nil, errors.Errorf("unexpected type %s found", reflect.TypeOf(d.Doc).String())
}

func (d *TestDocWrapper) CopierUnmarshal(a Any) error {
	var protoWrapperType TestDocWrapperProto
	protoWrapper, wrapperOk := a.(*TestDocWrapperProto)
	if !wrapperOk {
		return errors.Errorf("expected type %s and found %s", reflect.TypeOf(&protoWrapperType).String(), reflect.TypeOf(protoWrapper).String())
	}
	d.Type = TestMarshalId(protoWrapper.Type)
	switch protoWrapper.Type {
	case "A":
		var protoImplA TestDocWrapperProtoImplA
		protoImpl, ok := protoWrapper.Doc.(*TestDocWrapperProtoImplA)
		if !ok {
			return errors.Errorf("found type A and expected type %s and found %s", reflect.TypeOf(&protoImplA).String(), reflect.TypeOf(protoWrapper.Doc).String())
		}

		var impl TestInterfaceImplA
		err := Copy(&impl, protoImpl)
		if err != nil {
			return err
		}
		d.Doc = &impl
	case "B":
		var protoImplB TestDocWrapperProtoImplB
		protoImpl, ok := protoWrapper.Doc.(*TestDocWrapperProtoImplB)
		if !ok {
			return errors.Errorf("found type B and expected type %s and found %s", reflect.TypeOf(&protoImplB).String(), reflect.TypeOf(protoWrapper.Doc).String())
		}

		var impl TestInterfaceImplB
		err := Copy(&impl, protoImpl)
		if err != nil {
			return err
		}
		d.Doc = &impl
	default:
		return errors.Errorf("unexpected type %s found", reflect.TypeOf(d.Doc).String())
	}
	return nil
}

type TestInterface interface {
	Ping()
}

type TestInterfaceImplA struct {
	Int int
}

func (a *TestInterfaceImplA) Ping() {}

type TestInterfaceImplB struct {
	Float float32
}

func (a *TestInterfaceImplB) Ping() {}

type TestDocWrapperProto struct {
	Type string
	Doc  TestInterfaceProto
}

type TestInterfaceProto interface {
	IsTestInterfaceProto() bool
}

type TestDocWrapperProtoImplA struct {
	Int int
}

func (a *TestDocWrapperProtoImplA) IsTestInterfaceProto() bool {
	return true
}

func (a *TestDocWrapperProtoImplB) IsTestInterfaceProto() bool {
	return true
}

type TestDocWrapperProtoImplB struct {
	Float float32
}

type TestStructA struct {
	Uint         uint
	StringString string
}

type TestStructB struct {
	Uint         uint
	StringString string
}

type TestMarshalMac struct {
	mac net.HardwareAddr
}

func (a TestMarshalMac) String() string {
	if len(a.mac) == 0 {
		return ""
	}
	return a.mac.String()
}

func (a *TestMarshalMac) MarshalText() ([]byte, error) {
	if a == nil {
		return nil, nil
	}
	return []byte(a.String()), nil
}

func (a *TestMarshalMac) UnmarshalText(text []byte) error {
	var m = TestMarshalMac{}
	str := string(text)
	if len(str) > 0 {
		parsedMac, err := MarshalMacParseMac(str)
		if err != nil {
			return err
		}
		m = parsedMac
	}
	*a = m
	return nil
}

func MarshalMacParseMac(s string) (mac TestMarshalMac, err error) {
	var netMac net.HardwareAddr
	netMac, err = net.ParseMAC(s)
	if err != nil {
		return
	}
	mac = MarshalMacMakeMacFromHardwareAddr(netMac)
	return
}

func MarshalMacMakeMacFromHardwareAddr(m net.HardwareAddr) TestMarshalMac {
	// make a deep copy
	macCopy := make([]byte, len(m))
	copy(macCopy, m)

	var mac = TestMarshalMac{}
	mac.mac = macCopy
	return mac
}

type HttpHeader http.Header

type ProtoHttpHeaderElement struct {
	Values []string
}

func (h HttpHeader) CopierMarshal() (Any, error) {
	protoHeaderMap := make(map[string]*ProtoHttpHeaderElement)
	for hkey, hvalue := range h {
		protoHeaderElement := &ProtoHttpHeaderElement{
			Values: hvalue,
		}
		protoHeaderMap[hkey] = protoHeaderElement
	}
	return protoHeaderMap, nil
}

var protoHeaderType = reflect.TypeOf(make(map[string]*ProtoHttpHeaderElement))

func (h *HttpHeader) CopierUnmarshal(a Any) error {
	protoHeaderMap, ok := a.(map[string]*ProtoHttpHeaderElement)
	if !ok {
		return errors.Errorf("expected type %s and found type %s", protoHeaderType.String(), reflect.TypeOf(a).String())
	}

	//if h == nil {
	//	newHttpHeader := make(HttpHeader)
	//	h = &newHttpHeader
	//}
	modelMap := *h
	//if modelMap == nil {
	//	modelMap = make(HttpHeader)
	//}

	for hkey, hvalue := range protoHeaderMap {
		if hvalue == nil {
			modelMap[hkey] = nil
			continue
		}
		modelMap[hkey] = hvalue.Values
	}
	return nil
}
