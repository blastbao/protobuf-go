




protoc 与 其插件的工作过程：
hello.proto => protoc -> stdin -> protoc-gen-go --> CodeGeneratorRequest --> CodeGeneratorResponse -> stdout -> protoc => hello.pb.go

过程如下：

protoc 读入 hello.proto ，并把文件内容表达为 CodeGeneratorRequest 协议的 2 进制数据
protoc 把 2 进制数据 写入 stdin（管道相关知识）
protoc-gen-go 从 stdin 中读取 2 进制数据 （管道相关知识）
protoc-gen-go 解析 2 进制数据获得 CodeGeneratorRequest （内容是 descriptor.proto 中定义的元数据）
protoc-gen-go 根据 CodeGeneratorRequest ，在内存中生成 hello.pb.go 文件内容
protoc-gen-go 把生成的内容填充 CodeGeneratorResponse
protoc-gen-go 把 CodeGeneratorResponse 数据转化成 2 进程数据，写入 stdout （管道相关知识）
protoc 从 stdout 获取 2 进程数据 ，并解析得到 CodeGeneratorResponse
protoc 最后通过 CodeGeneratorResponse 获取输出文件内容，写入 hello.pb.go



一个简陋的AST定义如下:

FileDescriptor  -> ServiceDescriptor
                    -> ServiceOptionDescriptor
                    -> MethodDescriptor
                        -> MethodOptionDescriptor
                -> MessageDescriptor
                    -> MessageOptionDescriptor
                    -> FieldDescriptor
                        -> FieldOptionDescriptor
