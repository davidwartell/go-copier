///
// Copyright (c) 2022. StealthMode Inc. All Rights Reserved
///

package copier

import (
	"encoding"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"reflect"
	"strings"
	"time"
)

//goland:noinspection SpellCheckingInspection
const (
	structTag = "copier"
)

type Any = interface{}

//goland:noinspection GoNameStartsWithPackageName
type CopierMarshaler interface {
	CopierMarshal() (Any, error)
}

//goland:noinspection GoNameStartsWithPackageName
type CopierUnmarshaler interface {
	CopierUnmarshal(Any) error
}

var (
	copyMarshaler          = reflect.TypeOf((*CopierMarshaler)(nil)).Elem()
	copyUnmarshaler        = reflect.TypeOf((*CopierUnmarshaler)(nil)).Elem()
	textMarshalerType      = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
	textUnmarshalerType    = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	protoStringWrapperType = reflect.TypeOf((*wrapperspb.StringValue)(nil)).Elem()
	protoBoolWrapperType   = reflect.TypeOf((*wrapperspb.BoolValue)(nil)).Elem()
	protoInt32WrapperType  = reflect.TypeOf((*wrapperspb.Int32Value)(nil)).Elem()
	protoInt64WrapperType  = reflect.TypeOf((*wrapperspb.Int64Value)(nil)).Elem()
	protoUInt32WrapperType = reflect.TypeOf((*wrapperspb.UInt32Value)(nil)).Elem()
	protoUInt64WrapperType = reflect.TypeOf((*wrapperspb.UInt64Value)(nil)).Elem()
	protoFloatWrapperType  = reflect.TypeOf((*wrapperspb.FloatValue)(nil)).Elem()
	protoDoubleWrapperType = reflect.TypeOf((*wrapperspb.DoubleValue)(nil)).Elem()
	protoBytesWrapperType  = reflect.TypeOf((*wrapperspb.BytesValue)(nil)).Elem()
	timeType               = reflect.TypeOf((*time.Time)(nil)).Elem()
	protoTimestampType     = reflect.TypeOf((*timestamppb.Timestamp)(nil)).Elem()
	durationType           = reflect.TypeOf((*time.Duration)(nil)).Elem()
	protoDurationType      = reflect.TypeOf((*durationpb.Duration)(nil)).Elem()
)

// Copy copies from src to dst. Only pass pointers not concrete values. Fields that can not be copied will be ignored.
func Copy(dst, src Any) (err error) {
	srcVal := reflect.Indirect(reflect.ValueOf(src))
	dstVal := reflect.Indirect(reflect.ValueOf(dst))
	if !dstVal.CanSet() {
		return errors.New("dst must be addressable (pass a pointer to a struct not a struct by value)")
	}
	err = copyStruct(dstVal, srcVal)
	if err != nil {
		return
	}
	return
}

type structField struct {
	value       reflect.Value
	structField reflect.StructField
}

func copyStruct(dstVal, srcVal reflect.Value) (err error) {
	if dstVal.Kind() != reflect.Struct {
		return errors.Errorf("expected dstVal type %s and received type %s", reflect.Struct.String(), dstVal.Kind().String())
	}
	if srcVal.Kind() != reflect.Struct {
		return errors.Errorf("expected srcVal type %s and received type %s", reflect.Struct.String(), srcVal.Kind().String())
	}

	// make a map of destination fields
	dstFieldMap := make(map[string]*structField)
	for j := 0; j < dstVal.NumField(); j++ {
		sf := new(structField)
		sf.value = dstVal.Field(j)
		sf.structField = dstVal.Type().Field(j)
		dstFieldName := sf.structField.Name
		protoTag := sf.structField.Tag.Get(structTag)
		if protoTag == "-" {
			continue
		} else if len(protoTag) > 0 {
			dstFieldName = protoTag
		}
		dstFieldMap[strings.ToLower(dstFieldName)] = sf
	}

	for i := 0; i < srcVal.NumField(); i++ {
		srcField := srcVal.Field(i)
		srcType := srcVal.Type().Field(i)
		srcFieldName := srcType.Name
		protoTag := srcType.Tag.Get(structTag)
		if protoTag == "-" {
			continue
		} else if len(protoTag) > 0 {
			srcFieldName = protoTag
		}

		//fmt.Printf("%s\n", srcFieldName)

		// look for a matching field name in dst
		matchingDstField, foundMatch := dstFieldMap[strings.ToLower(srcFieldName)]
		if !foundMatch {
			continue
		}

		// field names match now try to copy
		err = copyStructField(matchingDstField.value, matchingDstField.structField.Type, srcField, srcType.Type)
		if err != nil {
			return
		}
	}
	return
}

