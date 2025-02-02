// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package impl

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"google.golang.org/protobuf/internal/genid"
	"google.golang.org/protobuf/reflect/protoreflect"
	pref "google.golang.org/protobuf/reflect/protoreflect"
	preg "google.golang.org/protobuf/reflect/protoregistry"
)

// MessageInfo provides protobuf related functionality for a given Go type
// that represents a message. A given instance of MessageInfo is tied to
// exactly one Go type, which must be a pointer to a struct type.
//
// The exported fields must be populated before any methods are called
// and cannot be mutated after set.
//
// 对于给定的 go 类型，MessageInfo 提供了 protobuf 相关功能。
// 一个给定的 MessageInfo 实例正好与一个 Go 类型相关联，该 Go 类型必须是一个指向结构类型的指针。
//
// 导出字段必须在调用任何方法之前填充，并且在设置之后不能被改变。???
type MessageInfo struct {
	// GoReflectType is the underlying message Go type and must be populated.
	// 底层的 Go 类型
	GoReflectType reflect.Type // pointer to struct

	// Desc is the underlying message descriptor type and must be populated.
	// 底层消息描述符
	Desc pref.MessageDescriptor

	// Exporter must be provided in a purego environment in order to provide
	// access to unexported fields.
	// 访问非导出字段
	Exporter exporter

	// OneofWrappers is list of pointers to oneof wrapper struct types.
	OneofWrappers []interface{}

	initMu   sync.Mutex // protects all unexported fields
	initDone uint32

	// 用于反射操作
	reflectMessageInfo // for reflection implementation
	// 用于快速操作
	coderMessageInfo   // for fast-path method implementations
}

// exporter is a function that returns a reference to the ith field of v,
// where v is a pointer to a struct. It returns nil if it does not support
// exporting the requested field (e.g., already exported).
type exporter func(v interface{}, i int) interface{}

// getMessageInfo returns the MessageInfo for any message type that
// is generated by our implementation of protoc-gen-go (for v2 and on).
// If it is unable to obtain a MessageInfo, it returns nil.
func getMessageInfo(mt reflect.Type) *MessageInfo {
	m, ok := reflect.Zero(mt).Interface().(pref.ProtoMessage)
	if !ok {
		return nil
	}
	mr, ok := m.ProtoReflect().(interface{ ProtoMessageInfo() *MessageInfo })
	if !ok {
		return nil
	}
	return mr.ProtoMessageInfo()
}

func (mi *MessageInfo) init() {
	// This function is called in the hot path. Inline the sync.Once logic,
	// since allocating a closure for Once.Do is expensive.
	// Keep init small to ensure that it can be inlined.
	//
	// 本函数在热路径中，这里没有使用 Once.Do 而是内联实现了 sync.Once 的逻辑，因为 Once.Do 分配一个闭包很昂贵。
	if atomic.LoadUint32(&mi.initDone) == 0 {
		mi.initOnce()
	}
}

func (mi *MessageInfo) initOnce() {
	mi.initMu.Lock()
	defer mi.initMu.Unlock()
	if mi.initDone == 1 {
		return
	}

	// 底层 Go 类型必须是结构体指针
	t := mi.GoReflectType
	if t.Kind() != reflect.Ptr && t.Elem().Kind() != reflect.Struct {
		panic(fmt.Sprintf("got %v, want *struct kind", t))
	}

	// 结构体类型
	t = t.Elem()

	// 把 t 解析并转换为 structInfo 对象
	si := mi.makeStructInfo(t)

	// 构造反射相关类型信息及函数
	mi.makeReflectFuncs(t, si)
	// 构造快速操作相关函数
	mi.makeCoderMethods(t, si)

	// 完成初始化
	atomic.StoreUint32(&mi.initDone, 1)
}

// getPointer returns the pointer for a message, which should be of
// the type of the MessageInfo. If the message is of a different type,
// it returns ok==false.
func (mi *MessageInfo) getPointer(m pref.Message) (p pointer, ok bool) {
	switch m := m.(type) {
	case *messageState:
		return m.pointer(), m.messageInfo() == mi
	case *messageReflectWrapper:
		return m.pointer(), m.messageInfo() == mi
	}
	return pointer{}, false
}

