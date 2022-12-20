// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package protodesc provides functionality for converting
// FileDescriptorProto messages to/from protoreflect.FileDescriptor values.
//
// The google.protobuf.FileDescriptorProto is a protobuf message that describes
// the type information for a .proto file in a form that is easily serializable.
// The protoreflect.FileDescriptor is a more structured representation of
// the FileDescriptorProto message where references and remote dependencies
// can be directly followed.
package protodesc

import (
	"google.golang.org/protobuf/internal/errors"
	"google.golang.org/protobuf/internal/filedesc"
	"google.golang.org/protobuf/internal/pragma"
	"google.golang.org/protobuf/internal/strs"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"google.golang.org/protobuf/types/descriptorpb"
)

// Resolver is the resolver used by NewFile to resolve dependencies.
// The enums and messages provided must belong to some parent file,
// which is also registered.
//
// It is implemented by protoregistry.Files.
//
// Resolver 是 NewFile 用来解决依赖关系的解析器。
// 引用的 enums 和 messages 必须属于某个父文件，且这个文件也注册了。
type Resolver interface {
	FindFileByPath(string) (protoreflect.FileDescriptor, error)
	FindDescriptorByName(protoreflect.FullName) (protoreflect.Descriptor, error)
}

// FileOptions configures the construction of file descriptors.
//
//
type FileOptions struct {
	pragma.NoUnkeyedLiterals

	// AllowUnresolvable configures New to permissively allow unresolvable
	// file, enum, or message dependencies. Unresolved dependencies are replaced
	// by placeholder equivalents.
	//
	// The following dependencies may be left unresolved:
	//	• Resolving an imported file.
	//	• Resolving the type for a message field or extension field.
	//	If the kind of the field is unknown, then a placeholder is used for both
	//	the Enum and Message accessors on the protoreflect.FieldDescriptor.
	//	• Resolving an enum value set as the default for an optional enum field.
	//	If unresolvable, the protoreflect.FieldDescriptor.Default is set to the
	//	first value in the associated enum (or zero if the also enum dependency
	//	is also unresolvable). The protoreflect.FieldDescriptor.DefaultEnumValue
	//	is populated with a placeholder.
	//	• Resolving the extended message type for an extension field.
	//	• Resolving the input or output message type for a service method.
	//
	// If the unresolved dependency uses a relative name,
	// then the placeholder will contain an invalid FullName with a "*." prefix,
	// indicating that the starting prefix of the full name is unknown.
	AllowUnresolvable bool
}

// NewFile creates a new protoreflect.FileDescriptor from the provided
// file descriptor message. See FileOptions.New for more information.
//
// NewFile 基于文件描述符 fd 创建一个新的 protoreflect.FileDescriptor 。
func NewFile(fd *descriptorpb.FileDescriptorProto, r Resolver) (protoreflect.FileDescriptor, error) {
	return FileOptions{}.New(fd, r)
}

// NewFiles creates a new protoregistry.Files from the provided
// FileDescriptorSet message. See FileOptions.NewFiles for more information.
func NewFiles(fd *descriptorpb.FileDescriptorSet) (*protoregistry.Files, error) {
	return FileOptions{}.NewFiles(fd)
}

