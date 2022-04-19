### 背景

#### protoc 
protoc 是 protobuf 文件（.proto）的编译器，可以借助这个工具把 .proto 文件转译成各种编程语言对应的源码，包含数据类型定义、调用接口等。

protoc 在设计上把 protobuf 和不同的语言解耦了，底层用 c++ 来实现 protobuf 结构的存储，然后通过插件的形式来生成不同语言的源码。

可以把 protoc 的编译过程分成简单的两个步骤：

1）解析 .proto 文件，转译成 protobuf 的原生数据结构在内存中保存；
2）把 protobu f相关的数据结构传递给相应语言的编译插件，由插件负责根据接收到的 protobuf 原生结构渲染输出特定语言的模板。

#### protoc-gen-go
protoc-gen-go 是 protobuf 编译插件系列中的 Go 版本，原生的 protoc 并不包含 Go 版本的插件，不过可以在 github 上发现专门的代码库。
由于 protoc-gen-go 是 Go 写的，所以安装它变得很简单，只需要运行 go get -u github.com/golang/protobuf/protoc-gen-go，便可以在 $GOPATH/bin 目录下发现这个工具。
至此，就可以通过下面的命令来使用 protoc-gen-go 了。

``` protoc --go_out=output_directory input_directory/file.proto ```

其中 "--go_out=" 表示生成 Go 文件，protoc 会自动寻找 PATH（系统执行路径）中的 protoc-gen-go 执行文件。

#### 总结
protobuf 的出现，为不同系统之间的连接提供了一种语言规范，只要遵循了这个规范，各个系统之间就是解耦的，非常适合近年来流行的微服务架构。

如果把 protoc 和 protoc-gen-go 看成两个微服务，可以发现这两个服务就是完全解耦的；

两者完全负责不同的功能，可以分别编码、升级，串接这两个服务的就是 proto 规范。

### 背景

将 proto 文件生成 go 代码的命令 

```protoc --go_out=. ./*.proto```

protoc 执行这个会生成 go 代码，这个原理是什么呢？

还记得我们装完 protoc 并不能马上生成代码，还需要装一个东西，命令如下

```go install github.com/golang/protobuf/protoc-gen-go@latest```

当 protoc 执行命令的时候，插件解析步骤如下

1. 解析 proto 文件，类似于 AST 树的解析，将整个 proto 文件有用的语法内容提取出来
2. 将解析的结果转为成二进制流，然后传入到 protoc-gen-xx 标准输入。也就是说 protoc 会去程序执行路径去找 protoc-gen-go 这个二进制，将解析结果写入到这个程序的标准输入。 如果命令是 --go_out ，那么找到是 protoc-gen-go ，如果自己定义一个插件叫 xxoo ，那么 --xxoo_out 找到就是 protoc-gen-xxoo 了。所以我们得确保有这个二进制执行文件，要不然会报找不到的错误。
3. protoc-gen-xx 必须实现从标准输入读取上面解析完的二进制流，然后将得到的 proto 信息按照自己的语言生成代码，最后将输出的代码写回到标准输出
4. protoc 接收 protoc-gen-xx 的标准输出，然后写入文件

实际调用过程中，可以将插件二进制文件放在环境变量 PATH 中，也可直接通过 --plugin=[plugin-name=plugin-path, ....] 传给 protoc

```protoc --plugin=protoc-gen-xx=path/to/xx --NAME_out=MY_OUT_DIR```

看到上面的步骤可能有点懵，下面通过例子来熟悉整个流程


### 为每个 proto message 添加自定义方法

#### 安装 protoc

注意这里也要安装 protoc-gen-go ，这个 demo 目的是在原生的 xx.pb.go 的文件下拓展，为每个消息添加额外的方法，后面会讲 protoc-gen-go 的源码解析及自定义开发

#### 创建项目目录
- demo 为项目目录
- proto 存放 proto 文件的地方
- out 为 proto 编译后生成的 go 文件目录
- go.mod 使用 go mod init protoc-gen-foo 命令生成
- main.go 项目主文件
- Makefile 项目执行脚本，每次敲命令难受，所以直接写 Makefile

#### 写一个消息

demo/proto/demo.proto

```
syntax = "proto3";
package test;
option go_package = "/test";

message User {
  //用户名
  string Name = 1;
  //用户资源
  map<int32,string> Res=2 ;
}
```

#### 写 main.go 解析这个文件

