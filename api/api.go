package api

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"
)

func ToJson(object any) string {
	j, err := json.MarshalIndent(object, "", "\t")
	if err != nil {
		return "Error Converting Object to Json"
	} else {
		return string(j)
	}
}

func DeleteEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func Now() string {
	dt := time.Now()
	//Format MM-DD-YYYY hh:mm:ss
	return dt.Format("2006-01-02 15:04:05")
}

func IsEmpty(str *string) bool {
	if str == nil {
		return true
	}
	if len(*str) == 0 {
		return true
	}
	if len(strings.TrimSpace(*str)) == 0 {
		return true
	}

	return false
}
