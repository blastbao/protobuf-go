// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package proto

import (
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/internal/encoding/messageset"
	"google.golang.org/protobuf/internal/order"
	"google.golang.org/protobuf/internal/pragma"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoiface"
)

// Proto 文件中定义字段的顺序与最终编码结果的字段顺序无关，两者有可能相同也可能不同。
// 当消息被编码时，Protobuf 无法保证消息的顺序，消息的顺序可能随着版本或者不同的实现而变化。
// 任何 Protobuf 的实现都应该保证字段以任意顺序编码的结果都能被读取。
//
// 	- 序列化后的消息字段顺序是不稳定的。
//	- 对同一段字节流进行解码，不同实现或版本的 Protobuf 解码得到的结果不一定完全相同（bytes 层面）。
//	  只能保证相同版本相同实现的 Protobuf 对同一段字节流多次解码得到的结果相同。
//  - 有两条逻辑上相等消息，但是序列化之后的内容（bytes 层面）不相同，部分可能的原因有：
//		- 其中一条消息可能使用了较老版本的 protobuf，不能处理某些类型的字段，设为 unknown 。
//		- 用了不同语言实现的 Protobuf，并且以不同的顺序编码字段。
//		- 消息中的字段使用了不稳定的算法进行序列化。
//		- 某条消息中有 bytes 类型的字段，用于储存另一条消息使用 Protobuf 序列化的结果，而这个 bytes 使用了不同的 Protobuf 进行序列化。
//		- 使用了新版本的 Protobuf，序列化实现不同。
//		- 消息字段顺序不同。


// Protobuf 消息的结构为键/值对的集合，其中数字标签为键，相应的字段为值。
// 字段名称是供人类阅读的，但是 protoc 编译器的确使用了字段名称来生成特定于语言的对应名称。
// 例如，Protobuf IDL 中的 oddA 和 small 名称在 Go 结构中分别成为字段 OddA 和 Small。
//
// 键和它们的值都被编码，但是有一个重要的区别：
//	一些数字值具有固定大小的 32 或 64 位的编码，而其他数字（包括消息标签）则是 varint 编码的，位数取决于整数的绝对值。
//	例如，整数值 1 到 15 需要 8 位 varint 编码，而值 16 到 2047 需要 16 位。
//
// varint 编码在本质上与 UTF-8 编码类似（但细节不同），它偏爱较小的整数值而不是较大的整数值。
// 结果是，Protobuf 消息应该在字段中具有较小的整数值（如果可能），并且键数应尽可能少，但每个字段至少得有一个键。
//
// 下表 1 列出了 Protobuf 编码的要点：
//		编码  		示例类型 						长度
//		varint 		int32、uint32、int64、uint64		可变长度
//		fixed		fixed32、float、double			固定的 32/64 位长度
//		字节序列		string、bytes					序列长度
//
//
// 可见，未明确固定长度的整数类型是 varint 编码的；
// 因此，在 varint 类型中，例如 uint32（u 代表无符号），数字 32 描述了整数的范围（在这种情况下为 0 到 232 - 1），
// 而不是其位的大小，该位大小取决于值。
//
// 相比之下，对于固定长度类型（例如 fixed32 或 double），Protobuf 编码分别需要 32 位和 64 位。
//
// Protobuf 中的字符串是字节序列；因此，字段编码的大小就是字节序列的长度。
//
// 另一个高效的方法值得一提，如下示例，其中的 DataItems 消息由重复的 DataItem 实例组成：
//		message DataItems {
//  		repeated DataItem item = 1;
//		}
// repeated 表示 DataItem 实例是打包的：集合具有单个标签，在这里是 1 。
// 因此，具有重复的 DataItem 实例的 DataItems 消息比具有多个但单独的 DataItem 字段、每个字段都需要自己的标签的消息的效率更高。
//

