package x

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func Contains[V comparable](slice []V, item V) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
