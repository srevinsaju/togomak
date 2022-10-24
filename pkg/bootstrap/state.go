package bootstrap

import (
	"encoding/json"
	"fmt"
	"github.com/chartmuseum/storage"
	"github.com/srevinsaju/togomak/pkg/ui"

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

func UnlockState(rootCtx *context.Context, stage schema.StageConfig) {
	ctx := rootCtx.AddChild("state", stage.Id)
	workspace := rootCtx.Data.GetString(state.WorkspaceDataKey)
	stateManager := LoadStateBackend(ctx, stage.State)
	stateDir := filepath.Join("state", workspace, stage.Id)
	lockPath := filepath.Join(stateDir, "lock.json")
	err := stateManager.DeleteObject(lockPath)
	if err != nil {
		ctx.Logger.Fatal(err)
	}
}

func UpdateStateForStage(ctx *context.Context, stage schema.StageConfig, stateManager storage.Backend, init bool) state.State {
	stateDir := filepath.Join("state", ctx.RootParent().Data.GetString(state.WorkspaceDataKey), stage.Id)
	statePath := filepath.Join(stateDir, "state.json")
	lockPath := filepath.Join(stateDir, "lock.json")
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
	if !init {
		err = stateManager.DeleteObject(lockPath)
		if err != nil {
			ctx.Logger.Fatal(err)
		}
	}

	return st
}

func RenderState(st state.State) {
	fmt.Println()
	fmt.Printf("Lock held by %s\n", ui.Yellow(st.LastUsernameHumanReadable))
	fmt.Printf("Last modified at %s\n", ui.Yellow(st.LastModified))
	fmt.Println()
}

func GetStateForStage(ctx *context.Context, stage schema.StageConfig) (state.State, storage.Backend) {
	rootCtx := ctx.RootParent()
	ctx = rootCtx.AddChild("state", stage.Id)
	workspace := rootCtx.Data.GetString(state.WorkspaceDataKey)
	stateBackend := LoadStateBackend(ctx, stage.State)

	stateDir := filepath.Join("state", workspace, stage.Id)
	statePath := filepath.Join(stateDir, "state.json")
	lockPath := filepath.Join(stateDir, "lock.json")

	obj, err := stateBackend.GetObject(statePath)
	if err != nil {
		// the state could not exist on the first run
		return UpdateStateForStage(ctx, stage, stateBackend, true), stateBackend
	}

	var st state.State
	err = json.Unmarshal(obj.Content, &st)
	if err != nil {
		panic(err)
	}
	// check if the state is locked
	_, err = stateBackend.GetObject(lockPath)
	if err == nil {
		RenderState(st)
		ctx.Logger.Fatalf("The state is locked. Please run `togomak force-unlock %s --workspace %s` to unlock the state", stage.Id, workspace)
	} else {
		// lock the state
		err = stateBackend.PutObject(lockPath, []byte{})
		if err != nil {
			ctx.Logger.Fatal(err)
		}
	}

	return st, stateBackend

}