// New creates a new protoreflect.FileDescriptor from the provided
// file descriptor message. The file must represent a valid proto file according
// to protobuf semantics. The returned descriptor is a deep copy of the input.
//
// Any imported files, enum types, or message types referenced in the file are
// resolved using the provided registry. When looking up an import file path,
// the path must be unique. The newly created file descriptor is not registered
// back into the provided file registry.
//
// New 基于文件描述符 fd 创建一个新的 protoreflect.FileDescriptor 。
// 根据 protobuf 的语义，该文件必须代表一个有效的 proto 文件，返回的描述符是对输入的深拷贝。
//
// 任何导入的文件、枚举类型或消息类型都是使用提供的注册表来解决的。
// 当查询导入文件的路径时，路径必须是唯一的。新创建的文件描述符不会被注册到所提供的文件注册表中。
//
func (o FileOptions) New(fd *descriptorpb.FileDescriptorProto, r Resolver) (protoreflect.FileDescriptor, error) {
	if r == nil {
		r = (*protoregistry.Files)(nil) // empty resolver
	}

	// Handle the file descriptor content.
	f := &filedesc.File{
		L2: &filedesc.FileL2{
		},
	}

	// 检查语法版本
	switch fd.GetSyntax() {
	case "proto2", "":
		f.L1.Syntax = protoreflect.Proto2
	case "proto3":
		f.L1.Syntax = protoreflect.Proto3
	default:
		return nil, errors.New("invalid syntax: %q", fd.GetSyntax())
	}

	// file name, relative to root of source tree
	f.L1.Path = fd.GetName()
	if f.L1.Path == "" {
		return nil, errors.New("file path must be populated")
	}

	// package name, e.g. "foo", "foo.bar", etc.
	f.L1.Package = protoreflect.FullName(fd.GetPackage())
	if !f.L1.Package.IsValid() && f.L1.Package != "" {
		return nil, errors.New("invalid package: %q", f.L1.Package)
	}

	// 选项
	if opts := fd.GetOptions(); opts != nil {
		opts = proto.Clone(opts).(*descriptorpb.FileOptions)
		f.L2.Options = func() protoreflect.ProtoMessage {
			return opts
		}
	}

	// 在 f.L2.Imports 中保存当前文件的所有依赖信息。
	f.L2.Imports = make(filedesc.FileImports, len(fd.GetDependency()))
	for _, i := range fd.GetPublicDependency() {
		// 数组越界或者重复 Import，报错
		if !(0 <= i && int(i) < len(f.L2.Imports)) || f.L2.Imports[i].IsPublic {
			return nil, errors.New("invalid or duplicate public import index: %d", i)
		}
		// 保存 Import
		f.L2.Imports[i].IsPublic = true
	}

	// 弱依赖，忽略～
	for _, i := range fd.GetWeakDependency() {
		if !(0 <= i && int(i) < len(f.L2.Imports)) || f.L2.Imports[i].IsWeak {
			return nil, errors.New("invalid or duplicate weak import index: %d", i)
		}
		f.L2.Imports[i].IsWeak = true
	}

	// 保存所有依赖的路径，自动去重
	imps := importSet{
		f.Path(): true,
	}

	// 遍历当前文件 fd 的所有直接依赖，解析该依赖文件对应的 fd ，并将其绑定到对应的依赖项上，同时保存到 importSet 用于重复检测。
	for i, path := range fd.GetDependency() {
		// 当前依赖项
		imp := &f.L2.Imports[i]

		// 根据 path 查找文件，这些依赖文件正常已经被解析过了
		impFd, err := r.FindFileByPath(path)

		// 未找到且允许弱依赖，则设置为占位符
		if err == protoregistry.NotFound && (o.AllowUnresolvable || imp.IsWeak) {
			impFd = filedesc.PlaceholderFile(path)
		// 否则，未找到直接报错
		} else if err != nil {
			return nil, errors.New("could not resolve import %q: %v", path, err)
		}

		// 找到依赖文件，则把 fd 保存到当前依赖项上
		imp.FileDescriptor = impFd

		// 检查是否重复导入
		if imps[imp.Path()] {
			return nil, errors.New("already imported %q", path)
		}
		// 保存到 set 中
		imps[imp.Path()] = true
	}

	// 遍历当前文件 fd 的所有依赖项，这些依赖也有自己的依赖，递归的把所有的依赖都保存到 imps 中
	for i := range fd.GetDependency() {
		imp := &f.L2.Imports[i]
		imps.importPublic(imp.Imports())
	}


	// Handle source locations.
	//
	// 把 fd 关联的源码位置信息拷贝一份，转存到 f.L2.Locations 中。
	f.L2.Locations.File = f
	for _, loc := range fd.GetSourceCodeInfo().GetLocation() {
		var l protoreflect.SourceLocation
		// TODO: Validate that the path points to an actual declaration?
		l.Path = protoreflect.SourcePath(loc.GetPath())
		s := loc.GetSpan()
		switch len(s) {
		case 3:
			l.StartLine, l.StartColumn, l.EndLine, l.EndColumn = int(s[0]), int(s[1]), int(s[0]), int(s[2])
		case 4:
			l.StartLine, l.StartColumn, l.EndLine, l.EndColumn = int(s[0]), int(s[1]), int(s[2]), int(s[3])
		default:
			return nil, errors.New("invalid span: %v", s)
		}

		// TODO: Validate that the span information is sensible?
		// See https://github.com/protocolbuffers/protobuf/issues/6378.
		if false && (l.EndLine < l.StartLine || l.StartLine < 0 || l.StartColumn < 0 || l.EndColumn < 0 ||
			(l.StartLine == l.EndLine && l.EndColumn <= l.StartColumn)) {
			return nil, errors.New("invalid span: %v", s)
		}
		l.LeadingDetachedComments = loc.GetLeadingDetachedComments()
		l.LeadingComments = loc.GetLeadingComments()
		l.TrailingComments = loc.GetTrailingComments()
		f.L2.Locations.List = append(f.L2.Locations.List, l)
	}

	// Step 1: Allocate and derive the names for all declarations.
	// This copies all fields from the descriptor proto except:
	//	google.protobuf.FieldDescriptorProto.type_name
	//	google.protobuf.FieldDescriptorProto.default_value
	//	google.protobuf.FieldDescriptorProto.oneof_index
	//	google.protobuf.FieldDescriptorProto.extendee
	//	google.protobuf.MethodDescriptorProto.input
	//	google.protobuf.MethodDescriptorProto.output
	var err error
	sb := new(strs.Builder)

	r1 := make(descsByName)
	if f.L1.Enums.List, err = r1.initEnumDeclarations(fd.GetEnumType(), f, sb); err != nil {
		return nil, err
	}

	if f.L1.Messages.List, err = r1.initMessagesDeclarations(fd.GetMessageType(), f, sb); err != nil {
		return nil, err
	}

	if f.L1.Extensions.List, err = r1.initExtensionDeclarations(fd.GetExtension(), f, sb); err != nil {
		return nil, err
	}

	if f.L1.Services.List, err = r1.initServiceDeclarations(fd.GetService(), f, sb); err != nil {
		return nil, err
	}

	// Step 2: Resolve every dependency reference not handled by step 1.
	r2 := &resolver{
		local: r1,
		remote: r,
		imports: imps,
		allowUnresolvable: o.AllowUnresolvable,
	}
	if err := r2.resolveMessageDependencies(f.L1.Messages.List, fd.GetMessageType()); err != nil {
		return nil, err
	}
	if err := r2.resolveExtensionDependencies(f.L1.Extensions.List, fd.GetExtension()); err != nil {
		return nil, err
	}
	if err := r2.resolveServiceDependencies(f.L1.Services.List, fd.GetService()); err != nil {
		return nil, err
	}

	// Step 3: Validate every enum, message, and extension declaration.
	if err := validateEnumDeclarations(f.L1.Enums.List, fd.GetEnumType()); err != nil {
		return nil, err
	}
	if err := validateMessageDeclarations(f.L1.Messages.List, fd.GetMessageType()); err != nil {
		return nil, err
	}
	if err := validateExtensionDeclarations(f.L1.Extensions.List, fd.GetExtension()); err != nil {
		return nil, err
	}

	return f, nil
}

