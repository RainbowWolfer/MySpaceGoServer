package route

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"rainbowwolfer/myspacegoserver/api"
	"rainbowwolfer/myspacegoserver/goGin/config"
	"rainbowwolfer/myspacegoserver/goGin/ginTools"
)

func UploadHandler() (string, ginTools.RouteMap) {
	routeMap := ginTools.NewRouteMap()

	routeMap.AddRoute(ginTools.Route{
		Name:   "/avatar",
		Fun:    uploadAvatar_post,
		Method: http.MethodPost,
	})
	// 可以读取url中的参数
	return "/upload", *routeMap
}
func uploadAvatar_post(context *gin.Context) {
	r := context.Request
	w := context.Writer
	if api.CheckRequestMethodReturn(w, r, "post") {
		return
	}

	query := r.URL.Query()
	if api.CheckMissingParamters(w, query, true, "id") {
		return
	}
	id := query["id"][0]

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		api.HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// get a reference to the fileHeaders
	files := r.MultipartForm.File["file"]

	fileLen := len(files)
	if fileLen != 1 {
		api.HttpError(w, fmt.Sprintf("Can only apply 1 file. Currently received %d file(s)", fileLen), http.StatusBadRequest)
		return
	}

	fileHeader := files[0]

	if fileHeader.Size > config.MAX_UPLOAD_SIZE {
		api.HttpError(w, fmt.Sprintf("The uploaded image is too big: %s. Please use an image less than 1MB in size", fileHeader.Filename), http.StatusBadRequest)
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer file.Close()

	buff := make([]byte, 512)
	_, err = file.Read(buff)
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filetype := http.DetectContentType(buff)
	if filetype != "image/jpeg" && filetype != "image/jpg" && filetype != "image/png" {
		api.HttpError(w, "The provided file format is not allowed. Please upload a JPEG(JPG) or PNG image", http.StatusBadRequest)
		return
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	database, err := api.GetDatabase()
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer database.Close()

	sql := fmt.Sprintf("select u_id from users where u_id = '%s'", id)
	rows, err := database.Query(sql)
	if err != nil {
		api.HttpError(w, "Query Error with :"+sql+"\n"+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	foundUser := rows.Next()

	if !foundUser {
		api.HttpError(w, "User Not Found", http.StatusBadRequest)
		return
	}

	err = os.MkdirAll("./uploads/avatars", os.ModePerm)
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	f, err := os.Create(fmt.Sprintf("./uploads/avatars/%s%s", "user_"+id, filepath.Ext(fileHeader.Filename)))
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	defer f.Close()

	pr := &api.Progress{
		TotalSize: fileHeader.Size,
	}

	_, err = io.Copy(f, io.TeeReader(file, pr))
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "Upload successful")
}
