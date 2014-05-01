# Tiny Go Web
---

Tiny Go Web (TGW)是一个非常简单的Web框架，甚至谈不上框架。TGW无意取代任何框架，TGW的诞生是因为作者在使用beego时有种挫败感，决定自己重新写一个适合自己网站用的([私人借书网](http://www.4jieshu.com)，因为网站没有完成备案，暂时由托管在us的vps进行反射代理到ucloud主机，访问可能会有一定的延时)，从构思到完成总共只花了一天的时间，因为觉得它已经够用了，就没有继续添加新的功能。

## Qiuck Start

```
> go get github.com/icattlecoder/tgw
> cd src/github.com/icattlecoder/tgw/example
> go build
> ./example 
```

![img](http://icattlecoder.qiniudn.com/tgw.png)


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

func (s *Server) AdminIndex(){}
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
/hello 		 ===> func (s *Server) Hello(args TestArgs, env tgw.ReqEnv)
/index 		 ===> func (s *Server) Index() (data map[string]interface{}) 
/Json 		 ===> func (s *Server) Json() (data map[string]interface{})
/admin/index ===> func (s *Server) AdminIndex()
```

即`localhost:8080/index`的处理函数是`service.Index`,`localhost:8080/admin/index`的处理函数是`service.AdminIndex`

## 视图

视图默认放在view文件夹中，其文件名与url有关，例如：`/index`对应 `view/index.html`
如果某个url没有对应的视图，但是它的处理函数却有返回值，那么将返回对象JOSN序列化的结果。
视图中可以通过`<include src="<src>" />`指令包含其它文件，如公共区域

## 参数解析

以下面的代码为例：

```
type TestArgs struct {
	Msg string
}
func (s *Server) Hello(args TestArgs, env tgw.ReqEnv)

```
对于请求`localhost:8080/Hello?msg=Hello world`，tgw将自动根据请求方法(POST或GET)识别并解析出TestArgs变量.
当前能够自动解析的类型有`int`、`string`、`bool`、`float64`

### 扩展参数解析
tgw自带`*Args`参数解析，即结构名符合`*Args`的参数都可以自动解析。如果需要定制解析，实现Parser接口即可

## Session支持

框架实现了一个简单的session管理,基本满足一般的需求，如果需要使用session，处理函数必须有一个类型为tgw.ReqEnv的参数,通过此函数可访问Session。另外，Session的值由memcached存储，因此实际运行时需要一个memecached服务。


## 自定义

如果函数包含类型为tgw.ReqEnv参数，且无返回值，可以直接向ReqEnv.RW中写入返回结果
