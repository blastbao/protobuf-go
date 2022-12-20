// Protocol Buffers - Google's data interchange format
// Copyright 2008 Google Inc.  All rights reserved.
// https://developers.google.com/protocol-buffers/
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//     * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//     * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//     * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// Author: kenton@google.com (Kenton Varda)
//
// WARNING:  The plugin interface is currently EXPERIMENTAL and is subject to
//   change.
//
// protoc (aka the Protocol Compiler) can be extended via plugins.  A plugin is
// just a program that reads a CodeGeneratorRequest from stdin and writes a
// CodeGeneratorResponse to stdout.
//
// Plugins written using C++ can use google/protobuf/compiler/plugin.h instead
// of dealing with the raw protocol defined here.
//
// A plugin executable needs only to be placed somewhere in the path.  The
// plugin should be named "protoc-gen-$NAME", and will then be used when the
// flag "--${NAME}_out" is passed to protoc.

// Code generated by protoc-gen-go. DO NOT EDIT.
// source: google/protobuf/compiler/plugin.proto

package pluginpb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
	reflect "reflect"
	sync "sync"
)

// Sync with code_generator.h.
type CodeGeneratorResponse_Feature int32

const (
	CodeGeneratorResponse_FEATURE_NONE            CodeGeneratorResponse_Feature = 0
	CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL CodeGeneratorResponse_Feature = 1
)

// Enum value maps for CodeGeneratorResponse_Feature.
var (
	CodeGeneratorResponse_Feature_name = map[int32]string{
		0: "FEATURE_NONE",
		1: "FEATURE_PROTO3_OPTIONAL",
	}
	CodeGeneratorResponse_Feature_value = map[string]int32{
		"FEATURE_NONE":            0,
		"FEATURE_PROTO3_OPTIONAL": 1,
	}
)

func (x CodeGeneratorResponse_Feature) Enum() *CodeGeneratorResponse_Feature {
	p := new(CodeGeneratorResponse_Feature)
	*p = x
	return p
}

func (x CodeGeneratorResponse_Feature) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (CodeGeneratorResponse_Feature) Descriptor() protoreflect.EnumDescriptor {
	return file_google_protobuf_compiler_plugin_proto_enumTypes[0].Descriptor()
}

func (CodeGeneratorResponse_Feature) Type() protoreflect.EnumType {
	return &file_google_protobuf_compiler_plugin_proto_enumTypes[0]
}

func (x CodeGeneratorResponse_Feature) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Do not use.
func (x *CodeGeneratorResponse_Feature) UnmarshalJSON(b []byte) error {
	num, err := protoimpl.X.UnmarshalJSONEnum(x.Descriptor(), b)
	if err != nil {
		return err
	}
	*x = CodeGeneratorResponse_Feature(num)
	return nil
}

// Deprecated: Use CodeGeneratorResponse_Feature.Descriptor instead.
func (CodeGeneratorResponse_Feature) EnumDescriptor() ([]byte, []int) {
	return file_google_protobuf_compiler_plugin_proto_rawDescGZIP(), []int{2, 0}
}

// The version number of protocol compiler.
type Version struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Major *int32 `protobuf:"varint,1,opt,name=major" json:"major,omitempty"`
	Minor *int32 `protobuf:"varint,2,opt,name=minor" json:"minor,omitempty"`
	Patch *int32 `protobuf:"varint,3,opt,name=patch" json:"patch,omitempty"`
	// A suffix for alpha, beta or rc release, e.g., "alpha-1", "rc2". It should
	// be empty for mainline stable releases.
	Suffix *string `protobuf:"bytes,4,opt,name=suffix" json:"suffix,omitempty"`
}

func (x *Version) Reset() {
	*x = Version{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_protobuf_compiler_plugin_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Version) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Version) ProtoMessage() {}

