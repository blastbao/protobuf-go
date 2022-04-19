

### 

https://www.infoq.com/news/2020/03/go-protobuf-apiv2/

根据新 protobuf 模块版本的作者 Joe Tsai、Damien Neil 和 Herbie Ong 的说法，之前的 protobuf 实现已经不能满足 Go 开发人员的期望了。

具体来说，它虽然提供了 Go 类型和值的视图，但是忽略了 protocol buffer 类型系统中的信息。

这样做的后果就是 portobuf 注解的丢失。

例如，我们可能想编写一个遍历日志条目并清除任何注解（annotation）为包含敏感数据的字段的函数。

注解不是 Go 类型系统的一部分。

旧的 protobuf 模块的另一个局限是依赖于静态绑定，从而阻碍了动态消息的使用，而动态消息的类型在编译时不是完全可知的。

新的 protobuf 模块 （版本为 APIv2）基于这个假设：

protobuf Message 必须完全指定消息的行为，并使用反射以提供 protobuf 类型的完整视图。

Go protobuf APIv2 的基石是新的 proto.Message 接口，可用于所有生成的消息类型，并提供了访问消息内容的方式。

这包括所有 protobuf 字段，可以使用 protoreflect.Message.Range 方式对它们进行迭代。

该方法既可以处理动态消息，也可以访问消息选项。

下面例子说明如何处理消息以在进一步处理前清除其包含的所有敏感信息：

```
// Redact清除pb中的每个敏感字段。
func Redact(pb proto.Message) {
    m := pb.ProtoReflect()
    m.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
        opts := fd.Options().(*descriptorpb.FieldOptions)
        if proto.GetExtension(opts, policypb.E_NonSensitive).(bool) {
            return true
        }
        m.Clear(fd)
        return true
    })
}
```

然而，几个Hacker News评论者指出，Go protobuf APIv2的版本控制有点让人感到困惑 。
开发人员需要把新版本绑定到一个特定的存储库，而不能使用新的版本扩展现有的存储库，并把它标记为 v2。

Damien Neil 解释了这一决定背后的原因，如下所示：

```
我们可以把新的 API 标记为 v2：

在导入路径中，把 v1 和 v2 清楚地区分开来。

让人感到困惑的：google.golang.org/protobuf@v1 不存在，而 v2 存在。

10 年后，希望没人关心这个旧的 github.com/golang/protobuf，那么，这个令人困惑的事就不存在了。

我们可以把新的 API 标记为 v1：

在导入路径中难以清楚地区分开来。

把 google.golang.org/protobuf 的第一个版本标记为 v1 是有意义的。

如果我们认为它是个糟糕的想法，那么，从 v1 转到 v2 比从 v2 回退到 v1 更容易一些。
```

此外，Go protobuf APIv2 将从 1.20 版开始。
Neil 对此做了的解释，这样的目的是避免在错误报告中出现版本重叠产生的歧义，他认为 Go protobuf APIv1 永远不会有 1.20 版。

最后要注意的是，APIv1 不会被 APIv2 淘汰，它会得到无限期维护。事实上，其最新的实现（1.4 版）是在 APIv2 之上实现的。

###

https://juejin.cn/post/6844904111268184078

### 对比

V2 生成的 go 代码含有 proto 文件中的注释，个人非常喜欢，V1 版本 go 代码不包含注释，复杂或者嵌套层次比较高的时候，没有注释， 真是另人抓狂；
在官宣说明中有提到旧版无意义的字段也没有了；
V2 代码中新增了 state，sizeCache，unknownFields 三个字段，具体意义有待研究；
V2 中多引入两个包，分别为 protoreflect，protoimpl ；

### 应用
若现有项目还在用 V1 的 API ，新项目想使用 V2 的 API 的简单的方法

#### 步骤
安装 V1 版本的 protoc-gen-go ，重命名为 protoc-gen-goV1 ;
protoc --goV1_out=:. recharge.proto

安装 V2 版本的 protoc-gen-go ，重命名为 protoc-gen-goV2 ；
protoc --goV2_out=:. recharge.proto

API 版本与 protocol buffer 语言的版本：proto1、proto2、proto3 是不同的，
APIv1 和 APIv2 是 Go 中的具体实现，他们都支持 proto2 和 proto3 语言版本。

github.com/golang/protobuf 模块是 APIv1
google.golang.org/protobuf 模块是 APIv2

在一个程序中，也有可能使用不同的 API 版本，这是至关重要的，所以，我们继续支持使用 APIv1 的程序。

github.com/golang/protobuf@v1.3.4  是 APIv1 最新 pre-APIv2 版本。
github.com/golang/protobuf@v1.4.0  是 APIv2 实现的 APIv1 的一个版本。API 是相同的，但是底层实现得到了新 API 的支持。

该版本包含 APIv1 和 APIv2 之间的转换函数，通过 proto.Message 接口来简化两者之间的转换。

google.golang.org/protobuf@v1.20.0 是 APIv2，该模块取决于 github.com/golang/protobuf@v1.4.0，所以任何使用 APIv2 的程序都将会自动选择一个与之对应的集成 APIv1 的版本。

