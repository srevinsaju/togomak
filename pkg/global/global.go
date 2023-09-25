package global

var tempDir string

func SetTempDir(dir string) {
	tempDir = dir
}

func TempDir() string {
	if tempDir == "" {
		panic("temp dir not set")
	}
	return tempDir
}