func (x *Version) ProtoReflect() protoreflect.Message {
	mi := &file_google_protobuf_compiler_plugin_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Version.ProtoReflect.Descriptor instead.
func (*Version) Descriptor() ([]byte, []int) {
	return file_google_protobuf_compiler_plugin_proto_rawDescGZIP(), []int{0}
}

func (x *Version) GetMajor() int32 {
	if x != nil && x.Major != nil {
		return *x.Major
	}
	return 0
}

func (x *Version) GetMinor() int32 {
	if x != nil && x.Minor != nil {
		return *x.Minor
	}
	return 0
}

func (x *Version) GetPatch() int32 {
	if x != nil && x.Patch != nil {
		return *x.Patch
	}
	return 0
}

func (x *Version) GetSuffix() string {
	if x != nil && x.Suffix != nil {
		return *x.Suffix
	}
	return ""
}

// An encoded CodeGeneratorRequest is written to the plugin's stdin.
type CodeGeneratorRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The .proto files that were explicitly listed on the command-line.  The
	// code generator should generate code only for these files.  Each file's
	// descriptor will be included in proto_file, below.
	//
	// 在命令行中明确列出的 .proto 文件名，generator 只为这些文件生成代码。
	// 每个文件的 descriptor 被包含在下面的 proto_file 中。
	FileToGenerate []string `protobuf:"bytes,1,rep,name=file_to_generate,json=fileToGenerate" json:"file_to_generate,omitempty"`

	// The generator parameter passed on the command-line.
	//
	// 命令行传递的 generator 参数。
	Parameter *string `protobuf:"bytes,2,opt,name=parameter" json:"parameter,omitempty"`

	// FileDescriptorProtos for all files in files_to_generate and everything
	// they import.  The files will appear in topological order, so each file
	// appears before any file that imports it.
	//
	// protoc guarantees that all proto_files will be written after
	// the fields above, even though this is not technically guaranteed by the
	// protobuf wire format.  This theoretically could allow a plugin to stream
	// in the FileDescriptorProtos and handle them one by one rather than read
	// the entire set into memory at once.  However, as of this writing, this
	// is not similarly optimized on protoc's end -- it will store all fields in
	// memory at once before sending them to the plugin.
	//
	// Type names of fields and extensions in the FileDescriptorProto are always
	// fully qualified.
	//
	// FileDescriptorProtos 包含 FileToGenerate 中的所有文件，以及它们导入文件的描述符。
	// 这些文件将以拓扑顺序出现，所以每个文件都出现在导入它的任何文件之前。
	//
	// protoc 保证所有的 proto_files 将被写在上面的字段之后，尽管这在技术上并不被 protobuf 线格式所保证。
	// 理论上，这可以让一个插件以流式方式输入 FileDescriptorProtos 并逐一处理它们，而不是一次性将整个集合读入内存。
	// 然而，在撰写本文时，protoc 并没有进行类似的优化 -- 它将在把所有字段发送到插件之前一次性存储在内存中。
	//
	// FileDescriptorProto 中字段和扩展名的类型名称总是合法的。
	//
	ProtoFile []*descriptorpb.FileDescriptorProto `protobuf:"bytes,15,rep,name=proto_file,json=protoFile" json:"proto_file,omitempty"`

	// The version number of protocol compiler.
	CompilerVersion *Version `protobuf:"bytes,3,opt,name=compiler_version,json=compilerVersion" json:"compiler_version,omitempty"`
}

func (x *CodeGeneratorRequest) Reset() {
	*x = CodeGeneratorRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_protobuf_compiler_plugin_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CodeGeneratorRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CodeGeneratorRequest) ProtoMessage() {}