```
package main
import (
    "bytes"
    "fmt"
    "google.golang.org/protobuf/compiler/protogen"
    "google.golang.org/protobuf/types/pluginpb"
    "google.golang.org/protobuf/proto"
    "io/ioutil"
    "os"
)
func main()  {
    // 1. 读取标准输入，接收 proto 解析的文件内容，并解析成结构体
    input, _ := ioutil.ReadAll(os.Stdin)
    var req pluginpb.CodeGeneratorRequest
    proto.Unmarshal(input, &req)
    
    // 2. 生成插件
    opts := protogen.Options{}
    plugin, err := opts.New(&req)
    if err != nil {
        panic(err)
    }
    
    // 3. 在插件 plugin.Files 就是 demo.proto 的内容了,是一个切片，每个切片元素代表一个文件内容，我们只需要遍历这个文件就能获取到文件的信息了
    for _, file := range plugin.Files {
        
        // 创建一个 buf 写入生成的文件内容
        var buf bytes.Buffer
        
        // 写入 go 文件的 package 名
        pkg := fmt.Sprintf("package %s", file.GoPackageName)
        buf.Write([]byte(pkg))
        
        // 遍历消息, 这个内容就是 protobuf 的每个消息
        for _, msg := range file.Messages {
            
            //接下来为每个消息生成 hello 方法
            buf.Write([]byte(fmt.Sprintf(`
             func (m*%s)Hello(){
             }
            `, msg.GoIdent.GoName)))
        }
        
        // 指定输入文件名, 输出文件名为 demo.foo.go
        filename := file.GeneratedFilenamePrefix + ".foo.go"
        file := plugin.NewGeneratedFile(filename, ".")
        
        // 将内容写入插件文件内容
        file.Write(buf.Bytes())
    }
    
    // 生成响应
    stdout := plugin.Response()
    out, err := proto.Marshal(stdout)
    if err != nil {
        panic(err)
    }
    
    // 将响应写回标准输入, protoc 会读取这个内容
    fmt.Fprintf(os.Stdout, string(out))
}
```
#### 安装自己的插件
在 demo 路径下执行下面命令，将生成自定义 proto 插件

```
demo> go install .
demo> ls $GOPATH/bin | grep protoc
protoc-gen-foo //自定义
protoc-gen-go
```

#### 使用自定义插件生成 pb 文件

```demo> protoc --foo_out=./out --go_out=./out ./proto/*.proto```

生成结果
 - demo.pb.go 是 protoc-gen-go 生成的文件
 - demo.foo.go 是我们自定义插件生成的文件

demo.pb.go

```
package test
func (m *User) Hello() {
}
```
至此，一个简单的插件完工

#### 优化命令，写入 Makefile

如果每次修改插件都要敲几个命令太难受了。所以我们需要写个脚本自动化做这个事情

```
p:
    protoc --foo_out=./out --go_out=./out ./proto/*.proto
is:
    go install .
all:
    make is
    make p
```
测试下
```
demo> make all
make is
go install .
make p
protoc    --foo_out=./out --go_out=./out ./proto/*.proto
```

### 进阶

接下来我们实现一个 map 的拷贝方法，假设项目中有这种场景，我们需要将 go 结构体传给别人，但是这个 map 如果被别人引用，在不同的协程中修改，这可能就悲剧了。

map 并发的 panic 是无法恢复的，如果线上的话，程序员需要拿去祭天了。

假设结构体是 protobuf 的消息结构体，我们实现了一个 CheckMap 方法，别人拿到消息结构体，如果里面有 map 的字段就会有这个方法。

在发送给别人时统一调用下，然后拷贝，就不会有问题了。

