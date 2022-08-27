package main

import (
	"github.com/srevinsaju/togomak/pkg/schema"
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

func (g *StageGit) SetContext(context schema.Context) {

	g.context.Data = context.Data
}

func (g *StageGit) GetContext() schema.Context {
	return g.context
}
