// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package order provides ordered access to messages and maps.
package order

import (
	"sort"
	"sync"

	pref "google.golang.org/protobuf/reflect/protoreflect"
)

type messageField struct {
	fd pref.FieldDescriptor	// 字段描述符
	v  pref.Value			// 字段值
}

var messageFieldPool = sync.Pool{
	New: func() interface{} {
		return new([]messageField)	// 字段数组缓冲池
	},
}

type (

	// FieldRanger is an interface for visiting all fields in a message.
	// The protoreflect.Message type implements this interface.
	//
	// 遍历 message 内的所有 fields .
	FieldRanger interface{
		Range(VisitField)
	}

	// VisitField is called everytime a message field is visited.
	VisitField = func(pref.FieldDescriptor, pref.Value) bool
)

// RangeFields iterates over the fields of fs according to the specified order.
//
// 因为默认的 range 是无序的，如果需要有序，需要先 range 把 fields 提取出来、排序、再 range + fn 逐个处理。
func RangeFields(fs FieldRanger, less FieldOrder, fn VisitField) {

	// 未指定 order ，直接调用 fs
	if less == nil {
		fs.Range(fn)
		return
	}

	// 若指定 order ，就需要提取出 message 中所有 fields ，排序后再逐个应用 fn 。



	// 分配 fields 数组，把 message 中的每个 field 提取出来缓存到数组中
	//
	// Obtain a pre-allocated scratch buffer.
	p := messageFieldPool.Get().(*[]messageField)
	fields := (*p)[:0]
	defer func() {
		if cap(fields) < 1024 {
			*p = fields
			messageFieldPool.Put(p)
		}
	}()
	// Collect all fields in the message and sort them.
	fs.Range(func(fd pref.FieldDescriptor, v pref.Value) bool {
		fields = append(fields, messageField{fd, v})
		return true
	})


	// 对 fields 数组进行排序
	sort.Slice(fields, func(i, j int) bool {
		return less(fields[i].fd, fields[j].fd)
	})

	// Visit the fields in the specified ordering.
	//
	// 对 fields 数组逐个执行 fn()
	for _, f := range fields {
		if !fn(f.fd, f.v) {
			return
		}
	}
}

type mapEntry struct {
	k pref.MapKey
	v pref.Value
}

var mapEntryPool = sync.Pool{
	New: func() interface{} { return new([]mapEntry) },
}

type (
	// EntryRanger is an interface for visiting all fields in a message.
	// The protoreflect.Map type implements this interface.
	EntryRanger interface{ Range(VisitEntry) }
	// VisitEntry is called everytime a map entry is visited.
	VisitEntry = func(pref.MapKey, pref.Value) bool
)

// RangeEntries iterates over the entries of es according to the specified order.
func RangeEntries(es EntryRanger, less KeyOrder, fn VisitEntry) {
	if less == nil {
		es.Range(fn)
		return
	}

	// Obtain a pre-allocated scratch buffer.
	p := mapEntryPool.Get().(*[]mapEntry)
	entries := (*p)[:0]
	defer func() {
		if cap(entries) < 1024 {
			*p = entries
			mapEntryPool.Put(p)
		}
	}()

	// Collect all entries in the map and sort them.
	es.Range(func(k pref.MapKey, v pref.Value) bool {
		entries = append(entries, mapEntry{k, v})
		return true
	})
	sort.Slice(entries, func(i, j int) bool {
		return less(entries[i].k, entries[j].k)
	})

	// Visit the entries in the specified ordering.
	for _, e := range entries {
		if !fn(e.k, e.v) {
			return
		}
	}
}