// MarshalOptions configures the marshaler.
//
// Example usage:
//   b, err := MarshalOptions{Deterministic: true}.Marshal(m)
type MarshalOptions struct {
	pragma.NoUnkeyedLiterals

	// AllowPartial allows messages that have missing required fields to marshal
	// without returning an error. If AllowPartial is false (the default),
	// Marshal will return an error if there are any missing required fields.
	AllowPartial bool

	// Deterministic controls whether the same message will always be
	// serialized to the same bytes within the same binary.
	//
	// Setting this option guarantees that repeated serialization of
	// the same message will return the same bytes, and that different
	// processes of the same binary (which may be executing on different
	// machines) will serialize equal messages to the same bytes.
	// It has no effect on the resulting size of the encoded message compared
	// to a non-deterministic marshal.
	//
	// Note that the deterministic serialization is NOT canonical across
	// languages. It is not guaranteed to remain stable over time. It is
	// unstable across different builds with schema changes due to unknown
	// fields. Users who need canonical serialization (e.g., persistent
	// storage in a canonical form, fingerprinting, etc.) must define
	// their own canonicalization specification and implement their own
	// serializer rather than relying on this API.
	//
	// If deterministic serialization is requested, map entries will be
	// sorted by keys in lexographical order. This is an implementation
	// detail and subject to change.
	Deterministic bool

	// UseCachedSize indicates that the result of a previous Size call
	// may be reused.
	//
	// Setting this option asserts that:
	//
	// 1. Size has previously been called on this message with identical
	// options (except for UseCachedSize itself).
	//
	// 2. The message and all its submessages have not changed in any
	// way since the Size call.
	//
	// If either of these invariants is violated,
	// the results are undefined and may include panics or corrupted output.
	//
	// Implementations MAY take this option into account to provide
	// better performance, but there is no guarantee that they will do so.
	// There is absolutely no guarantee that Size followed by Marshal with
	// UseCachedSize set will perform equivalently to Marshal alone.
	UseCachedSize bool
}

// Marshal returns the wire-format encoding of m.
func Marshal(m Message) ([]byte, error) {
	// Treat nil message interface as an empty message; nothing to output.
	if m == nil {
		return nil, nil
	}

	out, err := MarshalOptions{}.marshal(nil, m.ProtoReflect())
	if len(out.Buf) == 0 && err == nil {
		out.Buf = emptyBytesForMessage(m)
	}
	return out.Buf, err
}

// Marshal returns the wire-format encoding of m.
func (o MarshalOptions) Marshal(m Message) ([]byte, error) {
	// Treat nil message interface as an empty message; nothing to output.
	if m == nil {
		return nil, nil
	}

	out, err := o.marshal(nil, m.ProtoReflect())

	if len(out.Buf) == 0 && err == nil {
		out.Buf = emptyBytesForMessage(m)
	}

	return out.Buf, err
}

// emptyBytesForMessage returns a nil buffer if and only if m is invalid,
// otherwise it returns a non-nil empty buffer.
//
// This is to assist the edge-case where user-code does the following:
//	m1.OptionalBytes, _ = proto.Marshal(m2)
// where they expect the proto2 "optional_bytes" field to be populated
// if any only if m2 is a valid message.
func emptyBytesForMessage(m Message) []byte {
	if m == nil || !m.ProtoReflect().IsValid() {
		return nil
	}
	return emptyBuf[:]
}

// MarshalAppend appends the wire-format encoding of m to b,
// returning the result.
func (o MarshalOptions) MarshalAppend(b []byte, m Message) ([]byte, error) {
	// Treat nil message interface as an empty message; nothing to append.
	if m == nil {
		return b, nil
	}

	out, err := o.marshal(b, m.ProtoReflect())
	return out.Buf, err
}

// MarshalState returns the wire-format encoding of a message.
//
// This method permits fine-grained control over the marshaler.
// Most users should use Marshal instead.
func (o MarshalOptions) MarshalState(in protoiface.MarshalInput) (protoiface.MarshalOutput, error) {
	return o.marshal(in.Buf, in.Message)
}

