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