我们修改内容，如下 main.go
```
package main
import (
   "bytes"
   "fmt"
   "google.golang.org/protobuf/compiler/protogen"
   "google.golang.org/protobuf/types/pluginpb"
   "google.golang.org/protobuf/proto"
   "io/ioutil"
   "os"
)

func main()  {
   //1.读取标准输入，接收proto 解析的文件内容，并解析成结构体
   input, _ := ioutil.ReadAll(os.Stdin)
   var req pluginpb.CodeGeneratorRequest
   proto.Unmarshal(input, &req)
   
   //2.生成插件
   opts := protogen.Options{}
   plugin, err := opts.New(&req)
   if err != nil {
      panic(err)
   }
   
   // 3.在插件plugin.Files就是demo.proto 的内容了,是一个切片，每个切片元素代表一个文件内容
   // 我们只需要遍历这个文件就能获取到文件的信息了
   for _, file := range plugin.Files {
      
   	  //创建一个buf 写入生成的文件内容
      var buf bytes.Buffer
      
      // 写入go 文件的package名
      pkg := fmt.Sprintf("package %s", file.GoPackageName)
      buf.Write([]byte(pkg))
      content:=""
      
      // 遍历消息,这个内容就是 protobuf 的每个消息
      for _, msg := range file.Messages {
         
         mapSrc:=`
            newMap:=make(map[%v]%v)
            for k,v:=range x.%v {
                newMap[k]=v
            }
            x.%v=newMap
         `
         
         // 遍历消息的每个字段
         for _,field:=range msg.Fields{
            // 只有 map 才这样做
            if field.Desc.IsMap(){
               content += fmt.Sprintf(
                mapSrc,
                field.Desc.MapKey().Kind().String(),
                field.Desc.MapValue().Kind().String(), 
                field.GoName,
                field.GoName
               )
            }
         }
         
         buf.Write([]byte(fmt.Sprintf(`
           func (x *%s) CheckMap() {
            %s
         }`, msg.GoIdent.GoName, content)))
      }
      
      // 指定输入文件名,输出文件名为 demo.foo.go
      filename := file.GeneratedFilenamePrefix + ".foo.go"
      file := plugin.NewGeneratedFile(filename, ".")
      
      // 将内容写入插件文件内容
      file.Write(buf.Bytes())
   }
   
   // 生成响应
   stdout := plugin.Response()
   out, err := proto.Marshal(stdout)
   if err != nil {
      panic(err)
   }
   
   // 将响应写回标准输入, protoc会读取这个内容
   fmt.Fprintf(os.Stdout, string(out))
}
```

编译

```demo >make all```

生成结果

demo/out/test/demo.foo.go

```
package test
func (x *User) CheckMap() {
    newMap := make(map[int32]string)
    for k, v := range x.Res {
        newMap[k] = v
    }
    x.Res = newMap
}
```



### 消息结构体大体了解


#### file 结构体

```
// A File describes a .proto source file.
type File struct {
   Desc  protoreflect.FileDescriptor
   Proto *descriptorpb.FileDescriptorProto

   GoDescriptorIdent GoIdent       // name of Go variable for the file descriptor
   GoPackageName     GoPackageName // name of this file's Go package
   GoImportPath      GoImportPath  // import path of this file's Go package

   Enums      []*Enum      // top-level enum declarations
   Messages   []*Message   // top-level message declarations
   Extensions []*Extension // top-level extension declarations
   Services   []*Service   // top-level service declarations

   Generate bool // true if we should generate code for this file

   // GeneratedFilenamePrefix is used to construct filenames for generated
   // files associated with this source file.
   //
   // For example, the source file "dir/foo.proto" might have a filename prefix
   // of "dir/foo". Appending ".pb.go" produces an output file of "dir/foo.pb.go".
   GeneratedFilenamePrefix string
   
   location Location
}
```
其中：
- Enums 代表文件中的枚举
- Messages 代表 proto 文件中的所有消息
- Extensions 代表文件中的扩展信息
- Services 消息中定义的服务，跟 grpc 有关
- GeneratedFilenamePrefix 来源 proto 文件的前缀，上面有例子，例如 "dir/foo.proto" ，这个值就是 "dir/foo" ，后面加上 ".pb.go" 代表最后生成的文件
- Comments 字段代表字段上面，后面的注释

proto 消息 Message


```
// A Message describes a message.
type Message struct {
   Desc protoreflect.MessageDescriptor

   GoIdent GoIdent // name of the generated Go type

   Fields []*Field // message field declarations
   Oneofs []*Oneof // message oneof declarations

   Enums      []*Enum      // nested enum declarations
   Messages   []*Message   // nested message declarations
   Extensions []*Extension // nested extension declarations

   Location Location   // location of this message
   Comments CommentSet // comments associated with this message
}
```

- Fields 代表消息的每个字段，遍历这个字段就可以得到字段信息
- Oneofs 代表消息中 Oneof 结构
- Enums 消息中的枚举
- Messages 嵌套消息，消息是可以嵌套消息，所以这个代表嵌套的消息
- Extensions 代表扩展信息
- Comments 字段代表字段上面，后面的注释

消息字段 Field