func copyStructField(dstField reflect.Value, dstType reflect.Type, srcField reflect.Value, srcType reflect.Type) (err error) {
	// skip if src is nil
	if (srcField.Kind() == reflect.Ptr ||
		srcField.Kind() == reflect.Interface ||
		srcField.Kind() == reflect.Map ||
		srcField.Kind() == reflect.Slice) &&
		srcField.IsNil() {
		return
	}

	if srcField.Kind() == reflect.Ptr && srcField.Elem().Kind() == reflect.Ptr {
		// srcField pointer to pointer recursion
		err = copyStructField(dstField, dstType, srcField.Elem(), srcField.Elem().Type())
		return
	}

	var marshalValOk bool

	marshalValOk, err = copyProtocopMarshaler(dstField, dstType, srcField, srcType)
	if marshalValOk || err != nil {
		return
	}

	marshalValOk, err = copyProtoTimestamp(dstField, dstType, srcField)
	if marshalValOk || err != nil {
		return
	}

	marshalValOk, err = copyProtoDuration(dstField, dstType, srcField)
	if marshalValOk || err != nil {
		return
	}

	marshalValOk, err = copyTextField(dstField, dstType, srcField, srcType)
	if marshalValOk || err != nil {
		return
	}

	marshalValOk, err = copyBool(dstField, dstType, srcField)
	if marshalValOk || err != nil {
		return
	}

	marshalValOk, err = copyNumeric(dstField, dstType, srcField)
	if marshalValOk || err != nil {
		return
	}

	marshalValOk, err = copyArray(dstField, dstType, srcField)
	if marshalValOk || err != nil {
		return
	}

	marshalValOk, err = copyMap(dstField, dstType, srcField)
	if marshalValOk || err != nil {
		return
	}

	if dstField.Kind() == reflect.Ptr && dstType.Elem().Kind() == reflect.Ptr {
		ptrTo := reflect.PtrTo(dstType.Elem()) // create a **T type.
		ptrVal := reflect.New(ptrTo.Elem())
		dstField.Set(ptrVal)
		err = copyStructField(dstField.Elem(), dstField.Type().Elem(), srcField, srcType)
		if err == nil {
			ptrVal.Elem().Set(dstField.Elem())
		}
		return
	}

	if srcField.Kind() == reflect.Struct {
		if dstField.Kind() == reflect.Pointer {
			if dstField.IsNil() {
				dstField.Set(reflect.New(dstType.Elem()))
			}
			return copyStructField(dstField.Elem(), dstField.Elem().Type(), srcField, srcType)
		} else {
			return copyStruct(dstField, srcField)
		}
	} else if srcField.Kind() == reflect.Ptr && srcField.Elem().Kind() == reflect.Struct {
		if dstField.Kind() == reflect.Pointer {
			if dstField.IsNil() {
				dstField.Set(reflect.New(dstType.Elem()))
			}
			return copyStructField(dstField.Elem(), dstField.Elem().Type(), srcField, srcType)
		} else {
			return copyStruct(dstField, srcField.Elem())
		}
	}

	return
}

func copyMap(dstField reflect.Value, dstType reflect.Type, srcField reflect.Value) (ok bool, err error) {
	srcFieldPtrVal := srcField
	if srcField.Kind() == reflect.Ptr {
		if srcField.IsNil() {
			return
		}
		srcFieldPtrVal = srcField.Elem()
	}
	if srcFieldPtrVal.Kind() != reflect.Map {
		return
	}
	if srcFieldPtrVal.Len() == 0 {
		return
	}

	if dstField.Kind() == reflect.Ptr && dstType.Elem().Kind() == reflect.Ptr {
		ptrTo := reflect.PtrTo(dstType.Elem()) // create a **T type.
		ptrVal := reflect.New(ptrTo.Elem())
		dstField.Set(ptrVal)

		ok, err = copyMap(dstField.Elem(), dstField.Type().Elem(), srcField)
		if err == nil && ok {
			ptrVal.Elem().Set(dstField.Elem())

		}
		return
	}

	var dstMapVal reflect.Value
	if dstField.Kind() == reflect.Ptr && !dstField.IsNil() {
		dstMapVal = dstField.Elem()
	} else if dstField.Kind() == reflect.Ptr && dstField.IsNil() && dstType.Elem().Kind() == reflect.Map {
		dstMapVal = reflect.MakeMap(dstType.Elem())
		dstField.Set(makePtr(dstMapVal))
	} else if dstField.Kind() == reflect.Map {
		dstField.Set(reflect.MakeMap(dstType))
		dstMapVal = dstField
	}

	// FIXME key and value must be the same but values can be different
	if !dstMapVal.Type().Key().AssignableTo(srcFieldPtrVal.Type().Key()) || !dstMapVal.Type().AssignableTo(srcFieldPtrVal.Type()) {
		return
	}

	mapRange := srcFieldPtrVal.MapRange()
	for i := 0; mapRange.Next(); i++ {
		dstMapVal.SetMapIndex(mapRange.Key(), mapRange.Value())
	}
	ok = true
	return
}

