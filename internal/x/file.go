package x

import "os"

func FileExists(path string) bool {
	f, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	} else {
		return !f.IsDir()
	}
}
