package ginTools

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// WalkFiles 通过遍历加载文件
func WalkFiles(root string) []string {
	var files []string
	filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if strings.HasSuffix(path, ".html") {
			files = append(files, path)
		}

		return nil
	})
	return files
}