func copyArray(dstField reflect.Value, dstType reflect.Type, srcField reflect.Value) (ok bool, err error) {
	srcFieldPtrVal := srcField
	if srcField.Kind() == reflect.Ptr {
		if srcField.IsNil() {
			return
		}
		srcFieldPtrVal = srcField.Elem()
	}

	if srcFieldPtrVal.Kind() != reflect.Array && srcFieldPtrVal.Kind() != reflect.Slice && !srcFieldPtrVal.Type().AssignableTo(protoBytesWrapperType) {
		return
	}
	if (srcFieldPtrVal.Kind() == reflect.Array || srcFieldPtrVal.Kind() == reflect.Slice) && srcFieldPtrVal.Len() == 0 {
		return
	}

	if dstField.Kind() == reflect.Ptr && dstType.Elem().Kind() == reflect.Ptr {
		ptrTo := reflect.PtrTo(dstType.Elem()) // create a **T type.
		ptrVal := reflect.New(ptrTo.Elem())
		dstField.Set(ptrVal)

		ok, err = copyArray(dstField.Elem(), dstField.Type().Elem(), srcField)
		if err == nil && ok {
			ptrVal.Elem().Set(dstField.Elem())

		}
		return
	}

	var dstSliceVal reflect.Value
	if (srcFieldPtrVal.Kind() == reflect.Array || srcFieldPtrVal.Kind() == reflect.Slice) && srcFieldPtrVal.Type().Elem().Kind() == reflect.Uint8 && dstField.Kind() == reflect.Ptr && dstField.IsNil() && dstField.Type().Elem().AssignableTo(protoBytesWrapperType) {
		dstField.Set(reflect.New(dstType.Elem()))
		protoNumericValueFieldValue := dstField.Elem().FieldByName("Value")
		protoNumericValueFieldValue.SetBytes(srcFieldPtrVal.Bytes())
		ok = err == nil
		return
	} else if srcFieldPtrVal.Kind() == reflect.Struct && srcFieldPtrVal.Type().AssignableTo(protoBytesWrapperType) && dstField.Kind() == reflect.Ptr && dstField.IsNil() && dstField.Type().Elem().AssignableTo(protoBytesWrapperType) {
		dstField.Set(reflect.New(dstType.Elem()))
		protoNumericValueFieldValue := dstField.Elem().FieldByName("Value")
		srcProtoFieldValue := srcFieldPtrVal.FieldByName("Value")
		protoNumericValueFieldValue.SetBytes(srcProtoFieldValue.Bytes())
		ok = err == nil
		return
	} else if srcFieldPtrVal.Kind() == reflect.Struct && srcFieldPtrVal.Type().AssignableTo(protoBytesWrapperType) && dstField.Kind() == reflect.Slice && dstField.Type().Elem().Kind() == reflect.Uint8 {
		srcFieldPtrVal = srcFieldPtrVal.FieldByName("Value")
		dstField.Set(reflect.MakeSlice(reflect.SliceOf(dstType.Elem()), srcFieldPtrVal.Len(), srcFieldPtrVal.Len()))
		dstSliceVal = dstField
	} else if dstField.Kind() == reflect.Ptr && dstType.Elem().Kind() == reflect.Slice {
		dstSliceVal = reflect.MakeSlice(reflect.SliceOf(dstType.Elem().Elem()), srcFieldPtrVal.Len(), srcFieldPtrVal.Len())
		dstField.Set(makePtr(dstSliceVal))
	} else if dstField.Kind() == reflect.Slice {
		dstField.Set(reflect.MakeSlice(reflect.SliceOf(dstType.Elem()), srcFieldPtrVal.Len(), srcFieldPtrVal.Len()))
		dstSliceVal = dstField
	} else if dstField.Kind() == reflect.Ptr && !dstField.IsNil() {
		dstSliceVal = dstField.Elem()
	} else if dstField.Kind() == reflect.Array {
		dstSliceVal = dstField
	} else {
		return
	}

	if dstSliceVal.Len() < srcFieldPtrVal.Len() {
		err = errors.Errorf("invalid destination array/slice length for field type %s of length %d with src of length %d", dstSliceVal.Type().String(), dstSliceVal.Len(), srcFieldPtrVal.Len())
		return
	}

	for i := 0; i < srcFieldPtrVal.Len(); i++ {
		srcElement := srcFieldPtrVal.Index(i)
		if dstSliceVal.Index(i).Kind() == reflect.Ptr && dstSliceVal.Index(i).IsNil() {
			ptrVal := reflect.New(dstSliceVal.Type().Elem().Elem())
			dstSliceVal.Index(i).Set(ptrVal)
		}
		dstElement := dstSliceVal.Index(i)
		err = copyStructField(dstElement, dstElement.Type(), srcElement, srcElement.Type())
		if err != nil {
			break
		}
	}
	ok = true
	return
}