```
// A Field describes a message field.
type Field struct {
   Desc protoreflect.FieldDescriptor

   // GoName is the base name of this field's Go field and methods.
   // For code generated by protoc-gen-go, this means a field named
   // '{{GoName}}' and a getter method named 'Get{{GoName}}'.
   GoName string // e.g., "FieldName"

   // GoIdent is the base name of a top-level declaration for this field.
   // For code generated by protoc-gen-go, this means a wrapper type named
   // '{{GoIdent}}' for members fields of a oneof, and a variable named
   // 'E_{{GoIdent}}' for extension fields.
   GoIdent GoIdent // e.g., "MessageName_FieldName"

   Parent   *Message // message in which this field is declared; nil if top-level extension
   Oneof    *Oneof   // containing oneof; nil if not part of a oneof
   Extendee *Message // extended message for extension fields; nil otherwise

   Enum    *Enum    // type for enum fields; nil otherwise
   Message *Message // type for message or group fields; nil otherwise

   Location Location   // location of this field
   Comments CommentSet // comments associated with this field
}
```

- Desc 代表该字段的描述
- GoName 代表字段名
- Parent 代表父消息
- Comments 字段代表字段上面，后面的注释

### Options

可用的选项列表在 google/protobuf/descriptor.proto 
其它选项官方有，是其它语言相关的，这里就不细讲了，看官方文档 Options 。

#### Custom Options
protobuffer 提供大多人都不会使用的高级功能-自定义选项。
由于选项是由 google/protobuf/descriptor.proto（如 FileOptions 或 FieldOptions）中定义的消息定义的，因此定义你自己的选项只需要扩展这些消息。

如何定义一个选项？

```
import "google/protobuf/descriptor.proto";

extend google.protobuf.MessageOptions {
  optional string my_option = 51234;
}

message MyMessage {
  option (my_option) = "Hello world!";
}
```

获取选项

```
package main

import (
   "fmt"
   "grpcdemo/protobuf/any/protos/pbs"
)

func main()  {
   p:=&pbs.MyMessage{}

   fmt.Println(p.ProtoReflect().Descriptor().Options())
   //[my_option]:"Hello world!"
}
```

Protocol Buffers 可以为每种类型提供选项

```
import "google/protobuf/descriptor.proto";

extend google.protobuf.FileOptions {
  optional string my_file_option = 50000;
}

extend google.protobuf.MessageOptions {
  optional int32 my_message_option = 50001;
}

extend google.protobuf.FieldOptions {
  optional float my_field_option = 50002;
}

extend google.protobuf.OneofOptions {
  optional int64 my_oneof_option = 50003;
}

extend google.protobuf.EnumOptions {
  optional bool my_enum_option = 50004;
}

extend google.protobuf.EnumValueOptions {
  optional uint32 my_enum_value_option = 50005;
}

extend google.protobuf.ServiceOptions {
  optional MyEnum my_service_option = 50006;
}

extend google.protobuf.MethodOptions {
  optional MyMessage my_method_option = 50007;
}

option (my_file_option) = "Hello world!";

message MyMessage {
  option (my_message_option) = 1234;

  optional int32 foo = 1 [(my_field_option) = 4.5];
  optional string bar = 2;
  oneof qux {
    option (my_oneof_option) = 42;
    
    string quux = 3;
  }
}

enum MyEnum {
  option (my_enum_option) = true;

  FOO = 1 [(my_enum_value_option) = 321];
  BAR = 2;
}

message RequestType {}
message ResponseType {}

service MyService {
  option (my_service_option) = FOO;

  rpc MyMethod(RequestType) returns(ResponseType) {
    // Note:  
    //   my_method_option has type MyMessage. 
    //   We can set each field within it using a separate "option" line.
    option (my_method_option).foo = 567;
    option (my_method_option).bar = "Some string";
  }
}
```
引用其他包的选项需要加上包名

```

// foo.proto
import "google/protobuf/descriptor.proto";
package foo;
extend google.protobuf.MessageOptions {
  optional string my_option = 51234;
}

// bar.proto
import "foo.proto";
package bar;
message MyMessage {
  option (foo.my_option) = "Hello world!";
}

```
自定义选项是扩展名，必须分配字段号，像上面的例子一样。

在上面的示例中，使用了 50000-99999 范围内的字段编号。这个字段范围供个人组织使用，所以可以内部用。

在公共应用使用的话，要保持全球唯一数字，需要申请，申请地址为：

protobuf global extension registry

通常只需要一个扩展号，可以多个选项放在子消息中来实现一个扩展号声明多个选项

