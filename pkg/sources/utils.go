package sources

import (
	"github.com/srevinsaju/togomak/pkg/context"
	"github.com/srevinsaju/togomak/pkg/meta"
	"github.com/srevinsaju/togomak/pkg/schema"
	"net/url"
	"path/filepath"
)

func GetStorePath(ctx *context.Context, v schema.StageConfig) string {

	u, err := url.Parse(v.Source.URL)
	if err != nil {
		ctx.Logger.Fatal("Failed to parse extends parameter", err)
	}
	var dest string
	switch v.Source.Type {
	case TypeGit:
		dest = filepath.Join(ctx.Data.GetString("cwd"), meta.BuildDirPrefix, meta.BuildDir, meta.ExtendsDir, TypeGit, u.Host, u.Path)
	case TypeFile:
		dest = filepath.Join(ctx.Data.GetString(context.KeyCwd), meta.BuildDirPrefix, meta.BuildDir, meta.ExtendsDir, TypeFile, u.Path)
	}
	return dest

}
