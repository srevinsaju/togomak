package cache

import (
	"fmt"
	"github.com/srevinsaju/togomak/v1/pkg/meta"
	"os"
	"path/filepath"
)

func CleanCache(dir string) {
	dirPath := filepath.Join(dir, meta.BuildDirPrefix, "pipelines", "tmp")
	filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if path == dirPath {
			return nil
		}
		if info != nil && info.IsDir() {
			fmt.Println("removing", path)
			os.RemoveAll(path)
		}
		return nil
	})

}