// marshal is a centralized function that all marshal operations go through.
// For profiling purposes, avoid changing the name of this function or
// introducing other code paths for marshal that do not go through this.
func (o MarshalOptions) marshal(b []byte, m protoreflect.Message) (out protoiface.MarshalOutput, err error) {
	allowPartial := o.AllowPartial
	o.AllowPartial = true

	// 如果 m 提供了合法的 Marshal() 函数，就直接调用它。
	if methods := protoMethods(m);
		methods != nil &&
		methods.Marshal != nil &&
		!( o.Deterministic && methods.Flags&protoiface.SupportMarshalDeterministic == 0 ){

		// 构造输入
		in := protoiface.MarshalInput{
			Message: m,
			Buf:     b,
		}

		// 设置标记
		if o.Deterministic {
			in.Flags |= protoiface.MarshalDeterministic
		}
		if o.UseCachedSize {
			in.Flags |= protoiface.MarshalUseCachedSize
		}

		// 计算 size
		if methods.Size != nil {
			// 计算 size
			sout := methods.Size(protoiface.SizeInput{
				Message: m,
				Flags:   in.Flags,
			})
			// 如果 b 容量不足，则扩容
			if cap(b) < len(b)+sout.Size {
				in.Buf = make([]byte, len(b), growcap(cap(b), len(b)+sout.Size))
				copy(in.Buf, b)
			}
			// 设置标记
			in.Flags |= protoiface.MarshalUseCachedSize
		}

		// 执行 marshal()
		out, err = methods.Marshal(in)
	} else {
		// 如果 m 未提供合法的 Marshal() 函数，执行内置的 marshal() 函数。
		out.Buf, err = o.marshalMessageSlow(b, m)
	}

	// 出错
	if err != nil {
		return out, err
	}

	// 运行部分序列化
	if allowPartial {
		return out, nil
	}

	// 检查是否 OK
	return out, checkInitialized(m)
}

func (o MarshalOptions) marshalMessage(b []byte, m protoreflect.Message) ([]byte, error) {
	out, err := o.marshal(b, m)
	return out.Buf, err
}

// growcap scales up the capacity of a slice.
//
// Given a slice with a current capacity of oldcap and a desired
// capacity of wantcap, growcap returns a new capacity >= wantcap.
//
// The algorithm is mostly identical to the one used by append as of Go 1.14.
func growcap(oldcap, wantcap int) (newcap int) {
	if wantcap > oldcap*2 {
		newcap = wantcap
	} else if oldcap < 1024 {
		// The Go 1.14 runtime takes this case when len(s) < 1024,
		// not when cap(s) < 1024. The difference doesn't seem
		// significant here.
		newcap = oldcap * 2
	} else {
		newcap = oldcap
		for 0 < newcap && newcap < wantcap {
			newcap += newcap / 4
		}
		if newcap <= 0 {
			newcap = wantcap
		}
	}
	return newcap
}

func (o MarshalOptions) marshalMessageSlow(b []byte, m protoreflect.Message) ([]byte, error) {
	if messageset.IsMessageSet(m.Descriptor()) {
		return o.marshalMessageSet(b, m)
	}

	// 编码顺序，对复合结构有用，如 list/map 。
	fieldOrder := order.AnyFieldOrder
	if o.Deterministic {
		// TODO: This should use a more natural ordering like NumberFieldOrder,
		// but doing so breaks golden tests that make invalid assumption about
		// output stability of this implementation.
		fieldOrder = order.LegacyFieldOrder
	}

	var err error

	// 按 fieldOrder 顺序遍历 m 的 fields ，逐个调用 fn() 。
	order.RangeFields(m, fieldOrder, func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		// 将 fd 类型的 field 值 v 编码后存入 b 中。
		b, err = o.marshalField(b, fd, v)
		return err == nil
	})
	if err != nil {
		return b, err
	}

	// 追加未知字段
	b = append(b, m.GetUnknown()...)
	return b, nil
}

