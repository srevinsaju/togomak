package orchestra

import (
	"bytes"
	"fmt"
	"github.com/bmatcuk/doublestar"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/srevinsaju/togomak/v1/pkg/pipeline"
	"os"
)

func Format(cfg Config, check bool, recursive bool) error {
	t, ctx := NewContextWithTogomak(cfg)

	var toFormat []string

	if recursive {
		matches, err := doublestar.Glob("**/*.hcl")
		for _, path := range matches {
			t.Logger.Tracef("Found %s", path)
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			outSrc := hclwrite.Format(data)
			if !bytes.Equal(outSrc, data) {
				t.Logger.Tracef("%s needs formatting", path)
				toFormat = append(toFormat, path)
			}
		}
		if err != nil {
			t.Logger.Fatalf("Error while globbing for **/*.hcl: %s", err)
		}
	} else {
		fn := pipeline.ConfigFilePath(ctx)
		data, err := os.ReadFile(fn)
		if err != nil {
			return err
		}
		outSrc := hclwrite.Format(data)
		if !bytes.Equal(outSrc, data) {
			t.Logger.Tracef("%s needs formatting", fn)
			toFormat = append(toFormat, fn)
		}
	}
	for _, fn := range toFormat {
		fmt.Println(fn)
		if !check {
			data, err := os.ReadFile(fn)
			if err != nil {
				panic(err)
			}
			outSrc := hclwrite.Format(data)
			err = os.WriteFile(fn, outSrc, 0644)
			if err != nil {
				panic(err)
			}
		}
	}
	if check && len(toFormat) > 0 {
		os.Exit(1)
	}
	return nil

}
