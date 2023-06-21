package orchestra

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/srevinsaju/togomak/v1/pkg/pipeline"
	"os"
)

func Format(cfg Config, check bool) error {
	_, ctx := NewContextWithTogomak(cfg)
	fn := pipeline.ConfigFilePath(ctx)

	var err error
	var hasLocalChanges bool = false

	data, err := os.ReadFile(fn)
	if err != nil {
		panic(err)
	}
	outSrc := hclwrite.Format(data)

	if !bytes.Equal(outSrc, data) {
		hasLocalChanges = true
	}
	if check && hasLocalChanges {
		fmt.Println(fn)
		os.Exit(1)
	}

	if hasLocalChanges {
		fmt.Println(fn)
		return os.WriteFile(fn, outSrc, 0644)
	}
	return nil

}
