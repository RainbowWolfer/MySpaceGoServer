package route

import (
	"html/template"
	"time"
)

func AllMethodMap() template.FuncMap {
	t := template.FuncMap{}
	t["UnixToTime"] = UnixToTime
	t["StringJoin"] = StringJoin
	return t
}

func UnixToTime(timeStamp int) string {
	unix := time.Unix(int64(timeStamp), 0)
	return unix.Format("2021-01-02 15:04:05")
}
func StringJoin(str, str2 string) string {
	return str + "------" + str2
}
