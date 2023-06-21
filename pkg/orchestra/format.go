package orchestra

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/srevinsaju/togomak/v1/pkg/meta"
	"github.com/srevinsaju/togomak/v1/pkg/pipeline"
	"os"
	"path/filepath"
)

func Format(cfg Config, check bool, recursive bool) error {
	t, ctx := NewContextWithTogomak(cfg)

	var toFormat []string

	if recursive {
		err := filepath.WalkDir(t.cwd, func(path string, d os.DirEntry, err error) error {
			if filepath.Base(path) == meta.ConfigFileName {
				t.logger.Tracef("Found %s", path)
				data, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				outSrc := hclwrite.Format(data)
				if !bytes.Equal(outSrc, data) {
					t.logger.Tracef("%s needs formatting", path)
					toFormat = append(toFormat, path)
				}
			} else {
				t.logger.Tracef("Skipping %s", path)
			}
			return nil
		})
		if err != nil {
			t.logger.Fatalf("Error while formatting: %s", err)
		}
	} else {
		fn := pipeline.ConfigFilePath(ctx)
		data, err := os.ReadFile(fn)
		if err != nil {
			return err
		}
		outSrc := hclwrite.Format(data)
		if !bytes.Equal(outSrc, data) {
			t.logger.Tracef("%s needs formatting", fn)
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
	return nil

}
