package main

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/mitchellh/mapstructure"
	"github.com/srevinsaju/togomak/pkg/context"
	"github.com/srevinsaju/togomak/pkg/schema"
	"os"
	"strings"
)

func (g *StageGit) Name() string {
	return "git"
}

func (g *StageGit) Description() string {
	return "A git data provider"
}

func (g *StageGit) CanRun() bool {
	return true
}

func (g *StageGit) Run() error {

	return nil
}

func (g *StageGit) GatherInfo() error {
	err := g.init()
	if err != nil {
		return err
	}
	//g.context.Mutex.Lock()
	//defer g.context.Mutex.Unlock()
	if g.error != nil {
		return g.error
	}
	ref, err := g.g.Head()
	if err != nil {
		return err
	}
	sha := strings.Split(ref.String(), " ")[0]
	shortSha := sha[:7]
	g.context.Data["branch"] = ref.Name().String()
	w, err := g.g.Worktree()
	if err != nil {
		return err
	}
	s, err := w.Status()
	if err != nil {
		return err
	}
	if !s.IsClean() {
		sha += "-dirty"
		shortSha += "-dirty"
	}
	g.context.Data["sha"] = sha
	g.context.Data["short_sha"] = shortSha
	return nil
}

func (g *StageGit) SetContext(c schema.Context) error {
	g.context.Data = c.Data
	if g.context.Data == nil {
		g.context.Data = context.Data{}
		return nil
	}
	g.customUserConfig = true
	// some data might have got updated
	var gitCfg GitConfig
	err := mapstructure.Decode(g.context.Data, &gitCfg)
	if err != nil {
		return err
	}
	g.logger.Info("decoded data", g.context.Data) //gitCfg) //,
	g.gitConfig = gitCfg

	return nil
}

func (g *StageGit) init() error {

	var repo *git.Repository
	var err error
	if g.customUserConfig {
		g.logger.Info("decoded data", g.gitConfig) //gitCfg) //,
		g.logger.Info("Cloning repository from user config")
		g.logger.Info("Repository URL: " + g.gitConfig.Repository.URL)
		repo, err = git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
			URL:               g.gitConfig.Repository.URL,
			InsecureSkipTLS:   g.gitConfig.SkipTLSInsecure,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
			Depth:             g.gitConfig.Repository.Depth,
			SingleBranch:      true,
		})
		if err != nil {
			g.logger.Error("Error cloning repository: " + err.Error())
		}

	} else {
		g.logger.Info("Using existing current directory repo as source")
		var wd string
		wd, err = os.Getwd()
		if err != nil {
			g.logger.Warn("error getting working directory", "error", err)
		}

		repo, err = git.PlainOpenWithOptions(wd, &git.PlainOpenOptions{
			DetectDotGit:          true,
			EnableDotGitCommonDir: true,
		})
		if err != nil {
			g.logger.Error("Error opening existing repository: " + err.Error())
		}

	}

	g.g = repo
	if err != nil {
		return err
	}
	return nil

}

func (g *StageGit) GetContext() schema.Context {
	return g.context
}
