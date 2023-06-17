package x

import (
	"fmt"
	"testing"
)

func TestMust(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	Must(nil)
	Must(fmt.Errorf("test"))
}
