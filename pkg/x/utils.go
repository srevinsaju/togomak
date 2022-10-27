package x

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func MustReturn(v any, err error) any {
	if err != nil {
		panic(err)
	}
	return v
}

func Contains[V comparable](slice []V, item V) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
