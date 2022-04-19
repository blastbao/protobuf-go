### 前言

protoc 工具可以干什么？

protoc 工具可以通过相关插件，将 .proto 文件编译成 C、C++、Golang、Java、Python、PHP 等多种语言的代码。

本文主要讨论通过 protoc 生成 Golang 代码，例如我们常见的命令：

```
protoc -I . --go_out=xxx
```

想了解更多参数，执行 protoc --help 查看。


### 疑惑

#### 一、如何知道 protoc 使用的什么插件？

例如：--go_out 使用的是什么插件？最终了解到使用的是 protoc-gen-go 插件。
例如：--go-grpc_out 使用的是什么插件？最终了解到使用的是 protoc-gen-go-grpc 插件。

也通过使用其他插件，总结出一个规律：
- go_out 对应 protoc-gen-go 插件；
- go-grpc_out 对应 protoc-gen-go-grpc 插件；
- ...
- *_out 对应 protoc-gen-* 插件；


#### 二、例如新老项目使用的 protoc-gen-go 插件版本不同怎么办？

我能想到两个方案解决：

0. 通过两个环境去完成，例如，打两个 docker 环境，新项目在一个环境中生成，旧项目在另一个环境中生成。
1. 通过区分插件名称去完成，例如，将新版本命名为 protoc-gen-go-new，将旧版本命名为 protoc-gen-go-old，生成新版本时使用 --go-new_out，生成旧版本时使用 --go-old_out。
很显然，第 2 个方案成本更小。

#### 三、protoc-gen-go 和 protoc-gen-go-grpc 这两个插件有什么不同？

当使用参数 --go_out=plugins=grpc:xxx 生成时，生成的文件 *.pb.go 包含消息序列化代码和 gRPC 代码。
当使用参数 --go_out=xxx --go-grpc_out=xxx 生成时，会生成两个文件 *.pb.go 和 *._grpc.pb.go ，它们分别是消息序列化代码和 gRPC 代码。

为什么会存在这两种生成方式？它们有什么不同？这是我查询到的资料：

