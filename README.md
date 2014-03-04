# Tiny Go Web

Tiny Go Web (TGW)是一个非常简单的Web框架，甚至谈不上框架。TGW无意取代任何框架，TGW的诞生是因为作者在使用beego时有种挫败感，决定自己重新写一个适合自己网站用的，从构思到完成总共
只花了一天的时间，因为觉得它已经够用了，就没有继续添加新的功能。

TGW使用非常简单，没有固定的目录结构，不过遵循大众习惯，我把它组成以下结构：

│── controllers
│   ├── default.go
├── main.go
├── models
│   └── book.go
├── static
│   ├── css
│   ├── img
│   └── js
└──── view
    ├── include
    │   └── nav.html
    └── index.html



