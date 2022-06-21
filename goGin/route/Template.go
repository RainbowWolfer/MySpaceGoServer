package route

import (
	"GoWeb/domain"
	"GoWeb/goGin/ginTools"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func ReturnTemplateMap() (string, ginTools.RouteMap) {
	routeMap := ginTools.NewRouteMap()
	//获取界面 html
	routeMap.AddRoute(ginTools.Route{
		Name: "/paraHtml",
		Fun: func(context *gin.Context) {
			context.HTML(http.StatusOK, "para/para.html", gin.H{})
		},
		Method: http.MethodGet,
	})
	// index 传递数据
	routeMap.AddRoute(ginTools.Route{
		Name: "/",
		Fun: func(context *gin.Context) {
			context.HTML(http.StatusOK, "index.html", gin.H{
				"user": &domain.User{Name: "temp news数据", Addr: "DDs"},
			})
		},
		Method: http.MethodGet,
	})
	// return Json map
	routeMap.AddRoute(ginTools.Route{
		Name: "/backGoods",
		Fun: func(context *gin.Context) {
			context.HTML(http.StatusOK, "html/back/goods.html", gin.H{
				"user":    &domain.User{Name: "back goods数据", Addr: "DDs"},
				"names":   []string{"a", "b", "c"},
				"nilList": []string{"nua", "bas", "sc", "ds"},
				"age":     []string{"a", "b", "c", "d"},
				"date":    int(time.Now().Unix()), //当前时间的时间戳
			})
		},
		Method: http.MethodGet,
	})

	// return Json
	routeMap.AddRoute(ginTools.Route{
		Name: "/tempNews",
		Fun: func(context *gin.Context) {
			context.HTML(http.StatusOK, "html/temp/news.html", gin.H{
				"user": &domain.User{Name: "temp news数据", Addr: "/temp/news"},
			})
		},
		Method: http.MethodGet,
	})

	// return struct
	routeMap.AddRoute(ginTools.Route{
		Name: "/tempHead",
		Fun: func(context *gin.Context) {
			context.HTML(http.StatusOK, "html/temp/head.html", gin.H{
				"us":   &domain.User{Name: "temp news数据", Addr: "阿斯爱上的顿"},
				"mode": "爱上是否",
				"date": 1334086974,
			})
		},
		Method: http.MethodGet,
	})

	return "/temp", *routeMap
}