type (
	SizeCache       = int32
	WeakFields      = map[int32]protoreflect.ProtoMessage
	UnknownFields   = unknownFieldsA // TODO: switch to unknownFieldsB
	unknownFieldsA  = []byte
	unknownFieldsB  = *[]byte
	ExtensionFields = map[int32]ExtensionField
)

var (
	sizecacheType       = reflect.TypeOf(SizeCache(0))
	weakFieldsType      = reflect.TypeOf(WeakFields(nil))
	unknownFieldsAType  = reflect.TypeOf(unknownFieldsA(nil))
	unknownFieldsBType  = reflect.TypeOf(unknownFieldsB(nil))
	extensionFieldsType = reflect.TypeOf(ExtensionFields(nil))
)

type structInfo struct {
	sizecacheOffset offset			// "sizeCache"、"XXX_sizecache" 字段在 Message 结构体中的偏移量
	sizecacheType   reflect.Type	// "sizeCache"、"XXX_sizecache" 字段的 Go 类型
	weakOffset      offset
	weakType        reflect.Type
	unknownOffset   offset
	unknownType     reflect.Type
	extensionOffset offset
	extensionType   reflect.Type

	// 保存字段 Number 与 reflect.StructField 的映射。
	fieldsByNumber        map[pref.FieldNumber]reflect.StructField
	// 保存 OneOf 字段 Name 与 reflect.StructField 的映射。
	oneofsByName          map[pref.Name]reflect.StructField

	oneofWrappersByType   map[reflect.Type]pref.FieldNumber
	oneofWrappersByNumber map[pref.FieldNumber]reflect.Type
}

// 把 reflect.Type 解析并转换为 structInfo 对象。
func (mi *MessageInfo) makeStructInfo(t reflect.Type) structInfo {

	si := structInfo{
		sizecacheOffset: invalidOffset,
		weakOffset:      invalidOffset,
		unknownOffset:   invalidOffset,
		extensionOffset: invalidOffset,

		fieldsByNumber:        map[pref.FieldNumber]reflect.StructField{},
		oneofsByName:          map[pref.Name]reflect.StructField{},
		oneofWrappersByType:   map[reflect.Type]pref.FieldNumber{},
		oneofWrappersByNumber: map[pref.FieldNumber]reflect.Type{},
	}

fieldLoop:

	// 遍历结构体各个字段
	for i := 0; i < t.NumField(); i++ {
		// 根据字段名称进行特殊处理
		switch f := t.Field(i); f.Name {
		// "sizeCache"、"XXX_sizecache"
		case genid.SizeCache_goname, genid.SizeCacheA_goname:
			if f.Type == sizecacheType {
				si.sizecacheOffset = offsetOf(f, mi.Exporter) 	// 当前字段在结构体中的偏移量
				si.sizecacheType = f.Type						// 当前字段的反射类型
			}
		// "weakFields"、"XXX_weak"
		case genid.WeakFields_goname, genid.WeakFieldsA_goname:
			if f.Type == weakFieldsType {
				si.weakOffset = offsetOf(f, mi.Exporter)
				si.weakType = f.Type
			}
		// "unknownFields"、"XXX_unrecognized"
		case genid.UnknownFields_goname, genid.UnknownFieldsA_goname:
			if f.Type == unknownFieldsAType || f.Type == unknownFieldsBType {
				si.unknownOffset = offsetOf(f, mi.Exporter)
				si.unknownType = f.Type
			}
		// "extensionFields"、"XXX_InternalExtensions"、"XXX_extensions"
		case genid.ExtensionFields_goname, genid.ExtensionFieldsA_goname, genid.ExtensionFieldsB_goname:
			if f.Type == extensionFieldsType {
				si.extensionOffset = offsetOf(f, mi.Exporter)
				si.extensionType = f.Type
			}
		default:
			// eg.
			//
			// type TestNoEnforceUTF8 struct {
			//	OptionalString string       `protobuf:"bytes,1,opt,name=optional_string"`
			//	OptionalBytes  []byte       `protobuf:"bytes,2,opt,name=optional_bytes"`
			//	RepeatedString []string     `protobuf:"bytes,3,rep,name=repeated_string"`
			//	RepeatedBytes  [][]byte     `protobuf:"bytes,4,rep,name=repeated_bytes"`
			//	OneofField     isOneofField `protobuf_oneof:"oneof_field"`
			// }

			// 遍历 `protobuf:"..."` 的 tags
			for _, s := range strings.Split(f.Tag.Get("protobuf"), ",") {
				// 检查是否为纯数字
				if len(s) > 0 && strings.Trim(s, "0123456789") == "" {
					// 数值转换后得到字段的 Number ，然后保存它与 field 的映射。
					n, _ := strconv.ParseUint(s, 10, 64)
					si.fieldsByNumber[pref.FieldNumber(n)] = f
					continue fieldLoop
				}
			}

			// 遍历 `protobuf_oneof:"..."` 的 tags
			if s := f.Tag.Get("protobuf_oneof"); len(s) > 0 {
				si.oneofsByName[pref.Name(s)] = f
				continue fieldLoop
			}
		}
	}

	// Derive a mapping of oneof wrappers to fields.
	oneofWrappers := mi.OneofWrappers
	for _, method := range []string{"XXX_OneofFuncs", "XXX_OneofWrappers"} {
		if fn, ok := reflect.PtrTo(t).MethodByName(method); ok {
			for _, v := range fn.Func.Call([]reflect.Value{reflect.Zero(fn.Type.In(0))}) {
				if vs, ok := v.Interface().([]interface{}); ok {
					oneofWrappers = vs
				}
			}
		}
	}

	for _, v := range oneofWrappers {
		tf := reflect.TypeOf(v).Elem()
		f := tf.Field(0)
		for _, s := range strings.Split(f.Tag.Get("protobuf"), ",") {
			if len(s) > 0 && strings.Trim(s, "0123456789") == "" {
				n, _ := strconv.ParseUint(s, 10, 64)
				si.oneofWrappersByType[tf] = pref.FieldNumber(n)
				si.oneofWrappersByNumber[pref.FieldNumber(n)] = tf
				break
			}
		}
	}

	return si
}

