package route

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"rainbowwolfer/myspacegoserver/goGin/ginTools"
)

func GetParaMap() (string, ginTools.RouteMap) {
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
	return "/getPara", *routeMap
}
