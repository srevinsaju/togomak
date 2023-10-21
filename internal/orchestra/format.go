package orchestra

import (
	"bytes"
	"fmt"
	"github.com/bmatcuk/doublestar"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/srevinsaju/togomak/v1/internal/ci"
	"github.com/srevinsaju/togomak/v1/internal/parse"
	"os"
	"path/filepath"
)

func Format(cfg ci.ConductorConfig, check bool, recursive bool) error {
	conductor := ci.NewConductor(cfg)

	var toFormat []string

	if recursive {
		matches, err := doublestar.Glob("**/*.hcl")
		for _, path := range matches {
			conductor.Logger().Tracef("Found %s", path)
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			outSrc := hclwrite.Format(data)
			if !bytes.Equal(outSrc, data) {
				conductor.Logger().Tracef("%s needs formatting", path)
				toFormat = append(toFormat, path)
			}
		}
		if err != nil {
			conductor.Logger().Fatalf("Error while globbing for **/*.hcl: %s", err)
		}
	} else {
		fDir := parse.ConfigFileDir(conductor.Config.Paths)
		fNames, err := os.ReadDir(fDir)
		if err != nil {
			panic(err)
		}

		for _, f := range fNames {
			if f.IsDir() {
				continue
			}
			if filepath.Ext(f.Name()) != ".hcl" {
				continue
			}
			fn := filepath.Join(fDir, f.Name())
			data, err := os.ReadFile(fn)
			if err != nil {
				return err
			}
			outSrc := hclwrite.Format(data)
			if !bytes.Equal(outSrc, data) {
				conductor.Logger().Tracef("%s needs formatting", fn)
				toFormat = append(toFormat, fn)
			}
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