func (mi *MessageInfo) New() protoreflect.Message {
	return mi.MessageOf(reflect.New(mi.GoReflectType.Elem()).Interface())
}
func (mi *MessageInfo) Zero() protoreflect.Message {
	return mi.MessageOf(reflect.Zero(mi.GoReflectType).Interface())
}
func (mi *MessageInfo) Descriptor() protoreflect.MessageDescriptor {
	return mi.Desc
}
func (mi *MessageInfo) Enum(i int) protoreflect.EnumType {
	mi.init()
	fd := mi.Desc.Fields().Get(i)
	return Export{}.EnumTypeOf(mi.fieldTypes[fd.Number()])
}
func (mi *MessageInfo) Message(i int) protoreflect.MessageType {
	mi.init()
	fd := mi.Desc.Fields().Get(i)
	switch {
	case fd.IsWeak():
		mt, _ := preg.GlobalTypes.FindMessageByName(fd.Message().FullName())
		return mt
	case fd.IsMap():
		return mapEntryType{fd.Message(), mi.fieldTypes[fd.Number()]}
	default:
		return Export{}.MessageTypeOf(mi.fieldTypes[fd.Number()])
	}
}

type mapEntryType struct {
	desc    protoreflect.MessageDescriptor
	valType interface{} // zero value of enum or message type
}

func (mt mapEntryType) New() protoreflect.Message {
	return nil
}
func (mt mapEntryType) Zero() protoreflect.Message {
	return nil
}
func (mt mapEntryType) Descriptor() protoreflect.MessageDescriptor {
	return mt.desc
}
func (mt mapEntryType) Enum(i int) protoreflect.EnumType {
	fd := mt.desc.Fields().Get(i)
	if fd.Enum() == nil {
		return nil
	}
	return Export{}.EnumTypeOf(mt.valType)
}
func (mt mapEntryType) Message(i int) protoreflect.MessageType {
	fd := mt.desc.Fields().Get(i)
	if fd.Message() == nil {
		return nil
	}
	return Export{}.MessageTypeOf(mt.valType)
}
