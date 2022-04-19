// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package protoreflect

import (
	"google.golang.org/protobuf/internal/pragma"
)

// The following types are used by the fast-path Message.ProtoMethods method.
//
// To avoid polluting the public protoreflect API with types used only by
// low-level implementations, the canonical definitions of these types are
// in the runtime/protoiface package. The definitions here and in protoiface
// must be kept in sync.
//
// 以下类型被 Message.ProtoMethods 方法使用。
//
// 为了避免公共的 protoreflect API 被低级别的实现所使用的类型所污染，这些类型的标准定义在 runtime/protoiface 包中。
// 这里的定义和 protoiface 中的定义必须保持同步。
type (

	methods = struct {
		pragma.NoUnkeyedLiterals
		// 标记位
		Flags            supportFlags
		// 大小
		Size             func(sizeInput) sizeOutput
		// 序列化
		Marshal          func(marshalInput) (marshalOutput, error)
		// 反序列化
		Unmarshal        func(unmarshalInput) (unmarshalOutput, error)
		// 合并
		Merge            func(mergeInput) mergeOutput
		// 是否初始化
		CheckInitialized func(checkInitializedInput) (checkInitializedOutput, error)
	}

	supportFlags = uint64

	// Size Input/Output
	sizeInput    = struct {
		pragma.NoUnkeyedLiterals
		Message Message
		Flags   uint8
	}
	sizeOutput = struct {
		pragma.NoUnkeyedLiterals
		Size int
	}

	// Marshal Input/Output
	marshalInput = struct {
		pragma.NoUnkeyedLiterals
		Message Message
		Buf     []byte
		Flags   uint8
	}
	marshalOutput = struct {
		pragma.NoUnkeyedLiterals
		Buf []byte
	}

	// Unmarshal Input/Output
	unmarshalInput = struct {
		pragma.NoUnkeyedLiterals
		Message  Message
		Buf      []byte
		Flags    uint8
		Resolver interface {
			FindExtensionByName(field FullName) (ExtensionType, error)
			FindExtensionByNumber(message FullName, field FieldNumber) (ExtensionType, error)
		}
		Depth int
	}
	unmarshalOutput = struct {
		pragma.NoUnkeyedLiterals
		Flags uint8
	}

	// Merge Input/Output
	mergeInput = struct {
		pragma.NoUnkeyedLiterals
		Source      Message
		Destination Message
	}
	mergeOutput = struct {
		pragma.NoUnkeyedLiterals
		Flags uint8
	}

	// CheckInitialized Input/Output
	checkInitializedInput = struct {
		pragma.NoUnkeyedLiterals
		Message Message
	}
	checkInitializedOutput = struct {
		pragma.NoUnkeyedLiterals
	}
)