func copyProtoDuration(dstField reflect.Value, dstType reflect.Type, srcField reflect.Value) (ok bool, err error) {
	if srcField.Kind() == reflect.Ptr && srcField.IsNil() {
		return
	}

	if srcField.Kind() == reflect.Int64 && srcField.Type().AssignableTo(durationType) {
		srcFieldImpl := srcField.Interface().(time.Duration)
		ok, err = copyDurationInto(dstField, dstType, srcFieldImpl)
		return
	} else if srcField.Kind() == reflect.Ptr && srcField.Elem().Kind() == reflect.Int64 && srcField.Type().Elem().AssignableTo(durationType) {
		srcFieldImpl := srcField.Interface().(*time.Duration)
		ok, err = copyDurationInto(dstField, dstType, *srcFieldImpl)
		return
	} else if srcField.Kind() == reflect.Ptr && srcField.CanInterface() && srcField.Type().Elem().AssignableTo(protoDurationType) {
		srcFieldImpl := srcField.Interface().(*durationpb.Duration).AsDuration()
		ok, err = copyDurationInto(dstField, dstType, srcFieldImpl)
		return
	}
	return
}

func copyDurationInto(dstField reflect.Value, dstType reflect.Type, durationVal time.Duration) (ok bool, err error) {
	if !dstField.CanSet() {
		return
	}

	if dstField.Kind() == reflect.Ptr {
		if dstField.IsNil() {
			dstField.Set(reflect.New(dstType.Elem()))
		}
		if dstField.Elem().Kind() == reflect.Ptr {
			ok, err = copyDurationInto(dstField.Elem(), dstField.Elem().Type(), durationVal)
			if ok && err == nil {
				dstField.Set(makePtr(dstField.Elem()))
				ok = true
				return
			}
			return
		}
	}

	if dstField.Kind() == reflect.Int64 && dstField.Type().AssignableTo(durationType) {
		dstField.Set(reflect.ValueOf(durationVal))
		ok = true
		return
	} else if dstField.Kind() == reflect.Ptr && dstField.Elem().Kind() == reflect.Int64 && dstField.Type().Elem().AssignableTo(durationType) {
		dstField.Elem().Set(reflect.ValueOf(durationVal))
		ok = true
		return
	} else if dstField.Kind() == reflect.Ptr && dstField.Type().Elem().AssignableTo(protoDurationType) {
		dstField.Set(reflect.ValueOf(durationpb.New(durationVal)))
		ok = true
		return
	}
	return
}

