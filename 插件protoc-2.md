


### 插件命名规则 

protobuf 插件名称需要使用 protoc-gen-xxx 

当使用 protoc --xxx_out 时就会调用 proto-gen-xxx 插件


### 解析流程

#### 方法1:

流程：
0. 先通过标准输入生成 CodeGeneratorRequest 
0. 通过 CodeGeneratorRequest 初始化插件 plugin
0. 通过插件 plugin 获取文件 File
0. 遍历文件，处理内容，生成文件
0. 向标准输出写入响应 plugin.Response

示例代码：

```go

	s := flag.String("a", "", "a")
	flag.Parse()
	
	// 生成 Request 
	input, _ := ioutil.ReadAll(os.Stdin)
	var req pluginpb.CodeGeneratorRequest
	proto.Unmarshal(input, &req)

	// 设置参数, 生成 plugin
	opts := protogen.Options{
		ParamFunc: flag.CommandLine.Set,
	}
	plugin, err := opts.New(&req)
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(os.Stderr, "a=%s\n", *s)

	// protoc 将一组文件结构传递给程序处理,包含 proto 中import 中的文件
	for _, file := range plugin.Files {
		if !file.Generate { //显示传入的文件为 true
			continue
		}
    
		fmt.Fprintf(os.Stderr, "path:%s\n", file.GoImportPath)
		genF := plugin.NewGeneratedFile(fmt.Sprintf("%s_error.pb.go", 						file.GeneratedFilenamePrefix), file.GoImportPath) //用来处理生成文件的对象
		GenFile(genF, file, *s)
	}

	// 生成 response
	resp := plugin.Response()
	out, err := proto.Marshal(resp)
	if err != nil {
		panic(err)
	}

	// 相应输出到 stdout, 它将被 protoc 接收
	fmt.Fprintf(os.Stdout, string(out))
```
#### 方法2:

```go
var s = flag.String("aaa", "", "aaa")
var s1 = flag.String("bbb", "", "aaa")
flag.Parse()
protogen.Options{
	ParamFunc: flag.CommandLine.Set, //设置命令行参数 --xxx_out=aaa=10,bbb=20:. 
}.Run(
	func(gen *protogen.Plugin) error { //Run内部封装了方法1的步骤
        for _, f := range gen.Files {
            if !f.Generate { //判断是否需要生成代码, 影响这里是否是true的原因是protoc是否指定这个文件
				continue
		    }
            // 遍历各种 message/service... , 同方法1
        }
    }
)
```
示例代码:
```go
protogen.Options{
	ParamFunc: flag.CommandLine.Set,
}.Run(
    // Run 中已经封装好 request 和 response 
	func(gen *protogen.Plugin) error { 
		fmt.Fprintf(os.Stderr, "aaa=%s\n", *s)
		fmt.Fprintf(os.Stderr, "bbb=%s\n", *s1)
		for _, f := range gen.Files {
			fmt.Fprintf(os.Stderr, "f.Generate:%s==>%v\n", f.Desc.Name(), f.Generate)
            // 判断是否需要生成代码
			if !f.Generate { 
				continue
			}
			for _ ,file := range gen.Files {
				for _, serv := range file.Services {
					fmt.Fprintf(os.Stderr, "%s ==> %s\n", serv.Desc.Name(), serv.GoName)
				}
			}
		}
		return nil
	}
)
```
注意:
不能打印在标准输出，因为程序响应返回要占用标准输出传递到父进程 protoc , 要想调试打印到标准错误上。
参数传递方式为 --xxx_out=aaa=10,bbb=20:.
通过 GeneratedFile 写入的, 必须符合 go 语法，否则会报错


### 一些细节



```

// 导入包
// gen *protogen.Plugin
// protogen.GoImportPath("errors") // 把 errors 转换成 GoImportPath 类型
// protogen.GoImportPath("errors").Ident("") // 返回里面变量名或函数名等, 为空时返回包名

// 例如:
//  errors = protogen.GoImportPath("errors")
//  errors.Ident("A")           // errors 包中 A 的标识: errors.A
//  errors.Ident("")            // errors 包名:  errors
//  errors.Ident("New(\"a\")")  // 生成 errors.New("a") 

// genF.QualifiedGoIdent(b) 虽然有导入包和返回合法名称的功能, 如果 GoIdent 直接传入 P 函数会在 P 内部调用 QualifiedGoIdent , 也就是说会自动导包, 并且用合法的名称


// 代表 a/b 包下的 b 包
b := protogen.GoIdent{ 
	GoName:       "b",
	GoImportPath: "a/b",
}

// 自动导入 a/b 包, 并返回名称 b.b
name := genF.QualifiedGoIdent(b) 
fmt.Fprintf(os.Stderr, "b:%s\n", name)

// 代表 c/b 包下的 b 包 
c := protogen.GoIdent{
	GoName:       "b",
	GoImportPath: "c/b",
}

// 自动导入 c/b 包, 并返回名称, 这个包和上面名称都是 b , 这个会返回 b1.b 
name = genF.QualifiedGoIdent(c)
fmt.Fprintf(os.Stderr, "c:%s\n", name)

// 所以说如果使用的是 P 函数生成代码,一般不会手动调用 QualifiedGoIdent , 
// 如果使用的是text模板生成代码, 就需要自己手动调用 QualifiedGoIdent 来生成字符串名称了
```






















