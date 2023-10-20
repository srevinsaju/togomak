package x

import (
	"strings"
)

func RenderBlock(block ...string) string {
	return strings.Join(block, ".")
}