func (x *CodeGeneratorRequest) ProtoReflect() protoreflect.Message {
	mi := &file_google_protobuf_compiler_plugin_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CodeGeneratorRequest.ProtoReflect.Descriptor instead.
func (*CodeGeneratorRequest) Descriptor() ([]byte, []int) {
	return file_google_protobuf_compiler_plugin_proto_rawDescGZIP(), []int{1}
}

func (x *CodeGeneratorRequest) GetFileToGenerate() []string {
	if x != nil {
		return x.FileToGenerate
	}
	return nil
}

func (x *CodeGeneratorRequest) GetParameter() string {
	if x != nil && x.Parameter != nil {
		return *x.Parameter
	}
	return ""
}

func (x *CodeGeneratorRequest) GetProtoFile() []*descriptorpb.FileDescriptorProto {
	if x != nil {
		return x.ProtoFile
	}
	return nil
}

func (x *CodeGeneratorRequest) GetCompilerVersion() *Version {
	if x != nil {
		return x.CompilerVersion
	}
	return nil
}

// The plugin writes an encoded CodeGeneratorResponse to stdout.
type CodeGeneratorResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Error message.  If non-empty, code generation failed.  The plugin process
	// should exit with status code zero even if it reports an error in this way.
	//
	// This should be used to indicate errors in .proto files which prevent the
	// code generator from generating correct code.  Errors which indicate a
	// problem in protoc itself -- such as the input CodeGeneratorRequest being
	// unparseable -- should be reported by writing a message to stderr and
	// exiting with a non-zero status code.
	//
	// 生成失败
	Error *string `protobuf:"bytes,1,opt,name=error" json:"error,omitempty"`

	// A bitmask of supported features that the code generator supports.
	// This is a bitwise "or" of values from the Feature enum.
	//
	// generator 所支持特性的位掩码。
	SupportedFeatures *uint64                       `protobuf:"varint,2,opt,name=supported_features,json=supportedFeatures" json:"supported_features,omitempty"`
	File              []*CodeGeneratorResponse_File `protobuf:"bytes,15,rep,name=file" json:"file,omitempty"`
}

func (x *CodeGeneratorResponse) Reset() {
	*x = CodeGeneratorResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_protobuf_compiler_plugin_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CodeGeneratorResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CodeGeneratorResponse) ProtoMessage() {}

func (x *CodeGeneratorResponse) ProtoReflect() protoreflect.Message {
	mi := &file_google_protobuf_compiler_plugin_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CodeGeneratorResponse.ProtoReflect.Descriptor instead.
func (*CodeGeneratorResponse) Descriptor() ([]byte, []int) {
	return file_google_protobuf_compiler_plugin_proto_rawDescGZIP(), []int{2}
}

func (x *CodeGeneratorResponse) GetError() string {
	if x != nil && x.Error != nil {
		return *x.Error
	}
	return ""
}

func (x *CodeGeneratorResponse) GetSupportedFeatures() uint64 {
	if x != nil && x.SupportedFeatures != nil {
		return *x.SupportedFeatures
	}
	return 0
}

func (x *CodeGeneratorResponse) GetFile() []*CodeGeneratorResponse_File {
	if x != nil {
		return x.File
	}
	return nil
}

// Represents a single generated file.
type CodeGeneratorResponse_File struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The file name, relative to the output directory.  The name must not
	// contain "." or ".." components and must be relative, not be absolute (so,
	// the file cannot lie outside the output directory).  "/" must be used as
	// the path separator, not "\".
	//
	// If the name is omitted, the content will be appended to the previous
	// file.  This allows the generator to break large files into small chunks,
	// and allows the generated text to be streamed back to protoc so that large
	// files need not reside completely in memory at one time.  Note that as of
	// this writing protoc does not optimize for this -- it will read the entire
	// CodeGeneratorResponse before writing files to disk.
	Name *string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	// If non-empty, indicates that the named file should already exist, and the
	// content here is to be inserted into that file at a defined insertion
	// point.  This feature allows a code generator to extend the output
	// produced by another code generator.  The original generator may provide
	// insertion points by placing special annotations in the file that look
	// like:
	//   @@protoc_insertion_point(NAME)
	// The annotation can have arbitrary text before and after it on the line,
	// which allows it to be placed in a comment.  NAME should be replaced with
	// an identifier naming the point -- this is what other generators will use
	// as the insertion_point.  Code inserted at this point will be placed
	// immediately above the line containing the insertion point (thus multiple
	// insertions to the same point will come out in the order they were added).
	// The double-@ is intended to make it unlikely that the generated code
	// could contain things that look like insertion points by accident.
	//
	// For example, the C++ code generator places the following line in the
	// .pb.h files that it generates:
	//   // @@protoc_insertion_point(namespace_scope)
	// This line appears within the scope of the file's package namespace, but
	// outside of any particular class.  Another plugin can then specify the
	// insertion_point "namespace_scope" to generate additional classes or
	// other declarations that should be placed in this scope.
	//
	// Note that if the line containing the insertion point begins with
	// whitespace, the same whitespace will be added to every line of the
	// inserted text.  This is useful for languages like Python, where
	// indentation matters.  In these languages, the insertion point comment
	// should be indented the same amount as any inserted code will need to be
	// in order to work correctly in that context.
	//
	// The code generator that generates the initial file and the one which
	// inserts into it must both run as part of a single invocation of protoc.
	// Code generators are executed in the order in which they appear on the
	// command line.
	//
	// If |insertion_point| is present, |name| must also be present.
	InsertionPoint *string `protobuf:"bytes,2,opt,name=insertion_point,json=insertionPoint" json:"insertion_point,omitempty"`
	// The file contents.
	Content *string `protobuf:"bytes,15,opt,name=content" json:"content,omitempty"`
	// Information describing the file content being inserted. If an insertion
	// point is used, this information will be appropriately offset and inserted
	// into the code generation metadata for the generated files.
	GeneratedCodeInfo *descriptorpb.GeneratedCodeInfo `protobuf:"bytes,16,opt,name=generated_code_info,json=generatedCodeInfo" json:"generated_code_info,omitempty"`
}

