package main

import (
	"GoWeb/domain"
	"GoWeb/goGin/ginTools"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetParaMap2() (string, ginTools.RouteMap) {
	routeMap := ginTools.NewRouteMap()

	//Querystring指的是URL中?后面携带的参数。下面介绍获取querystring参数的几种方式

	//fromUrl?name = asdsd & psw=asd
	routeMap.AddRoute(ginTools.Route{
		Name: "/fromUrl",
		//
		Fun: func(context *gin.Context) {
			context.JSON(http.StatusOK, gin.H{
				//获取指定参数，并返回Json数据
				"name": context.Query("name"),
				"psw":  context.DefaultQuery("psw", "mima"),
			})
		},
		Method: http.MethodGet,
	})

	// ShouldBind 自动绑定
	routeMap.AddRoute(ginTools.Route{
		Name: "/getUser",
		Fun: func(context *gin.Context) {
			u := &domain.User{}

			/*
				ShouldBind能够基于请求的不同，自动提取JSON、form表单和QueryString类型的数据，并把值绑定到指定的结构体对象。
			*/
			if err := context.ShouldBind(u); err != nil {
				context.JSON(http.StatusOK, gin.H{
					"err": err.Error(),
				})
			} else {
				context.JSON(http.StatusOK, u)
				fmt.Printf("%#v", u)

			}
		},
		Method: http.MethodGet,
	})

	// 表单提交
	routeMap.AddRoute(ginTools.Route{
		Name: "/fromForm1",
		Fun: func(context *gin.Context) {
			context.JSON(200, gin.H{
				"name": context.PostForm("name"),
				"pwd":  context.PostForm("pwd"),
				"age":  context.DefaultPostForm("age", "0"),
			})

		},
		Method: http.MethodGet,
	})

	// 可以读取url中的参数
	routeMap.AddRoute(ginTools.Route{
		Name: "/fromForm2",
		Fun: func(context *gin.Context) {
			u := &domain.User{}
			if err := context.ShouldBind(&u); err != nil {
				context.JSON(http.StatusOK, gin.H{
					"err": err.Error(),
				})
			} else {
				context.JSON(http.StatusOK, u)
				fmt.Println("转换成功")
				fmt.Printf("%#v", u)

			}
		},
		Method: http.MethodGet,
	})
	return "/getPara", *routeMap
}

func test1() {
	route := ginTools.Route{
		Name: "/getPara",
		Fun: func(context *gin.Context) {
			h := gin.H{
				"name": "getPara",
			}
			context.JSON(http.StatusOK, h)
		},
		Method: http.MethodGet,
	}
	routeMap := ginTools.NewRouteMap()
	routeMap.AddRoute(route)
	engine := ginTools.NewDefaultEngine(8080)
	engine.AddRouteMap(routeMap)
	//engine.AddGroupRouteMap("/routeMap", *routeMap).RunServer()
	engine.AddGroupRouteMap(GetParaMap2()).RunServer()
}
func main() {
	test1()
}
