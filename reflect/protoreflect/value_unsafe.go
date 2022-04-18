// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !purego && !appengine
// +build !purego,!appengine

package protoreflect

import (
	"unsafe"

	"google.golang.org/protobuf/internal/pragma"
)

type (
	stringHeader struct {
		Data unsafe.Pointer
		Len  int
	}
	sliceHeader struct {
		Data unsafe.Pointer
		Len  int
		Cap  int
	}
	ifaceHeader struct {
		Type unsafe.Pointer
		Data unsafe.Pointer
	}
)

var (
	nilType     = typeOf(nil)
	boolType    = typeOf(*new(bool))
	int32Type   = typeOf(*new(int32))
	int64Type   = typeOf(*new(int64))
	uint32Type  = typeOf(*new(uint32))
	uint64Type  = typeOf(*new(uint64))
	float32Type = typeOf(*new(float32))
	float64Type = typeOf(*new(float64))
	stringType  = typeOf(*new(string))
	bytesType   = typeOf(*new([]byte))
	enumType    = typeOf(*new(EnumNumber))
)

// typeOf returns a pointer to the Go type information.
// The pointer is comparable and equal if and only if the types are identical.
func typeOf(t interface{}) unsafe.Pointer {
	return (*ifaceHeader)(unsafe.Pointer(&t)).Type
}

// value is a union where only one type can be represented at a time.
// The struct is 24B large on 64-bit systems and requires the minimum storage
// necessary to represent each possible type.
//
// The Go GC needs to be able to scan variables containing pointers.
// As such, pointers and non-pointers cannot be intermixed.
type value struct {
	pragma.DoNotCompare // 0B

	// typ stores the type of the value as a pointer to the Go type.
	// 类型
	typ unsafe.Pointer // 8B

	// ptr stores the data pointer for a String, Bytes, or interface value.
	// 数据
	ptr unsafe.Pointer // 8B

	// num stores a Bool, Int32, Int64, Uint32, Uint64, Float32, Float64, or
	// Enum value as a raw uint64.
	//
	// It is also used to store the length of a String or Bytes value;
	// the capacity is ignored.
	//
	// 对于数值类型，保存到 num 上，而非 ptr 。
	num uint64 // 8B
}

func valueOfString(v string) Value {
	// 将 string 转换为 *stringHeader 指针
	p := (*stringHeader)(unsafe.Pointer(&v))
	// 构造 Value
	return Value{typ: stringType, ptr: p.Data, num: uint64(len(v))}
}
func valueOfBytes(v []byte) Value {
	p := (*sliceHeader)(unsafe.Pointer(&v))
	return Value{typ: bytesType, ptr: p.Data, num: uint64(len(v))}
}

func valueOfIface(v interface{}) Value {
	// 解析 interface{} 头
	p := (*ifaceHeader)(unsafe.Pointer(&v))
	// 转换成 pointer
	return Value{
		typ: p.Type,	// 类型
		ptr: p.Data,	// 数据
	}
}

func (v Value) getString() (x string) {

	// [原理]
	// v.ptr 指向底层 string 数据，v.num 存储了 string 的长度。
	// 所以 stringHeader{ Data: v.ptr, Len: v.num } 就代表了 go 中的 String 数据。

	// 这里先取 x 的地址，转换成 pointer ，然后转换为 stringHeader ，最后通过赋值来实现返回。
	*(*stringHeader)(unsafe.Pointer(&x)) = stringHeader{
		Data: v.ptr,
		Len: int(v.num),
	}
	return x
}
func (v Value) getBytes() (x []byte) {

	// [原理]
	// v.ptr 指向底层 []byte 数据，v.num 存储了 []byte] 的长度。
	// 所以 sliceHeader{ Data: v.ptr, Len: v.num , Cap: v.num } 就代表了 go 中的 Slice 数据。

	*(*sliceHeader)(unsafe.Pointer(&x)) = sliceHeader{Data: v.ptr, Len: int(v.num), Cap: int(v.num)}
	return x
}
func (v Value) getIface() (x interface{}) {
	*(*ifaceHeader)(unsafe.Pointer(&x)) = ifaceHeader{Type: v.typ, Data: v.ptr}
	return x
}