func (o MarshalOptions) marshalField(b []byte, fd protoreflect.FieldDescriptor, value protoreflect.Value) ([]byte, error) {
	switch {
	// 编码 list
	case fd.IsList():
		return o.marshalList(b, fd, value.List())
	// 编码 map
	case fd.IsMap():
		return o.marshalMap(b, fd, value.Map())
	// 编码单值类型
	default:
		// 编码 wiretag
		b = protowire.AppendTag(b, fd.Number(), wireTypes[fd.Kind()])
		// 编码 data
		return o.marshalSingular(b, fd, value)
	}
}

func (o MarshalOptions) marshalList(b []byte, fd protoreflect.FieldDescriptor, list protoreflect.List) ([]byte, error) {

	// 如果是 packed 类型，则整个 list 只需要一个 wiretag
	if fd.IsPacked() && list.Len() > 0 {
		// 编码 wiretag
		b = protowire.AppendTag(b, fd.Number(), protowire.BytesType)

		// 占位，用于后面填充总字节数
		b, pos := appendSpeculativeLength(b)

		// 遍历 list
		for i, llen := 0, list.Len(); i < llen; i++ {
			var err error
			// 编码 list[i] 元素
			b, err = o.marshalSingular(b, fd, list.Get(i))
			if err != nil {
				return b, err
			}
		}

		// 填充总字节数
		b = finishSpeculativeLength(b, pos)
		return b, nil
	}


	kind := fd.Kind()
	for i, llen := 0, list.Len(); i < llen; i++ {
		var err error
		b = protowire.AppendTag(b, fd.Number(), wireTypes[kind])
		b, err = o.marshalSingular(b, fd, list.Get(i))
		if err != nil {
			return b, err
		}
	}

	return b, nil
}

func (o MarshalOptions) marshalMap(b []byte, fd protoreflect.FieldDescriptor, mapv protoreflect.Map) ([]byte, error) {
	keyf := fd.MapKey()
	valf := fd.MapValue()

	keyOrder := order.AnyKeyOrder
	if o.Deterministic {
		keyOrder = order.GenericKeyOrder
	}

	var err error

	// 遍历 mapv ，逐个 kv 执行序列化
	order.RangeEntries(mapv, keyOrder, func(key protoreflect.MapKey, value protoreflect.Value) bool {
		// 编码 mapv 的 wiretag
		b = protowire.AppendTag(b, fd.Number(), protowire.BytesType)

		// 占位 length
		var pos int
		b, pos = appendSpeculativeLength(b)

		// 编码 key
		b, err = o.marshalField(b, keyf, key.Value())
		if err != nil {
			return false
		}

		// 编码 value
		b, err = o.marshalField(b, valf, value)
		if err != nil {
			return false
		}

		// 填充 length
		b = finishSpeculativeLength(b, pos)
		return true
	})
	return b, err
}

// When encoding length-prefixed fields, we speculatively set aside some number of bytes
// for the length, encode the data, and then encode the length (shifting the data if necessary
// to make room).
//
// 这里默认为 1Byte ，如果超过会整体后移，以腾出空间。
const speculativeLength = 1

// 预留长度字段的空间
func appendSpeculativeLength(b []byte) ([]byte, int) {
	pos := len(b)
	b = append(b, "\x00\x00\x00\x00"[:speculativeLength]...)
	return b, pos
}

// 回填长度
func finishSpeculativeLength(b []byte, pos int) []byte {
	// 计算写入的 data 长度 mlen
	mlen := len(b) - pos - speculativeLength
	// 对 mlen 进行 varint 数值编码，返回编码后所占字节数
	msiz := protowire.SizeVarint(uint64(mlen))
	// 如果 msiz 超过预留长度，则需要整体后移，腾出位置
	if msiz != speculativeLength {
		for i := 0; i < msiz-speculativeLength; i++ {
			b = append(b, 0)
		}
		copy(b[pos+msiz:], b[pos+speculativeLength:])
		b = b[:pos+msiz+mlen]
	}
	// 填入长度字段
	protowire.AppendVarint(b[:pos], uint64(mlen))
	return b
}
