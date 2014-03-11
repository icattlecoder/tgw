# Tiny Go Web
---

Tiny Go Web (TGW)是一个非常简单的Web框架，甚至谈不上框架。TGW无意取代任何框架，TGW的诞生是因为作者在使用beego时有种挫败感，决定自己重新写一个适合自己网站用的，从构思到完成总共只花了一天的时间，因为觉得它已经够用了，就没有继续添加新的功能。

## 运行示例

```
> go get github.com/icattlecoder/tgw
> cd src/github.com/icattlecoder/tgw/example
> go build
> ./example 
```

![img](http://icattlecoder.qiniudn.com/tgw.png)


TGW使用非常简单，没有固定的目录结构，不过遵循大众习惯，建议组织以下结构：

```
│── controllers
│   ├── default.go
├── main.go
├── models
│   └── Author.go
├── static
│   ├── css
│   ├── img
│   └── js
└──── view
    ├── include
    │   └── nav.html
    └── index.html
```

## 控制器

控制器实现自动路由注册，例如有以下的结构

```go

type Server struct {
	//成员由业务逻辑而定，如mgo的数据库连接信息等
}

func NewServer( /*入参，例如从配置文件中读取*/) *Server {
	return &Server{}
}

//对应模板为index.html ,返回值data用于渲染模板
func (s *Server) Index() (data map[string]interface{}) {
	data = map[string]interface{}{}
	author := Author{
		Name:  "icattlecoder",
		Email: []string{"icattlecoder@gmail.com", "iwangming@hotmail.com"},
		QQ:    "405283013",
		Blog:  "http://blog.segmentfault.com/icattlecoder",
	}
	data["author"] = author
	return
}

//由于没有json.html模板，但是却有data返回值，此data将以json字符串的格式返回
func (s *Server) Json() (data map[string]interface{}) {
	data = map[string]interface{}{}
	author := Author{
		Name:  "icattlecoder",
		Email: []string{"icattlecoder@gmail.com", "iwangming@hotmail.com"},
		QQ:    "405283013",
		Blog:  "http://blog.segmentfault.com/icattlecoder",
	}
	data["author"] = author
	return
}


//这里根据请求自动解析出args
//例如可将 /hello?msg=hello world的函数解析为TestArgs{Msg:"hello world"}
//由于没有hello.html模板，并且没有返回值，可以通过env中的RW成员写入返回数据
func (s *Server) Hello(args TestArgs, env tgw.ReqEnv) {

	env.RW.Write([]byte(args.Msg))
	err = env.Session.Set("key", args)
	if err != nil {
		log.Println(err)
	}
}
```

以下是程序启动代码
``` go
func main() {
	ser := controllers.NewServer()
	t:=tgw.NewTGW()
	log.Fatal(t.Register(&ser).Run(":8080"))
}
```

tgw的Register方法会自动注册以下的路由：

```
/hello 		 ===> Hello
/index 		 ===> Index
/Json 		 ===> Json
/admin/index ===> AdminIndex
```

即`localhost:8080/index`的处理函数是`service.Index`,`localhost:8080/admin/index`的处理函数是`service.AdminIndex`

## 视图

视图默认放在view文件夹中，其文件名与url有关，例如：`/hello`对应 `view/index.html`
如果某个url没有对应的视图，但是它的处理函数却有返回值，那么将返回对象JOSN序列化的结果。
视图中可以通过`<include src="<src>" />`指令包含其它文件，如公共区域

## 参数解析

```
type TestArgs struct {
	Msg string
}
func (s *Server) Hello(args TestArgs, env tgw.ReqEnv)

```
对于请求`localhost:8080/Hello?msg=Hello world`，tgw将自动根据语法方法识别并解析出TestArgs变量.
当前能够自动解析的类型有`int`、`string`、`bool`、`float64`

## Session支持

框架实现了一个简单的基于内存的session管理,如果需要使用session，处理函数必须有一个类型为tgw.ReqEnv的参数,通过此函数可访问Session。

## 自定义

如果函数包含类型为tgw.ReqEnv函数，且无返回值，可以直接向ReqEnv.RW中写入返回结果