原文：Differences between protoc-gen-go and protoc-gen-go-grpc
[https://stackoverflow.com/questions/64828054/differences-between-protoc-gen-go-and-protoc-gen-go-grpc]

```
The old-way is using the github.com/golang/protobuf module. It comes with protoc-gen-go that generates both serialization of the protobuf messages and grpc code (when --go_out=plugins=grpc is used).

The so-called new-way is using the google.golang.org/protobuf module = a major revision of the Go bindings for protocol buffers. It comes with a different protoc-gen-go that no longer supports generating gRPC service definitions. For gRPC code, a new plugin called protoc-gen-go-grpc was developed by Go gRPC project. The plugins flag, which provided a way to invoke the gRPC code generator in the old-way, is deprecated.

Important note: Users should strive to use the same version for both the runtime library and the protoc-gen-go plugin used to generate the Go bindings.
```

```
The v1.20 protoc-gen-go does not support generating gRPC service definitions. 
In the future, gRPC service generation will be supported by a new protoc-gen-go-grpc plugin provided by the Go gRPC project.

The github.com/golang/protobuf version of protoc-gen-go continues to support gRPC and will continue to do so for the foreseeable future.
```

#### 四、protoc 和 protoc-gen-xxx 插件 和 grpc 和 protobuf 在选择哪个版本组合使用时，有没有推荐组合的版本号？

https://github.com/protocolbuffers/protobuf-go/releases/tag/v1.20.0#v1.20-backwards-compatibility
https://go.dev/blog/protobuf-apiv2

##### case

We call the original version of Go protocol buffers APIv1, and the new one APIv2. 
Because APIv2 is not backwards compatible with APIv1, we need to use different module paths for each.

These API versions are not the same as the versions of the protocol buffer language: proto1, proto2, and proto3.
APIv1 and APIv2 are concrete implementations in Go that both support the proto2 and proto3 language versions.

The github.com/golang/protobuf module is APIv1.
The google.golang.org/protobuf module is APIv2.

We have taken advantage of the need to change the import path to switch to one that is not tied to a specific hosting provider. 
(We considered google.golang.org/protobuf/v2, to make it clear that this is the second major version of the API, but settled on the shorter path as being the better choice in the long term.)


We know that not all users will move to a new major version of a package at the same rate. 
Some will switch quickly; others may remain on the old version indefinitely. 
Even within a single program, some parts may use one API while others use another. 
It is essential, therefore, that we continue to support programs that use APIv1.


- github.com/golang/protobuf@v1.3.4 is the most recent pre-APIv2 version of APIv1.
- github.com/golang/protobuf@v1.4.0 is a version of APIv1 implemented in terms of APIv2. 
  The API is the same, but the underlying implementation is backed by the new one. 
  This version contains functions to convert between the APIv1 and APIv2 proto.Message interfaces to ease the transition between the two.
- google.golang.org/protobuf@v1.20.0 is APIv2. 
  This module depends upon github.com/golang/protobuf@v1.4.0, 
  so any program which uses APIv2 will automatically pick a version of APIv1 which integrates with it.

Why start at version v1.20.0? 
To provide clarity. We do not anticipate APIv1 to ever reach v1.20.0, 
so the version number alone should be enough to unambiguously differentiate between APIv1 and APIv2.

We intend to maintain support for APIv1 indefinitely.

This organization ensures that any given program will use only a single protocol buffer implementation, 
regardless of which API version it uses.
It permits programs to adopt the new API gradually, or not at all, 
while still gaining the advantages of the new implementation. 

The principle of minimum version selection means that programs may remain on the old implementation until the maintainers choose to update to the new one (either directly, or by updating a dependency).



##### case 

由于 etcd 版本管理的问题，导致 etcd 的代码和新版本的 grpc 冲突,会在编译时报错:

```
undefined: resolver.BuildOption
undefined: resolver.ResolveNowOption
undefined: balancer.PickOptions
undefined: balancer.PickOptions
```

可以在 go.mod 里使用的 replace 命令指定使用 v1.26.0 老版本，可以解决 grpc 版本的问题。

```
replace google.golang.org/grpc => google.golang.org/grpc v1.26.0

protoc --go_out=plugins=grpc:. --go_opt=paths=source_relative ./your-module.proto
```

由于 protoc 的 go 语言插件 protoc-gen-go 与 google.golang.org/grpc 版本不兼容所致,
因为 grpc 降了到了 v1.26.0 ，高版本 protoc-gen-go 编译出来的 your-module.pb.go 不兼容低版本的 grpc ,
所以 protoc-gen-go 也要相应降级。

我使用二分法找到 protoc-gen-go 兼容 grpc v1.26.0 的最新版本是 v1.3.2 .

知道对应版本之后接下来就简单了，运行下面的命令获取该版本并编译二进制文件 GOPATH/bin/protoc-gen-go 

go get github.com/golang/protobuf/protoc-gen-go@v1.3.2


##### case

protoc-gen-go@v1.20 不支持生成 gRPC 服务定义。未来，gRPC 服务生成将由 Go gRPC 项目提供的新 protoc-gen-go-grpc 插件支持。

该 github.com/golang/protobufprotoc 的版本继续支持 GRPC ，将继续在可预见的将来也会如此。

##### case  

protobuf，protoc-gen-go，grpc 的兼容问题，并不是大版本号对应上就是兼容的，也并不是大版本号兼容小版本号的（向前兼容是不存在的）。

```
proto3 ProtoPackageIsVersion3
proto2 ProtoPackageIsVersion2
```

是否有对应关系？

syntax = proto3 protobuf 3.6.1 protoc-gen-go v1.2.0             ProtoPackageIsVersion2
syntax = proto3 protobuf 3.6.1 protoc-gen-go latest(2020-11-11) ProtoPackageIsVersion3


注意，如果 protoc 是 2.x.x 版本，无法处理 proto3 的文件。因此需要升级替换 protoc 为 v3.0.0 版本。

```
protoc --version
libprotoc 3.11.4
```




例如，组合的版本号为：

protoc v3.18.1
protoc-gen-go v1.27.1
protoc-gen-go-grpc v1.1.0
grpc v1.41.0
protobuf v1.27.1

or

github.com/golang/protobuf v1.5.2
google.golang.org/grpc v1.39.0
google.golang.org/protobuf v1.27.1



关于上述的版本号，有没有官方文档推荐使用的版本组合？有朋友们知道吗？欢迎留言评论 ~


### 插件
参数验证：protoc-gen-validate
参数验证：go-proto-validators
文档生成：protoc-gen-doc
grpc-gateway：
    - protoc-gen-grpc-gateway
    - protoc-gen-openapiv2