type importSet map[string]bool

func (is importSet) importPublic(imps protoreflect.FileImports) {
	for i := 0; i < imps.Len(); i++ {
		if imp := imps.Get(i); imp.IsPublic {
			is[imp.Path()] = true
			is.importPublic(imp.Imports())
		}
	}
}

// NewFiles creates a new protoregistry.Files from the provided
// FileDescriptorSet message. The descriptor set must include only
// valid files according to protobuf semantics. The returned descriptors
// are a deep copy of the input.
func (o FileOptions) NewFiles(fds *descriptorpb.FileDescriptorSet) (*protoregistry.Files, error) {
	files := make(map[string]*descriptorpb.FileDescriptorProto)
	for _, fd := range fds.File {
		if _, ok := files[fd.GetName()]; ok {
			return nil, errors.New("file appears multiple times: %q", fd.GetName())
		}
		files[fd.GetName()] = fd
	}
	r := &protoregistry.Files{}
	for _, fd := range files {
		if err := o.addFileDeps(r, fd, files); err != nil {
			return nil, err
		}
	}
	return r, nil
}
func (o FileOptions) addFileDeps(r *protoregistry.Files, fd *descriptorpb.FileDescriptorProto, files map[string]*descriptorpb.FileDescriptorProto) error {
	// Set the entry to nil while descending into a file's dependencies to detect cycles.
	files[fd.GetName()] = nil
	for _, dep := range fd.Dependency {
		depfd, ok := files[dep]
		if depfd == nil {
			if ok {
				return errors.New("import cycle in file: %q", dep)
			}
			continue
		}
		if err := o.addFileDeps(r, depfd, files); err != nil {
			return err
		}
	}
	// Delete the entry once dependencies are processed.
	delete(files, fd.GetName())
	f, err := o.New(fd, r)
	if err != nil {
		return err
	}
	return r.RegisterFile(f)
}
