package main

import (
	"rainbowwolfer/myspacegoserver/goGin/ginTools"
	"rainbowwolfer/myspacegoserver/goGin/route"
	"rainbowwolfer/myspacegoserver/goGin/route/admin"
	"rainbowwolfer/myspacegoserver/goGin/route/collections"
	"rainbowwolfer/myspacegoserver/goGin/route/post"
	"rainbowwolfer/myspacegoserver/goGin/route/user"
	"rainbowwolfer/myspacegoserver/goGin/route/validation"
)

func tests() {

	engine := ginTools.NewDefaultEngine(8080)
	engine.SetFuncMap(route.AllMethodMap())
	//tempPath := "goGin/templates/**/**/*"
	//engine.LoadHTMLGlob(tempPath)
	engine.Engine.LoadHTMLFiles(ginTools.WalkFiles("goGin/templates/")...)

	staticPath := "goGin/static"
	engine.Static("/static", staticPath)

	path, routeMap := route.ReturnJsonMap()

	engine.AddRouteMap(route.IndexHandler())
	engine.AddRouteMap(route.LoginHandler())

	engine.AddGroupRouteMap(post.PostHandler())
	engine.AddGroupRouteMap(path, *routeMap)
	engine.AddGroupRouteMap(route.UploadHandler())

	engine.AddGroupRouteMap(validation.ValidationHandler())
	engine.AddGroupRouteMap(collections.CollectionsHandler())
	engine.AddGroupRouteMap(admin.AdminHandler())

	engine.AddGroupRouteMap(user.UserHandler()).RunServer()
}
func main() {

	tests()
	//test2()
}
