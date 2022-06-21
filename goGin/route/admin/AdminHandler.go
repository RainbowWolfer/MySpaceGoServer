package admin

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"math/rand"
	"net/http"
	"rainbowwolfer/myspacegoserver/api"
	"rainbowwolfer/myspacegoserver/goGin/ginTools"
	"time"
)

func AdminHandler() (string, ginTools.RouteMap) {
	routeMap := ginTools.NewRouteMap()

	routeMap.AddRoute(ginTools.Route{
		Name:   "/clearunusedpostimages",
		Fun:    ClearUnusedPostImages,
		Method: http.MethodPost,
	})
	routeMap.AddRoute(ginTools.Route{
		Name:   "/reinflatedefaultposts",
		Fun:    ReinflateDefaultPosts,
		Method: http.MethodPost,
	})

	// 可以读取url中的参数
	return "/admin", *routeMap
}

func ClearUnusedPostImages(context *gin.Context) {
	r := context.Request
	w := context.Writer
	if api.CheckRequestMethodReturn(w, r, "get") {
		return
	}
	query := r.URL.Query()
	if api.CheckMissingParamters(w, query, true, "key") {
		return
	}

}

func ReinflateDefaultPosts(context *gin.Context) {
	r := context.Request
	w := context.Writer
	rand.Seed(time.Now().UnixNano())
	if api.CheckRequestMethodReturn(w, r, "get") {
		return
	}
	query := r.URL.Query()
	if api.CheckMissingParamters(w, query, true, "key") {
		return
	}
	key := query["key"][0]
	if key != "eb9f60e5c17ec16a7dfbf79321b79afa" {
		api.HttpError(w, "key error", http.StatusBadRequest)
		return
	}

	database, err := api.GetDatabase()
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer database.Close()

	_, err = database.Exec("DELETE FROM posts;")
	if err != nil {
		api.HttpError(w, "delete error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = database.Exec("DELETE FROM users;")
	if err != nil {
		api.HttpError(w, "delete error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sql := "INSERT INTO users VALUES (1,'myspace','myspace','RainbowWolfer@outlook.com','This is official account for MySpace. Feel free to tell us what improvoments should be made or just come small talking. All are welcomed!');"
	// println(sql)
	_, err = database.Exec(sql)

	if err != nil {
		api.HttpError(w, "insert user error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	usersCount := rand.Intn(20) + 2

	for i := 2; i < usersCount; i++ {
		sql = fmt.Sprintf("INSERT INTO users VALUES (%d,'Test Dummy %d','123456','%d@test.com','Test Dummy #%d');", i, i, i, i)
		// println(sql)
		_, err = database.Exec(sql)

		if err != nil {
			api.HttpError(w, "insert user error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	list := []string{
		"Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry\\'s standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book.",
		" It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum.",
		"It is a long established fact that a reader will be distracted by the readable content of a page when looking at its layout. ",
		"The point of using Lorem Ipsum is that it has a more-or-less normal distribution of letters, as opposed to using \"Content here, content here\", making it look like readable English. ",
		"Many desktop publishing packages and web page editors now use Lorem Ipsum as their default model text, and a search for \"lorem ipsum\" will uncover many web sites still in their infancy. ",
		"Various versions have evolved over the years, sometimes by accident, sometimes on purpose (injected humour and the like).",
		"Contrary to popular belief, Lorem Ipsum is not simply random text. It has roots in a piece of classical Latin literature from 45 BC, making it over 2000 years old. Richard McClintock, a Latin professor at Hampden-Sydney College in Virginia, looked up one of the more obscure Latin words, consectetur, from a Lorem Ipsum passage, and going through the cites of the word in classical literature, discovered the undoubtable source.",
		"Lorem Ipsum comes from sections 1.10.32 and 1.10.33 of \"de Finibus Bonorum et Malorum\" (The Extremes of Good and Evil) by Cicero, written in 45 BC. This book is a treatise on the theory of ethics, very popular during the Renaissance. The first line of Lorem Ipsum, \"Lorem ipsum dolor sit amet..\", comes from a line in section 1.10.32.",
		"The standard chunk of Lorem Ipsum used since the 1500s is reproduced below for those interested.",
		"Sections 1.10.32 and 1.10.33 from \"de Finibus Bonorum et Malorum\" by Cicero are also reproduced in their exact original form, accompanied by English versions from the 1914 translation by H. Rackham.",
		"Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed vitae finibus metus. Etiam ultrices magna vitae ligula sodales suscipit. Sed eu nibh in dolor pharetra varius vitae sit amet dolor.",
		"Morbi auctor pharetra ipsum vitae tempus. Sed risus risus, iaculis eget sapien eu, suscipit vulputate nulla.",
		"Donec vel purus non lacus euismod imperdiet eu tempus sapien. Donec non dui sed odio eleifend sagittis quis commodo enim. Proin nec magna sem.",
	}

	random := rand.Intn(100)
	println(random)
	for i := 1; i < random+100; i++ {
		text := list[rand.Intn(len(list))]
		publisher_id := rand.Intn(usersCount-1) + 1
		sql = fmt.Sprintf("INSERT INTO posts VALUES (%d,%d, TIMESTAMPADD(SECOND,%d,NOW()), TIMESTAMPADD(SECOND,%d,NOW()),0,'%s',FALSE,0,'official,test,LoremIpsum',%d,%d,0,0,'all','all',FALSE,-1,-1);", i, publisher_id, i, i, text, rand.Intn(500), rand.Intn(100))
		// println(sql)
		_, err = database.Exec(sql)

		if err != nil {
			api.HttpError(w, "insert data error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	fmt.Fprintf(w, "Successfully infalte default data (%d) with users (%d)", random+100, usersCount-1)
}
