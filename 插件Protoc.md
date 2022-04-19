

### 背景 ###

Protocol Buffers 是谷歌推出的编码标准，它在传输效率和编解码性能上都要优于 JSON。但其代价则是需要依赖中间描述语言（IDL）来定义数据（和服务）的结构（通过 *.proto 文件），并且需要一整套的工具链（protoc 及其插件）来生成对应的序列化和反序列化代码。

除了谷歌官方提供的工具和插件（比如生成 go 代码的 protoc-gen-go）外，我们还可以开发或定制自己的插件，根据业务需要按照 proto 文件的定义生成代码或者文档。

由 IDL 生成代码或者文档是元编程的一种形式，可以极大的解放程序员的生产力，建议每一位开发者都能掌握这一技能。


### 安装 protoc ###

要根据 proto 文件生成代码（或文档），首先要安装 protoc 编译工具。

protoc 可以通过操作系统的包管理器（apt/homebrew等）安装，也可以到 GitHub 上下载编译好的版本。

推荐直接下载，因为包管理器可能无法安装最新版本。

解压后有两个文件夹：
- bin 下的 protoc 直接复制 /usr/local/bin/ 目录就好
- include 下的 google 文件夹要复制到 /usr/local/include 目录

### 开发插件 ###

如果要开发某插件，我们首先需要取个名字。比如要生成 markdown 的插件，名字就定成 markdown。

如果我们想使用 markdown 插件，我们需要执行以下命令：

```protoc --markdown_out=. hello.proto```

这里有一个 --markdown_out 参数。因为我们的名字叫 markdown，所以参数名字叫 markdown_out；如果名字叫 XXX，那参数名字就叫 XXX_out。

protoc 在运行的时候首先会解析 proto 文件的所有内容，生成一组 PB 编码的描述数据，然后尝试运行 protoc-gen-markdown 命令，并且通过 stdin 将描述数据发送给插件命令。

插件基于输入的描述数据生成好文件内容后再向 stdout 输入 PB 编码的数据来告诉 protoc 生成具体的文件。

注意，插件的命令叫 protoc-gen-markdown ，插件命令统一使用 protoc-gen- 前缀，这也是 protoc 的约定。

我们刚才说 protoc 跟 protoc-gen-XXX 通信用的也是 PB 编码，是不是有点自举的意思。

PB 编码需要 proto 文件定义，protoc 跟插件通信用的 proto 文件就是 protobuf/descriptor.proto ，你可以在 /usr/include/google 或者 /usr/local/include/google 目录下找到它。而且，这个 proto 用的是 v2 版本语法！


### 开发插件 ###

理论上，现在我们就可以开发 protoc 插件了，过程也很简单：

0. 选择一种编程语言，比如 go
0. 根据 protobuf/descriptor.proto 生成对应语言的编解码代码
0. 开发 proto-gen-XXX 程序
    0. 从 stdin 读取数据并使用生成的 pb 文件解码
    0. 根据 descriptor 描述数据生成代码片断
    0. 拼装新生成的文件描述数据并通过 stdout 输出

这样做会有两个问题：
 - 直接读取 descriptor 数据非常麻烦，就连最简单的查询某字段的注释这样的功能都需要非常复杂的代码才能实现。所以说直接处理 descriptor 绝非明智之举。
 - 这些 descriptor 数据使用 v2 版本的语法，protoc 的好多语言插件（比如 PHP）已经不支持 v2 语法了。

总之，我们需要更好的框架来完成这一工作。

protoc-gen-markdown 早期使用 github.com/pseudomuto/protokit 来简化开发工作。可以提取注释，但不支持直接查询内嵌消息的结构，需要大量的辅助代码。protoc-gen-markdown 的代码也很难维护（除了我很少有人能改）。

直到去年三月份，谷歌发布了新版 Go 语言 API，其中引入了一个新包 `google.golang.org/protobuf/compiler/protogen` 可以大幅简化插件的开发。

我这两天基于 protogen 重构了 protoc-gen-markdown 的实现，所有才好意思拿出来跟大家分享😭

插件代码的主要结构如下：

```
func main() {
  g := markdown{}
  var flags flag.FlagSet
  flags.StringVar(&g.Prefix, "prefix", "/", "API path prefix")

  protogen.Options{
    ParamFunc: flags.Set,
  }.Run(g.Generate)
}

func (md *markdown) Generate(plugin *protogen.Plugin) error {
  // ...
}
```
首先我们需要定义一个 protogen.Options ，然后调用它的 Run 方法，并传入一个 func(*protogen.Plugin) error 回调，主流程代码到此就结束了。

我们还可以设置 protogen.Options 的 ParamFunc 参数，这样 protogen 会自动为我们解析命令行传入的参数。

