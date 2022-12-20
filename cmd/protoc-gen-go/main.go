// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The protoc-gen-go binary is a protoc plugin to generate Go code for
// both proto2 and proto3 versions of the protocol buffer language.
//
// For more information about the usage of this plugin, see:
//	https://developers.google.com/protocol-buffers/docs/reference/go-generated
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	gengo "google.golang.org/protobuf/cmd/protoc-gen-go/internal_gengo"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/internal/version"
)

const genGoDocURL = "https://developers.google.com/protocol-buffers/docs/reference/go-generated"
const grpcDocURL = "https://grpc.io/docs/languages/go/quickstart/#regenerate-grpc-code"

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Fprintf(os.Stdout, "%v %v\n", filepath.Base(os.Args[0]), version.String())
		os.Exit(0)
	}
	if len(os.Args) == 2 && os.Args[1] == "--help" {
		fmt.Fprintf(os.Stdout, "See "+genGoDocURL+" for usage information.\n")
		os.Exit(0)
	}

	var (
		flags   flag.FlagSet	// 保存所有命令行参数
		plugins = flags.String("plugins", "", "deprecated option")
	)

	// 入口
	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(gen *protogen.Plugin) error {
		// 不支持 plugin
		if *plugins != "" {
			return errors.New("protoc-gen-go: plugins are not supported; use 'protoc --go-grpc_out=...' to generate gRPC\n\n" +
				"See " + grpcDocURL + " for more information.")
		}

		// 遍历所有文件，包含待生成、被导入的所有文件，对于其中指定需要生成代码的文件，执行生成
		for _, f := range gen.Files {
			if f.Generate {
				gengo.GenerateFile(gen, f) 	// [重要] 核心生成逻辑，生成 xx.pb.go
			}
		}

		//
		gen.SupportedFeatures = gengo.SupportedFeatures
		return nil
	})
}
