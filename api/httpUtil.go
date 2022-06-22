package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Output struct {
	Code    int
	Content string
	Error   int
}

func HttpError(w http.ResponseWriter, content string, code int) {
	c := fmt.Sprintf("%s (%d)", content, code)
	json := ToJson(Output{
		Code:    code,
		Content: c,
		Error:   0,
	})
	println(json)
	http.Error(w, json, code)
}

func HttpErrorWithCode(w http.ResponseWriter, content string, code int, errorCode int) {
	c := fmt.Sprintf("%s (%d)", content, code)
	json := ToJson(Output{
		Code:    code,
		Content: c,
		Error:   errorCode,
	})
	println(json)
	http.Error(w, json, code)
}

func CheckRequestMethodReturn(w http.ResponseWriter, r *http.Request, method string) bool {
	println(r.RequestURI)
	if !strings.EqualFold(r.Method, method) {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return true
	} else {
		return false
	}
}

func CheckRequestMethod(r *http.Request, method string) error {
	println(r.RequestURI)
	if strings.EqualFold(r.Method, method) {
		return nil
	} else {
		return errors.New("method not allowed")
	}
}

func CheckMissingParamters(w http.ResponseWriter, query url.Values, and bool, paras ...string) bool {
	var missings []string = []string{}
	for _, p := range paras {
		if !query.Has(p) {
			missings = append(missings, p)
			continue
		}
		content := query[p][0]
		if len(strings.TrimSpace(content)) == 0 {
			missings = append(missings, p)
		}
	}
	if and {
		if len(missings) == 0 {
			return false
		} else {
			var errorLines []string
			for _, m := range missings {
				errorLines = append(errorLines, "'"+m+"'")
			}
			HttpError(w, "Missing Parameter "+strings.Join(errorLines, " And "), http.StatusBadRequest)
			return true
		}
	} else {
		if len(missings) == len(paras) {
			var errorLines []string
			for _, m := range missings {
				errorLines = append(errorLines, "'"+m+"'")
			}
			HttpError(w, "Missing Parameter "+strings.Join(errorLines, " Or "), http.StatusBadRequest)
			return true
		} else {
			return false
		}
	}
}
