package ginTools

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

type Route struct {
	Name   string
	Fun    gin.HandlerFunc
	Method string
}

// RouteMap 使用map 解决路径重复问题
type RouteMap struct {
	route1Map map[string]Route
}

// AddRouteList  多个添加
func (receiver *RouteMap) AddRouteList(routes []Route) {
	for _, route := range routes {
		receiver.AddRoute(route)
	}
}

// AddRoute  添加
func (receiver *RouteMap) AddRoute(routes Route) {
	maps := receiver.route1Map
	maps[routes.Name] = routes
}

// DelRoute  删除
func (receiver *RouteMap) DelRoute(routes Route) {
	maps := receiver.route1Map
	_, ok := maps[routes.Name]
	if ok {
		fmt.Printf("移除元素:%v :%v", routes.Name, ok)
		delete(maps, routes.Name)
	}
}

func NewRouteMap() *RouteMap {
	return &RouteMap{
		route1Map: make(map[string]Route),
	}
}