诸如从标准输入读取并解码 protobuf 信息，将输入信息编码成 protobuf 写入 stdout 等操作全部由 protogen 包办了。

我们要做的就是与 protogen.Plugin 交互实现代码生成逻辑。

如果大家对这种简化没有概念，可以去翻一下 protoc-gen-markdown 最初的代码。

没有对比就没有伤害。

接下来，我们把注意力转向 Generate 方法。读代码（尤其是读别人的代码）是枯燥的，我尽量说思路。

要想生成接口文档，我们肯定是要遍历每一个 proto 文件的每一个 service 的每一个 rpc 定义，所以主流程是这样的：

```
func (md *markdown) Generate(plugin *protogen.Plugin) error {
  for _, f := range plugin.Files {
    
    // 跳过没有定义 service 的 proto 文件
    if len(f.Services) == 0 {
      continue
    }
    
    // 指定要生成的文件名，策略可以根据需要指定
    fname := f.GeneratedFilenamePrefix + ".md"
    
    // 告诉 protoc 要生成一个新文件，并获取一个引用，后面可以通过 t 来写入新文件的内容
    t := plugin.NewGeneratedFile(fname, f.GoImportPath)
    
    // 遍历每一个 service
    for _, s := range f.Services {
      // ...
      for _, m := range s.Methods {
        // ...
      }
      // ...
    }
  }
}
```

对于 proto 的每一级结构，protogen 都有对应的数据类型，分别是 File, Service, Method, Field 。

除了 File 外，其他数据类型都有 Comments 属性，每个 Comments 又包含 LeadingDetached, Leading, Trailing 三种类型的注释，效果如下：

```
// LeadingDetached

// Leading
message Foo {} // Trailing
```

生成文档就需要恰当处理这些注释，而注释跟被描述对象的关联关系 protogen 已经帮我们处理好了。

除了注释，我们还需要读取更多信息，比如消息或者字段的名字、类型等，这些需要通过每个对象的 Desc 字段读取，它的类型是 protoreflect.ServiceDescriptor，跟我们前面说的 descriptor.proto 对应。

生成的文档结构如下，逻辑见注释：