```
message FooOptions {
  optional int32 opt1 = 1;
  optional string opt2 = 2;
}

extend google.protobuf.FieldOptions {
  optional FooOptions foo_options = 1234;
}

// usage:
message Bar {
  optional int32 a = 1 [(foo_options).opt1 = 123, (foo_options).opt2 = "baz"];
  // alternative aggregate syntax (uses TextFormat):
  optional int32 b = 2 [(foo_options) = { opt1: 123 opt2: "baz" }];
}
```

每种选项类型（文件级别，消息级别，字段级别等）都有自己的数字空间，例如：

可以使用相同的数字声明 FieldOptions 和 MessageOptions 的扩展名。









#### 实践

##### 1、创建 option 

自定义的选项需要扩展 google.protobuf.FieldOptions 

```
syntax = "proto2";
package ext;
option go_package = "/my_ext";
import "google/protobuf/descriptor.proto";

extend google.protobuf.FieldOptions {
  optional bool flag = 65004;
  optional string jsontag = 65005;
}

```

修改 demo.proto，添加自定义选项

```
syntax = "proto3";
package test;
option go_package = "/test";
import "proto/ext/my_ext.proto";

message User {
  // 用户名
  string Name = 1 [(ext.flag)=true]; //自定义选项: 添加选项 flag
  // 用户资源
  map<int32,string> Res=2 [(ext.jsontag)="res"]; //自定义选项: jsontag
}

```
上面的 [] 内容就是我们自定义的选项

##### 2、编译

下面修改 Makefile , 将 ext 文件夹也添加进去，因为 my_ext.proto 定义在其中。

```
p:
   protoc --foo_out=./out --go_out=./out ./proto/*.proto ./proto/ext/*.proto
is:
   go install .
all:
   make is
   make p
```

接下来来编译

```
demo>make all
```
在 my_ext.pb.go 有两句重要代码，我们等会会用到。

```
// Extension fields to descriptorpb.FieldOptions.
var (
   // optional bool flag = 65004;
   E_Flag = &file_proto_ext_my_ext_proto_extTypes[0]
   // optional string jsontag = 65005;
   E_Jsontag = &file_proto_ext_my_ext_proto_extTypes[1]
)
```

##### 3、在自定义插件获取选项的值

我们修改 main.go ，将选项值生成出来

```
package main
import (
    "bytes"
    "fmt"
    "google.golang.org/protobuf/compiler/protogen"
    "google.golang.org/protobuf/proto"
    "google.golang.org/protobuf/types/descriptorpb"
    "google.golang.org/protobuf/types/pluginpb"
    "io/ioutil"
    "os"
    "protoc-gen-foo/out/my_ext"
)

func main()  {
    //1.读取标准输入，接收proto 解析的文件内容，并解析成结构体
    input, _ := ioutil.ReadAll(os.Stdin)
    var req pluginpb.CodeGeneratorRequest
    proto.Unmarshal(input, &req)
    
    //2.生成插件
    opts := protogen.Options{}
    plugin, err := opts.New(&req)
    if err != nil {
        panic(err)
    }

    // 3.在插件plugin.Files就是demo.proto 的内容了,是一个切片，每个切片元素代表一个文件内容
    // 我们只需要遍历这个文件就能获取到文件的信息了
    for _, file := range plugin.Files {
        //创建一个buf 写入生成的文件内容
        var buf bytes.Buffer

        // 写入go 文件的package名
        pkg := fmt.Sprintf("package %s", file.GoPackageName)
        buf.Write([]byte(pkg))
        context:=""
       
        //遍历消息,这个内容就是protobuf的每个消息
        for _, msg := range file.Messages {
           
            // 遍历消息的每个字段
            for _,field:=range msg.Fields{
                // 提取 option 
                op, ok:=field.Desc.Options().(*descriptorpb.FieldOptions)
                if ok{
                    value := GetJsonTag(op)
                    context+=fmt.Sprintf("%v\n",value)
                    value = GetFlagTag(op)
                    context += fmt.Sprintf("%v\n",value)
                }
            }
            
            buf.Write([]byte(fmt.Sprintf(`
                func (x *%s) optionsTest() {
                %s
                }`, msg.GoIdent.GoName, context)))
        }
        
        // 指定输入文件名,输出文件名为demo.foo.go
        filename := file.GeneratedFilenamePrefix + ".foo.go"
        file := plugin.NewGeneratedFile(filename, ".")

        // 将内容写入插件文件内容
        file.Write(buf.Bytes())
    }

    // 生成响应
    stdout := plugin.Response()
    out, err := proto.Marshal(stdout)
    if err != nil {
        panic(err)
    }

    // 将响应写回标准输入, protoc会读取这个内容
    fmt.Fprintf(os.Stdout, string(out))
}

// 提取 json tag 
func GetJsonTag(field *descriptorpb.FieldOptions) interface{} {
    if field == nil {
        return ""
    }
    v:= proto.GetExtension(field,my_ext.E_Jsontag)
    return v.(string)
}

// 提取 flag tag 
func GetFlagTag(field *descriptorpb.FieldOptions) interface{} {
    if field == nil {
        return ""
    }
    if !proto.HasExtension(field, my_ext.E_Flag){
        return ""
    }
    v:= proto.GetExtension(field,my_ext.E_Flag)
    return v.(bool)
}
```