func (x *CodeGeneratorResponse_File) Reset() {
	*x = CodeGeneratorResponse_File{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_protobuf_compiler_plugin_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CodeGeneratorResponse_File) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CodeGeneratorResponse_File) ProtoMessage() {}

func (x *CodeGeneratorResponse_File) ProtoReflect() protoreflect.Message {
	mi := &file_google_protobuf_compiler_plugin_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CodeGeneratorResponse_File.ProtoReflect.Descriptor instead.
func (*CodeGeneratorResponse_File) Descriptor() ([]byte, []int) {
	return file_google_protobuf_compiler_plugin_proto_rawDescGZIP(), []int{2, 0}
}

func (x *CodeGeneratorResponse_File) GetName() string {
	if x != nil && x.Name != nil {
		return *x.Name
	}
	return ""
}

func (x *CodeGeneratorResponse_File) GetInsertionPoint() string {
	if x != nil && x.InsertionPoint != nil {
		return *x.InsertionPoint
	}
	return ""
}

func (x *CodeGeneratorResponse_File) GetContent() string {
	if x != nil && x.Content != nil {
		return *x.Content
	}
	return ""
}

func (x *CodeGeneratorResponse_File) GetGeneratedCodeInfo() *descriptorpb.GeneratedCodeInfo {
	if x != nil {
		return x.GeneratedCodeInfo
	}
	return nil
}

var File_google_protobuf_compiler_plugin_proto protoreflect.FileDescriptor

var file_google_protobuf_compiler_plugin_proto_rawDesc = []byte{
	0x0a, 0x25, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2f, 0x63, 0x6f, 0x6d, 0x70, 0x69, 0x6c, 0x65, 0x72, 0x2f, 0x70, 0x6c, 0x75, 0x67, 0x69,
	0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x18, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x63, 0x6f, 0x6d, 0x70, 0x69, 0x6c, 0x65,
	0x72, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0x63, 0x0a, 0x07, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x14,
	0x0a, 0x05, 0x6d, 0x61, 0x6a, 0x6f, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x6d,
	0x61, 0x6a, 0x6f, 0x72, 0x12, 0x14, 0x0a, 0x05, 0x6d, 0x69, 0x6e, 0x6f, 0x72, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x05, 0x6d, 0x69, 0x6e, 0x6f, 0x72, 0x12, 0x14, 0x0a, 0x05, 0x70, 0x61,
	0x74, 0x63, 0x68, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x70, 0x61, 0x74, 0x63, 0x68,
	0x12, 0x16, 0x0a, 0x06, 0x73, 0x75, 0x66, 0x66, 0x69, 0x78, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x06, 0x73, 0x75, 0x66, 0x66, 0x69, 0x78, 0x22, 0xf1, 0x01, 0x0a, 0x14, 0x43, 0x6f, 0x64,
	0x65, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x28, 0x0a, 0x10, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x74, 0x6f, 0x5f, 0x67, 0x65, 0x6e,
	0x65, 0x72, 0x61, 0x74, 0x65, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0e, 0x66, 0x69, 0x6c,
	0x65, 0x54, 0x6f, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x70,
	0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09,
	0x70, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x12, 0x43, 0x0a, 0x0a, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x18, 0x0f, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x24, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e,
	0x46, 0x69, 0x6c, 0x65, 0x44, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x50, 0x72,
	0x6f, 0x74, 0x6f, 0x52, 0x09, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x46, 0x69, 0x6c, 0x65, 0x12, 0x4c,
	0x0a, 0x10, 0x63, 0x6f, 0x6d, 0x70, 0x69, 0x6c, 0x65, 0x72, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69,
	0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x63, 0x6f, 0x6d, 0x70, 0x69,
	0x6c, 0x65, 0x72, 0x2e, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x0f, 0x63, 0x6f, 0x6d,
	0x70, 0x69, 0x6c, 0x65, 0x72, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x22, 0x94, 0x03, 0x0a,
	0x15, 0x43, 0x6f, 0x64, 0x65, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x2d, 0x0a, 0x12,
	0x73, 0x75, 0x70, 0x70, 0x6f, 0x72, 0x74, 0x65, 0x64, 0x5f, 0x66, 0x65, 0x61, 0x74, 0x75, 0x72,
	0x65, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x11, 0x73, 0x75, 0x70, 0x70, 0x6f, 0x72,
	0x74, 0x65, 0x64, 0x46, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x73, 0x12, 0x48, 0x0a, 0x04, 0x66,
	0x69, 0x6c, 0x65, 0x18, 0x0f, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x34, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x63, 0x6f, 0x6d, 0x70,
	0x69, 0x6c, 0x65, 0x72, 0x2e, 0x43, 0x6f, 0x64, 0x65, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74,
	0x6f, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x46, 0x69, 0x6c, 0x65, 0x52,
	0x04, 0x66, 0x69, 0x6c, 0x65, 0x1a, 0xb1, 0x01, 0x0a, 0x04, 0x46, 0x69, 0x6c, 0x65, 0x12, 0x12,
	0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x12, 0x27, 0x0a, 0x0f, 0x69, 0x6e, 0x73, 0x65, 0x72, 0x74, 0x69, 0x6f, 0x6e, 0x5f,
	0x70, 0x6f, 0x69, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x69, 0x6e, 0x73,
	0x65, 0x72, 0x74, 0x69, 0x6f, 0x6e, 0x50, 0x6f, 0x69, 0x6e, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x63,
	0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x0f, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x6f,
	0x6e, 0x74, 0x65, 0x6e, 0x74, 0x12, 0x52, 0x0a, 0x13, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74,
	0x65, 0x64, 0x5f, 0x63, 0x6f, 0x64, 0x65, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x18, 0x10, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x22, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x64, 0x43, 0x6f,
	0x64, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x11, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65,
	0x64, 0x43, 0x6f, 0x64, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x22, 0x38, 0x0a, 0x07, 0x46, 0x65, 0x61,
	0x74, 0x75, 0x72, 0x65, 0x12, 0x10, 0x0a, 0x0c, 0x46, 0x45, 0x41, 0x54, 0x55, 0x52, 0x45, 0x5f,
	0x4e, 0x4f, 0x4e, 0x45, 0x10, 0x00, 0x12, 0x1b, 0x0a, 0x17, 0x46, 0x45, 0x41, 0x54, 0x55, 0x52,
	0x45, 0x5f, 0x50, 0x52, 0x4f, 0x54, 0x4f, 0x33, 0x5f, 0x4f, 0x50, 0x54, 0x49, 0x4f, 0x4e, 0x41,
	0x4c, 0x10, 0x01, 0x42, 0x57, 0x0a, 0x1c, 0x63, 0x6f, 0x6d, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x63, 0x6f, 0x6d, 0x70, 0x69,
	0x6c, 0x65, 0x72, 0x42, 0x0c, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x50, 0x72, 0x6f, 0x74, 0x6f,
	0x73, 0x5a, 0x29, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x67, 0x6f, 0x6c, 0x61, 0x6e, 0x67,
	0x2e, 0x6f, 0x72, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x79,
	0x70, 0x65, 0x73, 0x2f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x70, 0x62,
}

var (
	file_google_protobuf_compiler_plugin_proto_rawDescOnce sync.Once
	file_google_protobuf_compiler_plugin_proto_rawDescData = file_google_protobuf_compiler_plugin_proto_rawDesc
)

func file_google_protobuf_compiler_plugin_proto_rawDescGZIP() []byte {
	file_google_protobuf_compiler_plugin_proto_rawDescOnce.Do(func() {
		file_google_protobuf_compiler_plugin_proto_rawDescData = protoimpl.X.CompressGZIP(file_google_protobuf_compiler_plugin_proto_rawDescData)
	})
	return file_google_protobuf_compiler_plugin_proto_rawDescData
}

var file_google_protobuf_compiler_plugin_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_google_protobuf_compiler_plugin_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_google_protobuf_compiler_plugin_proto_goTypes = []interface{}{
	(CodeGeneratorResponse_Feature)(0),       // 0: google.protobuf.compiler.CodeGeneratorResponse.Feature
	(*Version)(nil),                          // 1: google.protobuf.compiler.Version
	(*CodeGeneratorRequest)(nil),             // 2: google.protobuf.compiler.CodeGeneratorRequest
	(*CodeGeneratorResponse)(nil),            // 3: google.protobuf.compiler.CodeGeneratorResponse
	(*CodeGeneratorResponse_File)(nil),       // 4: google.protobuf.compiler.CodeGeneratorResponse.File
	(*descriptorpb.FileDescriptorProto)(nil), // 5: google.protobuf.FileDescriptorProto
	(*descriptorpb.GeneratedCodeInfo)(nil),   // 6: google.protobuf.GeneratedCodeInfo
}
var file_google_protobuf_compiler_plugin_proto_depIdxs = []int32{
	5, // 0: google.protobuf.compiler.CodeGeneratorRequest.proto_file:type_name -> google.protobuf.FileDescriptorProto
	1, // 1: google.protobuf.compiler.CodeGeneratorRequest.compiler_version:type_name -> google.protobuf.compiler.Version
	4, // 2: google.protobuf.compiler.CodeGeneratorResponse.file:type_name -> google.protobuf.compiler.CodeGeneratorResponse.File
	6, // 3: google.protobuf.compiler.CodeGeneratorResponse.File.generated_code_info:type_name -> google.protobuf.GeneratedCodeInfo
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_google_protobuf_compiler_plugin_proto_init() }
func file_google_protobuf_compiler_plugin_proto_init() {
	if File_google_protobuf_compiler_plugin_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_google_protobuf_compiler_plugin_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Version); i {
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
		file_google_protobuf_compiler_plugin_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CodeGeneratorRequest); i {
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
		file_google_protobuf_compiler_plugin_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CodeGeneratorResponse); i {
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
		file_google_protobuf_compiler_plugin_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CodeGeneratorResponse_File); i {
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
			RawDescriptor: file_google_protobuf_compiler_plugin_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_google_protobuf_compiler_plugin_proto_goTypes,
		DependencyIndexes: file_google_protobuf_compiler_plugin_proto_depIdxs,
		EnumInfos:         file_google_protobuf_compiler_plugin_proto_enumTypes,
		MessageInfos:      file_google_protobuf_compiler_plugin_proto_msgTypes,
	}.Build()
	File_google_protobuf_compiler_plugin_proto = out.File
	file_google_protobuf_compiler_plugin_proto_rawDesc = nil
	file_google_protobuf_compiler_plugin_proto_goTypes = nil
	file_google_protobuf_compiler_plugin_proto_depIdxs = nil
}