func copyProtoTimestamp(dstField reflect.Value, dstType reflect.Type, srcField reflect.Value) (ok bool, err error) {
	if (srcField.Kind() == reflect.Ptr && srcField.IsNil()) || !srcField.CanInterface() {
		return
	}

	if srcField.Kind() == reflect.Struct && srcField.Type().AssignableTo(timeType) {
		srcFieldImpl := srcField.Interface().(time.Time)
		ok, err = copyTimestampInto(dstField, dstType, srcFieldImpl)
		return
	} else if srcField.Kind() == reflect.Ptr && srcField.Type().Elem().AssignableTo(timeType) {
		srcFieldImpl := srcField.Interface().(*time.Time)
		ok, err = copyTimestampInto(dstField, dstType, *srcFieldImpl)
		return
	} else if srcField.Kind() == reflect.Ptr && srcField.Type().Elem().AssignableTo(protoTimestampType) {
		srcFieldImpl := srcField.Interface().(*timestamppb.Timestamp).AsTime()
		ok, err = copyTimestampInto(dstField, dstType, srcFieldImpl)
		return
	}
	return
}

func copyTimestampInto(dstField reflect.Value, dstType reflect.Type, timeVal time.Time) (ok bool, err error) {
	if !dstField.CanSet() {
		return
	}

	if dstField.Kind() == reflect.Ptr {
		if dstField.IsNil() {
			dstField.Set(reflect.New(dstType.Elem()))
		}
		if dstField.Elem().Kind() == reflect.Ptr {
			ok, err = copyTimestampInto(dstField.Elem(), dstField.Elem().Type(), timeVal)
			if ok && err == nil {
				dstField.Set(makePtr(dstField.Elem()))
				ok = true
				return
			}
			return
		}
	}

	if dstField.Kind() == reflect.Struct && dstField.Type().AssignableTo(timeType) {
		dstField.Set(reflect.ValueOf(timeVal))
		ok = true
		return
	} else if dstField.Kind() == reflect.Ptr && dstField.Type().Elem().AssignableTo(timeType) {
		dstField.Elem().Set(reflect.ValueOf(timeVal))
		ok = true
		return
	} else if dstField.Kind() == reflect.Ptr && dstField.Type().Elem().AssignableTo(protoTimestampType) {
		dstField.Set(reflect.ValueOf(timestamppb.New(timeVal)))
		ok = true
		return
	}
	return
}

func copyNumeric(dstField reflect.Value, dstType reflect.Type, srcField reflect.Value) (ok bool, err error) {
	srcFieldPtrVal := srcField
	if srcField.Kind() == reflect.Ptr {
		srcFieldPtrVal = srcField.Elem()
		switch {
		case srcFieldPtrVal.Type().AssignableTo(protoInt32WrapperType), srcFieldPtrVal.Type().AssignableTo(protoInt64WrapperType):
			protoNumericValueFieldValue := srcFieldPtrVal.FieldByName("Value")
			numVal := protoNumericValueFieldValue.Int()
			err = copyIntInto(dstField, dstType, numVal)
			ok = err == nil
			return

		case srcFieldPtrVal.Type().AssignableTo(protoUInt32WrapperType), srcFieldPtrVal.Type().AssignableTo(protoUInt64WrapperType):
			protoNumericValueFieldValue := srcFieldPtrVal.FieldByName("Value")
			numVal := protoNumericValueFieldValue.Uint()
			err = copyUintInto(dstField, dstType, numVal)
			ok = err == nil
			return

		case srcFieldPtrVal.Type().AssignableTo(protoFloatWrapperType), srcFieldPtrVal.Type().AssignableTo(protoDoubleWrapperType):
			protoNumericValueFieldValue := srcFieldPtrVal.FieldByName("Value")
			numVal := protoNumericValueFieldValue.Float()
			err = copyFloatInto(dstField, dstType, numVal)
			ok = err == nil
			return
		}
	}
	switch srcFieldPtrVal.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		numVal := srcFieldPtrVal.Int()
		err = copyIntInto(dstField, dstType, numVal)
		ok = err == nil
		return

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		numVal := srcFieldPtrVal.Uint()
		err = copyUintInto(dstField, dstType, numVal)
		ok = err == nil
		return

	case reflect.Float32, reflect.Float64:
		numVal := srcFieldPtrVal.Float()
		err = copyFloatInto(dstField, dstType, numVal)
		ok = err == nil
		return
	}
	return
}

