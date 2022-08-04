package main

import "github.com/srevinsaju/buildsys/pkg/schema"


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
	g.context.Data["sha"] = "a34cef"
	return nil
}


func (g *StageGit) SetContext(context schema.Context) {
	
	
	g.context.Data = context.Data
}

func (g *StageGit) GetContext() schema.Context {
	return g.context
}