```
for _, s := range f.Services {
  // 输出 service 名称
  t.P("# ", s.Desc.Name())
  t.P()
  
  // 输出 service 的注释
  t.P(string(s.Comments.Leading))
  
  // 生成接口目录
  for _, m := range s.Methods {
    // ...
    t.P(fmt.Sprintf("- [%s](#%s)", api, anchor))
  }
  
  t.P()
  for _, m := range s.Methods {
  
    // 输出接口路径
    t.P("## ", md.api(string(m.Desc.FullName())))
    t.P()
    
    // 输出接口注释
    t.P(string(m.Comments.Leading))
    t.P()
    
    // 输出入参结构，json 格式
    t.P("### Request")
    
    // 语言指定为 javacript，这样可以在 json 里加入注释内容
    t.P("```javascript")
    
    // 每个 method 都有 Input 和 Output 两个字段
    t.P(md.jsDocForMessage(m.Input))
    t.P("```")
    t.P()
    
    // 输出入参结构，json 格式
    t.P("### Reply")
    t.P("```javascript")
    t.P(md.jsDocForMessage(m.Output))
    t.P("```")
  }
}
```

我们是想用 json 表示每一个 Protocol Buffers 结构。但 json 不能添加注释，所以我把语言指定为 javacript ，这样既能使用 json 又能加字段注释了。

但为什么用 json 呢？因为 Protocol Buffers 官方定义了 json 的表示格式，我们的业务框架同时支持 protobuf 和 json 两种编码。使用 json 表达不会丢失定义信息，而且阅读也稍微方便一些。

每个 method 都有 Input 和 Output 两个字段，分别表示接口的入参和出参。下面开始讨论如何提取每个字段的信息。

在开始之前，我们先简单说一下 markdown 消息输出效果（这部分是我自己定义的，仅供大家参考）。

对于简单类型字段，比如

```
// 字段注释
int32 a = 1; // 行尾注释
double b = 2;
string c = 3;
```
我会转换成（这是合法的 javascript 代码，类 json）

```
// 字段注释
a: 0, // type<int32>, 行尾注释
b: 0.0,
c: "",
```

字段名保持不变，字段值根据类型自动生成，一般取对应的类型的空值（整数类的为 0 ，字符串为 "" ，布尔值为 false ，浮点数为 0.0 等，但 64 位整数对应 "0" ，这个也是遵守 Protocol Buffers 与 json 的官方映射规则）。

然后会给每个字段添加类型注释，格式是 type<T> ，这里的 T 对应 proto 文件的类型声明。如果 proto 字段本身也有行为注释，则会追加到类型注释之后。

如果字段的类型是 message，则会转换成 json 的 object 语法，并且类型注释 type<MessageName> 。消息的每个字段会按照前面说的字段规则转换。

比如：
```
Foo a = 1;
```
会变成：
```
a: {
  // ...
} // type<Foo>
```
如果是 repeated 类型的字段，则会在转换之后的结果两侧添加 [] ，然后再追加类型注释 list<T> 。这里 T 可以是普通类型，也可以是 message 类型。

比如：

```
repeated int32 a = 1;
repeated Foo b = 2;
```

会变成：

```
a: {
  "0": ""
}, // map<in32,string>
```

protoc-gen-markdown 对于枚举类型会直接转换成字符串，值为枚举的默认值（零值），同时追加类型注释enum<E1,E2,...>。

比如：

```
enum Platform {
  All = 0;
  Web = 1;
  Ios = 2;
  Android = 3;
}
Platform a = 1;
```
会转化成：

```
a: "All", // enum<All,Web,Ios,Android>
```

因为 proto 支持嵌套定义，所以我们需要从最外层的 message 开始递归处理所有字段。

对于每一个 message，我定义了一个方法：

```
func (md *markdown) jsDocForMessage(m *protogen.Message) string {
  js := "{\n"
  for _, field := range m.Fields {
    js += md.jsDocForField(field)
  }
  js += "}"
  return js
}
```

主要是生成最外面的 "{}" ，然后是处理每一个字段。jsDocForField 的整体结构如下：

```
func (md *markdown) jsDocForField(field *protogen.Field) string {
  // 追加字段注释
  js := field.Comments.Leading.String()
  // 生成字段名
  js += string(field.Desc.Name()) + ":"
  // 确定字段值的取值和类型
  var vv,vt string
  if field.Desc.IsMap() {
  } else if field.Message != nil {
  } else if field.Enum != nil {
  } else if field.Oneof != nil {
  } else {
  }
  // 添加类型注释
  if field.Desc.IsList() {
    js += fmt.Sprintf("[%s], // list<%s>", vv, vt)
  } else if field.Desc.IsMap() {
    js += vv + fmt.Sprintf(", // map<%s>", vt)
  } else if field.Enum != nil {
    js += vv + fmt.Sprintf(", // enum<%s>", vt)
  } else {
    js += vv + fmt.Sprintf(", // type<%s>", vt)
  }
  // 追加行尾注释
  return js
}
```

对于简单字段，可以直接调用 scalarDefaultValue 确定默认值。而对应的类型可以直接通过 field.Desc.Kind().String() 获取。这个比较简单。

Protocol Buffers 在内部会把 map 转换成下面这样的 message，

```
message MapFieldEntry {
  key_type key = 1;
  value_type value = 2;
}
```

所以我们需要先处理 map 类型的字段，再处理 message 类型的字段。protogen 提供了 IsMap 方法供我们使用。如果 IsMap 返回 false，而 filed.Message 不为 nil ，那才 message 类型字段。

对于 message 类型的字段也很简单，直接递归调用 jsDocForMessage 获取取值的 json 表示，其类型可以通过 field.Message.Desc.Name() 获得。

如果是 map 类型，我们可以通过 .Message 字段取到 MapFieldEntry 消息。然后根据 Fields[0] 和 Fields[1] 构造 map 的取值和类型。

对于 Enum 类型，可以直接通过 .Enum 读取相关信息，也比较简单。

另外，我们需要美化一下 json 的格式，主要是调整缩进。

本身处理 protogen 消息已经很复杂了，如果还要手工维护缩进只会难上加难。

于是我找了一个 go 语言实现的 javascript 格式化库 jsbeautifier，在消息返回之前统一格式化一下：

```
options := jsbeautifier.DefaultOptions()
js, _ = jsbeautifier.Beautify(&js, options)
```

因为我们是递归实现的，如果在 proto 文件中使用了递归定义，比如：

```
message Foo {
  int32 a = 1;
  Foo b = 2;
}
```

不加处理的话会导致栈溢出错误。

为此，我维护了一个栈，保存当前正在处理消息，如果碰到递归定义，则直接返回{}。

如此一来，刚才的定义会转化成：

```
{
  a: 0,
  b: {}, // type<Foo>
}
```

以上就是本文的全部内容了。本文定义了一套从 proto 生成 markdown 文档的规则并基于最新的google.golang.org/protobuf/compiler/protogen包实现了 protoc-gen-markdown 插件。

阅读本文和 protoc-gen-markdown 源码，可以掌握 protoc 插件开发的基础知识。

希望每一位读者都能掌握这一技能。



### 参考文献 ###

- https://taoshu.in/go/create-protoc-plugin.html
- https://github.com/go-kiss/protoc-gen-markdown

















