func copyFloatInto(dstField reflect.Value, dstType reflect.Type, numVal float64) (err error) {
	if !dstField.CanSet() {
		return
	}
	dstFieldAddrVal := dstField
	if dstField.Kind() == reflect.Ptr {
		if dstField.IsNil() {
			dstField.Set(reflect.New(dstType.Elem()))
		}

		if dstField.Elem().Kind() == reflect.Ptr {
			err = copyFloatInto(dstField.Elem(), dstField.Elem().Type(), numVal)
			if err == nil {
				dstField.Set(makePtr(dstField.Elem()))
			}
			return
		}

		dstFieldAddrVal = dstField.Elem()

		if dstFieldAddrVal.Type().AssignableTo(protoFloatWrapperType) || dstFieldAddrVal.Type().AssignableTo(protoDoubleWrapperType) {
			protoNumValueFieldValue := dstFieldAddrVal.FieldByName("Value")
			protoNumValueFieldValue.SetFloat(numVal)
			return
		}
	}

	switch dstFieldAddrVal.Kind() {
	case reflect.Float32, reflect.Float64:
		dstFieldAddrVal.SetFloat(numVal)
		return
	}

	return
}

func copyUintInto(dstField reflect.Value, dstType reflect.Type, numVal uint64) (err error) {
	if !dstField.CanSet() {
		return
	}
	dstFieldAddrVal := dstField
	if dstField.Kind() == reflect.Ptr {
		if dstField.IsNil() {
			dstField.Set(reflect.New(dstType.Elem()))
		}

		if dstField.Elem().Kind() == reflect.Ptr {
			err = copyUintInto(dstField.Elem(), dstField.Elem().Type(), numVal)
			if err == nil {
				dstField.Set(makePtr(dstField.Elem()))
			}
			return
		}

		dstFieldAddrVal = dstField.Elem()

		if dstFieldAddrVal.Type().AssignableTo(protoUInt32WrapperType) || dstFieldAddrVal.Type().AssignableTo(protoUInt64WrapperType) {
			protoNumValueFieldValue := dstFieldAddrVal.FieldByName("Value")
			protoNumValueFieldValue.SetUint(numVal)
			return
		}
	}

	switch dstFieldAddrVal.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		dstFieldAddrVal.SetUint(numVal)
		return
	}

	return
}

func copyIntInto(dstField reflect.Value, dstType reflect.Type, numVal int64) (err error) {
	if !dstField.CanSet() {
		return
	}
	dstFieldAddrVal := dstField
	if dstField.Kind() == reflect.Ptr {
		if dstField.IsNil() {
			dstField.Set(reflect.New(dstType.Elem()))
		}

		if dstField.Elem().Kind() == reflect.Ptr {
			err = copyIntInto(dstField.Elem(), dstField.Elem().Type(), numVal)
			if err == nil {
				dstField.Set(makePtr(dstField.Elem()))
			}
			return
		}

		dstFieldAddrVal = dstField.Elem()

		if dstFieldAddrVal.Type().AssignableTo(protoInt32WrapperType) || dstFieldAddrVal.Type().AssignableTo(protoInt64WrapperType) {
			protoNumValueFieldValue := dstFieldAddrVal.FieldByName("Value")
			protoNumValueFieldValue.SetInt(numVal)
		}
	}

	switch dstFieldAddrVal.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		dstFieldAddrVal.SetInt(numVal)
		return
	}

	return
}

func copyBool(dstField reflect.Value, dstType reflect.Type, srcField reflect.Value) (ok bool, err error) {
	if srcField.Kind() == reflect.Bool {
		boolVal := srcField.Bool()
		err = copyBoolInto(dstField, dstType, boolVal)
		ok = err == nil
		return
	} else if srcField.Kind() == reflect.Ptr && srcField.Elem().Kind() == reflect.Bool {
		boolVal := srcField.Elem().Bool()
		err = copyBoolInto(dstField, dstType, boolVal)
		ok = err == nil
		return
	} else if srcField.Kind() == reflect.Ptr && srcField.Elem().Type().AssignableTo(protoBoolWrapperType) {
		protoBoolValueFieldValue := srcField.Elem().FieldByName("Value")
		if !protoBoolValueFieldValue.IsZero() {
			boolVal := protoBoolValueFieldValue.Bool()
			err = copyBoolInto(dstField, dstType, boolVal)
			ok = err == nil
			return
		}
	}
	return
}

