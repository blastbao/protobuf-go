// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package proto

import (
	"bytes"
	"math"
	"reflect"

	"google.golang.org/protobuf/encoding/protowire"
	pref "google.golang.org/protobuf/reflect/protoreflect"
)

// Equal reports whether two messages are equal.
// If two messages marshal to the same bytes under deterministic serialization,
// then Equal is guaranteed to report true.
//
// Two messages are equal if they belong to the same message descriptor,
// have the same set of populated known and extension field values,
// and the same set of unknown fields values. If either of the top-level
// messages are invalid, then Equal reports true only if both are invalid.
//
// Scalar values are compared with the equivalent of the == operator in Go,
// except bytes values which are compared using bytes.Equal and
// floating point values which specially treat NaNs as equal.
// Message values are compared by recursively calling Equal.
// Lists are equal if each element value is also equal.
// Maps are equal if they have the same set of keys, where the pair of values
// for each key is also equal.
func Equal(x, y Message) bool {
	if x == nil || y == nil {
		return x == nil && y == nil
	}
	mx := x.ProtoReflect()
	my := y.ProtoReflect()
	if mx.IsValid() != my.IsValid() {
		return false
	}
	return equalMessage(mx, my)
}

// equalMessage compares two messages.
//
// message => field => value => message
func equalMessage(mx, my pref.Message) bool {
	// 必须是同一类型
	if mx.Descriptor() != my.Descriptor() {
		return false
	}

	// 遍历 mx 各个字段 <fx, vx>，取 my 中对应的 <fy, vy> ，比较是否相等，同时记录字段数目 nx 。
	nx := 0
	equal := true
	mx.Range(func(fd pref.FieldDescriptor, vx pref.Value) bool {
		nx++
		vy := my.Get(fd)
		equal = my.Has(fd) && equalField(fd, vx, vy)
		return equal
	})
	if !equal {
		return false
	}

	// 遍历 my 各个字段，统计字段总数。
	ny := 0
	my.Range(func(fd pref.FieldDescriptor, vx pref.Value) bool {
		ny++
		return true
	})

	// 字段数目不同，则不等
	if nx != ny {
		return false
	}

	return equalUnknown(mx.GetUnknown(), my.GetUnknown())
}

// equalField compares two fields.
func equalField(fd pref.FieldDescriptor, x, y pref.Value) bool {
	switch {
	// list
	case fd.IsList():
		return equalList(fd, x.List(), y.List())
	// map
	case fd.IsMap():
		return equalMap(fd, x.Map(), y.Map())
	// 值
	default:
		return equalValue(fd, x, y)
	}
}

// equalMap compares two maps.
func equalMap(fd pref.FieldDescriptor, x, y pref.Map) bool {
	// 数目
	if x.Len() != y.Len() {
		return false
	}
	// 字段
	equal := true
	x.Range(func(k pref.MapKey, vx pref.Value) bool {
		vy := y.Get(k)
		// 数值比较
		equal = y.Has(k) && equalValue(fd.MapValue(), vx, vy)
		return equal
	})
	return equal
}

// equalList compares two lists.
func equalList(fd pref.FieldDescriptor, x, y pref.List) bool {
	// 数目
	if x.Len() != y.Len() {
		return false
	}
	// 字段
	for i := x.Len() - 1; i >= 0; i-- {
		// 数值比较
		if !equalValue(fd, x.Get(i), y.Get(i)) {
			return false
		}
	}
	return true
}

// equalValue compares two singular values.
func equalValue(fd pref.FieldDescriptor, x, y pref.Value) bool {
	// 基础类型: bool/enum/int/float/string/bytes
	// 符合类型: message/interface

	switch fd.Kind() {
	case pref.BoolKind:
		return x.Bool() == y.Bool()
	case pref.EnumKind:
		return x.Enum() == y.Enum()
	case pref.Int32Kind, pref.Sint32Kind,
		pref.Int64Kind, pref.Sint64Kind,
		pref.Sfixed32Kind, pref.Sfixed64Kind:
		return x.Int() == y.Int()
	case pref.Uint32Kind, pref.Uint64Kind,
		pref.Fixed32Kind, pref.Fixed64Kind:
		return x.Uint() == y.Uint()
	case pref.FloatKind, pref.DoubleKind:
		fx := x.Float()
		fy := y.Float()
		if math.IsNaN(fx) || math.IsNaN(fy) {
			return math.IsNaN(fx) && math.IsNaN(fy)
		}
		return fx == fy
	case pref.StringKind:
		return x.String() == y.String()
	case pref.BytesKind:
		return bytes.Equal(x.Bytes(), y.Bytes())
	// [重要] 递归比较 message ...
	case pref.MessageKind, pref.GroupKind:
		return equalMessage(x.Message(), y.Message())
	default:
		return x.Interface() == y.Interface()
	}
}

// equalUnknown compares unknown fields by direct comparison on the raw bytes
// of each individual field number.
func equalUnknown(x, y pref.RawFields) bool {
	if len(x) != len(y) {
		return false
	}
	if bytes.Equal([]byte(x), []byte(y)) {
		return true
	}

	mx := make(map[pref.FieldNumber]pref.RawFields)
	my := make(map[pref.FieldNumber]pref.RawFields)
	for len(x) > 0 {
		fnum, _, n := protowire.ConsumeField(x)
		mx[fnum] = append(mx[fnum], x[:n]...)
		x = x[n:]
	}
	for len(y) > 0 {
		fnum, _, n := protowire.ConsumeField(y)
		my[fnum] = append(my[fnum], y[:n]...)
		y = y[n:]
	}
	return reflect.DeepEqual(mx, my)
}