编译

```
demo>make all
```

结果
demo/out/test/demo.foo.go

```
func (x *User) optionsTest() {
true //flag 的值
res //jsontag 的值
}
```

这个例子只是展示如何获取选项的值，并不是完整的 demo 。

##### 4、内容优化

因为扩展都是固定，提前定义好的，所以我们并不需要和业务 proto 一起编译，
我们将扩展文件夹的结构变为如下，将刚刚生成的扩展文件移动到 ext 里面来。

helper.go 内容

```
package ext

import (
   "google.golang.org/protobuf/proto"
   "google.golang.org/protobuf/types/descriptorpb"
)

//获取JsonTag 的值
func GetJsonTag(field *descriptorpb.FieldOptions) interface{} {
   if field == nil {
      return ""
   }
   v:= proto.GetExtension(field, E_Jsontag)
   return v.(string)
}

//获取flag 的值
func GetFlagTag(field *descriptorpb.FieldOptions) interface{} {
   if field == nil {
      return ""
   }
    //判断字段有没有这个选项
   if !proto.HasExtension(field, E_Flag){
      return ""
   }
    //有的话获取这个选项值
   v:= proto.GetExtension(field, E_Flag)
   return v.(bool)
}
```

main.go

```
            .....
            // 遍历消息的每个字段
            for _,field:=range msg.Fields{
                op,ok:=field.Desc.Options().(*descriptorpb.FieldOptions)
                if ok{
                    value:=ext.GetJsonTag(op)
                    context+=fmt.Sprintf("%v\n",value)
                    value=ext.GetFlagTag(op)
                    context+=fmt.Sprintf("%v\n",value)
                }
                ...
            }
            ...
```

##### 5、文件优化

经过上面的操作，读者可能有疑惑，多出来的几个文件是错误的，怎么去掉多余的文件。

文件 1 出现的原因是系统的默认消息也被加进来了，protoc-gen-go 是不会解析这个文件，但是我们的插件没有过滤
文件 2 出现的原因是执行命令，protoc 把所有文件都解析出来，包括 ext 里面的 proto 文件，但是 protoc-gen-go 在解析时并不会递归解析子文件，原因是也做了过滤操作，所以我们也可以这样做。

```protoc --foo_out=./out --go_out=./out ./proto/*.proto```

解决这个问题，我们只需要将这些文件忽略掉

在循环遍历 file 的时候，continue，如下，跳过不该出现的文件

```
    // 3. 在插件 plugin.Files 就是 demo.proto 的内容了, 是一个切片，每个切片元素代表一个文件内容
    // 我们只需要遍历这个文件就能获取到文件的信息了
    for _, file := range plugin.Files {
        if strings.Contains(file.Desc.Path(),"my_ext.proto"){
            continue
        }
        if strings.Contains(file.Desc.Path(),"descriptor.proto"){
            continue
        }
        ....
    }
```
再生成一下，可以发现到我们要的目的了


### protoc-gen-go 介绍

相信经过上面的实践后，大家对 protoc 插件有更深刻的认识了，如果我们使用命令如下

```protoc --foo_out=./out  ./proto/*.proto```

那么结局就只会生成一个文件了，demo.foo.go

所以 protoc-gen-go 是一个实现更完善的插件，整个解析流程是一样的，只不过在生成 pb 文件做了更多事情。
如果你想更加深入学习的话，那么就可以去官方下载源码找自己感兴趣的部分看看了。

下期将带来 protoc-gen-go源码解析，如果大家喜欢的话不要忘记点赞。


### 参考
Writing a protoc plugin with google.golang.org/protobufhttps://pkg.go.dev/mod/google.golang.org/protobuf
github.com/gogo/protobuf/protoc-gen-gogo
https://github.com/golang/protobuf
Protocol Buffers






