func copyBoolInto(dstField reflect.Value, dstType reflect.Type, boolVal bool) (err error) {
	if !dstField.CanSet() {
		return
	}
	if dstField.Kind() == reflect.Ptr && dstField.IsNil() {
		dstField.Set(reflect.New(dstType.Elem()))
	}

	if dstField.Kind() == reflect.Ptr && dstField.Elem().Kind() == reflect.Ptr {
		err = copyBoolInto(dstField.Elem(), dstField.Elem().Type(), boolVal)
		if err == nil {
			dstField.Set(makePtr(dstField.Elem()))
		}
		return
	}

	dstFieldAddrVal := dstField
	if dstField.Kind() != reflect.Ptr {
		dstFieldAddrVal = dstField.Addr()
		if dstFieldAddrVal.IsNil() {
			return
		}
	}

	if dstFieldAddrVal.Elem().Type().AssignableTo(protoBoolWrapperType) {
		protoBoolValueFieldValue := dstFieldAddrVal.Elem().FieldByName("Value")
		protoBoolValueFieldValue.SetBool(boolVal)
		return
	}

	if dstField.Kind() == reflect.Bool {
		dstField.SetBool(boolVal)
		return
	}
	if dstField.Kind() == reflect.Ptr && dstField.Elem().Kind() == reflect.Bool {
		dstField.Elem().SetBool(boolVal)
		return
	}
	return
}

func copyProtocopMarshaler(dstField reflect.Value, dstType reflect.Type, srcField reflect.Value, srcType reflect.Type) (ok bool, err error) {
	var srcFieldAddrVal reflect.Value
	if srcField.Kind() != reflect.Ptr && reflect.PtrTo(srcType).Implements(copyMarshaler) {
		srcFieldAddrVal = srcField.Addr()
	} else {
		srcFieldAddrVal = srcField
	}

	var dstFieldAddrVal reflect.Value
	if dstField.Kind() != reflect.Ptr && reflect.PtrTo(dstType).Implements(copyUnmarshaler) {
		dstFieldAddrVal = dstField.Addr()
	} else {
		dstFieldAddrVal = dstField
	}

	//fmt.Printf("\tdstFieldAddrVal: %s (%v, %v) srcFieldAddrVal: %s (%v, %v) %v \n", dstFieldAddrVal.Type().String(), dstFieldAddrVal.Type().Implements(copyMarshaler), dstFieldAddrVal.Type().Implements(copyUnmarshaler), srcFieldAddrVal.Type().String(), srcFieldAddrVal.Type().Implements(copyMarshaler), srcFieldAddrVal.Type().Implements(copyUnmarshaler), srcFieldAddrVal.CanInterface())

	if (srcFieldAddrVal.Kind() != reflect.Ptr && srcFieldAddrVal.Kind() != reflect.Map) || !srcFieldAddrVal.CanInterface() || srcFieldAddrVal.IsNil() {
		return
	}

	if srcFieldAddrVal.Type().Implements(copyMarshaler) {
		var marshaledVal reflect.Value
		sm, sok := srcFieldAddrVal.Interface().(CopierMarshaler)
		if !sok {
			return
		}
		var marshaled Any
		marshaled, err = sm.CopierMarshal()
		if err != nil {
			return
		}
		marshaledVal = reflect.ValueOf(marshaled)
		if !marshaledVal.Type().AssignableTo(dstFieldAddrVal.Type()) {
			return
		}

		dstFieldAddrVal.Set(marshaledVal)
		ok = true
		return
	} else if dstFieldAddrVal.Type().Implements(copyUnmarshaler) {
		if dstFieldAddrVal.Kind() == reflect.Ptr && dstFieldAddrVal.Elem().Kind() == reflect.Map && dstFieldAddrVal.Elem().IsNil() {
			dstFieldAddrVal.Elem().Set(reflect.MakeMap(dstFieldAddrVal.Type().Elem()))
		} else if dstFieldAddrVal.Kind() == reflect.Ptr && dstFieldAddrVal.IsNil() {
			//fmt.Printf("dstFieldAddrVal %s\n", dstFieldAddrVal.Type().String())
			dstFieldAddrVal.Set(reflect.New(dstFieldAddrVal.Type().Elem()))
		}

		sm, sok := dstFieldAddrVal.Interface().(CopierUnmarshaler)
		if !sok {
			return
		}
		err = sm.CopierUnmarshal(srcFieldAddrVal.Interface())
		if err != nil {
			return
		}
		ok = true
		return
	}
	return
}

