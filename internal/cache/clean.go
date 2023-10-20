package cache

import (
	"fmt"
	"github.com/srevinsaju/togomak/v1/internal/meta"
	"github.com/srevinsaju/togomak/v1/internal/x"
	"os"
	"path/filepath"
	"sync"
)

func CleanCache(dir string, recursive bool) {
	dirPath := filepath.Join(dir, meta.BuildDirPrefix, "pipelines", "tmp")
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if path == dirPath {
			return nil
		}
		if info != nil && info.IsDir() {
			fmt.Println("removing", path)
			x.Must(os.RemoveAll(path))
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	var wg sync.WaitGroup
	if recursive {
		entries, err := os.ReadDir(dir)
		if err != nil {
			panic(err)
		}
		for _, entry := range entries {

			if entry.IsDir() {
				wg.Add(1)
				go func(entry os.DirEntry) {
					defer wg.Done()
					CleanCache(filepath.Join(dir, entry.Name()), recursive)
				}(entry)
			}
		}
	}
	wg.Wait()

}
