package bootstrap

import (
	"encoding/json"
	"fmt"
	"github.com/chartmuseum/storage"

	uuid "github.com/satori/go.uuid"
	"github.com/srevinsaju/togomak/pkg/context"
	"github.com/srevinsaju/togomak/pkg/schema"
	"github.com/srevinsaju/togomak/pkg/state"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"time"
)

func LoadStateBackend(ctx *context.Context, state string) storage.Backend {
	if state == "" {
		// use default state manager then
		return ctx.RootParent().Data["default_state_manager"].(storage.Backend)
	}
	u, err := url.Parse(state)
	if err != nil {
		ctx.Logger.Fatal(err)
	}
	if u.Scheme == "" {
		u.Scheme = "file"
	}
	if u.Scheme == "file" {
		ctx.Logger.Tracef("Using file state")
		return storage.NewLocalFilesystemBackend(u.Host + u.Path)
	} else {
		ctx.Logger.Fatalf("Unsupported state backend %s", u.Scheme)
	}
	return nil
}

func UpdateStateForStage(ctx *context.Context, stage schema.StageConfig, stateManager storage.Backend) state.State {
	statePath := filepath.Join("state", ctx.RootParent().Data.GetString(state.WorkspaceDataKey), stage.Id, "state.json")
	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "localhost"
	}

	st := state.State{
		LastModified:              time.Now(),
		Version:                   state.Version,
		LastUsernameHumanReadable: fmt.Sprintf("%s@%s", u.Username, hostname),
		StageId:                   stage.Id,
		LastUsernameUUID:          uuid.NewV4(),
	}
	d, err := json.Marshal(st)
	if err != nil {
		panic(err)
	}
	err = stateManager.PutObject(statePath, d)
	if err != nil {
		ctx.Logger.Fatal(err)
	}
	return st
}

func GetStateForStage(ctx *context.Context, stage schema.StageConfig) (state.State, storage.Backend) {

	ctx = ctx.RootParent().AddChild("state", stage.Id)
	stateBackend := LoadStateBackend(ctx, stage.State)
	statePath := filepath.Join("state", ctx.RootParent().Data.GetString(state.WorkspaceDataKey), stage.Id, "state.json")
	obj, err := stateBackend.GetObject(statePath)
	if err != nil {
		// the state could not exist on the first run
		return UpdateStateForStage(ctx, stage, stateBackend), stateBackend

	}
	var st state.State
	err = json.Unmarshal(obj.Content, &st)
	if err != nil {
		panic(err)
	}
	return st, stateBackend

}