func copyTextField(dstField reflect.Value, dstType reflect.Type, srcField reflect.Value, srcType reflect.Type) (ok bool, err error) {
	if srcField.Kind() == reflect.String {
		textVal := srcField.String()
		err = copyTextInto(dstField, dstType, []byte(textVal))
		ok = err == nil
		return
	} else if srcField.Kind() == reflect.Ptr && srcField.Elem().Kind() == reflect.String {
		textVal := srcField.Elem().String()
		err = copyTextInto(dstField, dstType, []byte(textVal))
		ok = err == nil
		return
	} else if srcField.Kind() != reflect.Ptr && reflect.PtrTo(srcType).Implements(textMarshalerType) {
		srcFieldAddrVal := srcField.Addr()
		if srcFieldAddrVal.IsNil() || !srcFieldAddrVal.CanInterface() {
			return
		}
		sm, sok := srcFieldAddrVal.Interface().(encoding.TextMarshaler)
		if !sok {
			return
		}
		var textVal []byte
		textVal, err = sm.MarshalText()
		if err != nil {
			return
		}
		err = copyTextInto(dstField, dstType, textVal)
		ok = err == nil
		return
	} else if srcField.Kind() != reflect.Ptr && srcType.Implements(textMarshalerType) {
		srcFieldAddrVal := srcField
		if srcFieldAddrVal.IsNil() || !srcFieldAddrVal.CanInterface() {
			return
		}
		sm, sok := srcFieldAddrVal.Interface().(encoding.TextMarshaler)
		if !sok {
			return
		}
		var textVal []byte
		textVal, err = sm.MarshalText()
		if err != nil {
			return
		}
		err = copyTextInto(dstField, dstType, textVal)
		ok = err == nil
		return
	} else if srcField.Kind() == reflect.Ptr && srcType.Implements(textMarshalerType) {
		if srcField.IsNil() {
			return
		}
		sm, sok := srcField.Interface().(encoding.TextMarshaler)
		if !sok {
			return
		}
		var textVal []byte
		textVal, err = sm.MarshalText()
		if err != nil {
			return
		}
		err = copyTextInto(dstField, dstType, textVal)
		ok = err == nil
		return
	} else if srcField.Kind() == reflect.Ptr && srcField.Elem().Type().AssignableTo(protoStringWrapperType) {
		protoStrValueFieldValue := srcField.Elem().FieldByName("Value")
		if !protoStrValueFieldValue.IsZero() {
			textVal := protoStrValueFieldValue.String()
			err = copyTextInto(dstField, dstType, []byte(textVal))
			ok = err == nil
			return
		}
	}
	return
}

func copyTextInto(dstField reflect.Value, dstType reflect.Type, textVal []byte) (err error) {
	if !dstField.CanSet() {
		return
	}
	if len(textVal) == 0 {
		return
	}

	if dstField.Kind() == reflect.Ptr && dstField.IsNil() {
		dstField.Set(reflect.New(dstType.Elem()))
	}

	if dstField.Kind() == reflect.Ptr && dstField.Elem().Kind() == reflect.Ptr {
		err = copyTextInto(dstField.Elem(), dstField.Elem().Type(), textVal)
		if err == nil {
			dstField.Set(makePtr(dstField.Elem()))
		}
		return
	}

	dstFieldAddrVal := dstField
	if dstField.Kind() != reflect.Ptr {
		dstFieldAddrVal = dstField.Addr()
		if dstFieldAddrVal.IsNil() {
			return
		}
	}

	if dstFieldAddrVal.Type().Implements(textUnmarshalerType) {
		dm, dok := dstFieldAddrVal.Interface().(encoding.TextUnmarshaler)
		if dok {
			err = dm.UnmarshalText(textVal)
			return
		}
	} else if dstFieldAddrVal.Elem().Type().AssignableTo(protoStringWrapperType) {
		protoStrValueFieldValue := dstFieldAddrVal.Elem().FieldByName("Value")
		protoStrValueFieldValue.SetString(string(textVal))
		return
	} else if dstFieldAddrVal.Kind() == reflect.Ptr && dstFieldAddrVal.Elem().Kind() == reflect.String {
		dstFieldAddrVal.Elem().SetString(string(textVal))
		return
	}
	return
}

// makePtr wraps val with a pointer: val.Type() => *val.Type()
func makePtr(val reflect.Value) reflect.Value {
	ptrTo := reflect.PtrTo(val.Type())  // create a *T type.
	ptrVal := reflect.New(ptrTo.Elem()) // create a reflect.Value of type *T.
	ptrVal.Elem().Set(val)              // sets ptrVal to point to underlying value of val.
	return ptrVal
}
