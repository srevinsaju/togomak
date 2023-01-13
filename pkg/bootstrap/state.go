package bootstrap

import (
	"encoding/json"
	"fmt"
	"github.com/chartmuseum/storage"
	uuid "github.com/satori/go.uuid"
	"github.com/srevinsaju/togomak/pkg/context"
	"github.com/srevinsaju/togomak/pkg/schema"
	"github.com/srevinsaju/togomak/pkg/state"
	"github.com/srevinsaju/togomak/pkg/ui"
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

func UnlockAllStates(rootCtx *context.Context, data schema.SchemaConfig) {
	rootCtx.Logger.Info("waiting for all processes to finish")
	for i := range rootCtx.Processes {
		stage := data.Stages.GetStageById(rootCtx.Processes[i].Id)
		rootCtx.Logger.Debug("waiting for stage ", stage.Id)
		p := rootCtx.Processes[i].Process
		if p != nil {
			err := p.Wait()
			if err != nil {
				rootCtx.Logger.Warn(err)
			} else {
				rootCtx.Logger.Debug("stage ", stage.Id, " finished")
			}
		}
		rootCtx.Logger.Debug("unlocking stage ", stage.Id)
		UnlockState(rootCtx, stage, true)
	}
}

func UnlockState(rootCtx *context.Context, stage schema.StageConfig, ignoreErrors bool) {
	if stage.DisableLock {
		return
	}
	ctx := rootCtx.AddChild("state", stage.Id)
	ctx.Logger.Tracef("unlocking state for stage %s", stage.Id)
	workspace := rootCtx.Data.GetString(state.WorkspaceDataKey)
	stateManager := LoadStateBackend(ctx, stage.State)
	stateDir := filepath.Join("state", workspace, stage.Id)
	lockPath := filepath.Join(stateDir, "lock.json")
	err := stateManager.DeleteObject(lockPath)
	if err != nil {
		errFmt := fmt.Sprintf("failed unlocking state %s: %s", stage.Id, err)

		if ignoreErrors {
			ctx.Logger.Debugf(errFmt)
			return
		}
		ctx.Logger.Warnf(errFmt)
	}
}

func LockState(ctx *context.Context, lockPath string, stateBackend storage.Backend) {

	ctx.Logger.Tracef("Writing lock file %s", lockPath)
	err := stateBackend.PutObject(lockPath, []byte{})
	if err != nil {
		ctx.Logger.Fatal(err)
	}
}

func UpdateStateForStage(ctx *context.Context, stage schema.StageConfig, stateManager storage.Backend, init bool) state.State {
	ctx.Logger.Tracef("updating state for stage %s", stage.Id)
	stateDir := filepath.Join("state", ctx.RootParent().Data.GetString(state.WorkspaceDataKey), stage.Id)
	statePath := filepath.Join(stateDir, "state.json")
	//lockPath := filepath.Join(stateDir, "lock.json")
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

func RenderState(st state.State) {
	fmt.Println()
	fmt.Printf("Lock held by %s\n", ui.Yellow(st.LastUsernameHumanReadable))
	fmt.Printf("Last modified at %s\n", ui.Yellow(st.LastModified))
	fmt.Println()
}

func GetStateForStage(ctx *context.Context, stage schema.StageConfig) (state.State, storage.Backend) {
	rootCtx := ctx.RootParent()
	ctx = rootCtx.AddChild("state", stage.Id)
	ctx.Logger.Tracef("Fetching state for stage")
	workspace := rootCtx.Data.GetString(state.WorkspaceDataKey)
	stateBackend := LoadStateBackend(ctx, stage.State)

	stateDir := filepath.Join("state", workspace, stage.Id)
	statePath := filepath.Join(stateDir, "state.json")
	lockPath := filepath.Join(stateDir, "lock.json")

	obj, err := stateBackend.GetObject(statePath)
	if err != nil {
		// the state could not exist on the first run
		s := UpdateStateForStage(ctx, stage, stateBackend, true)
		LockState(ctx, lockPath, stateBackend)
		return s, stateBackend
	}

	var st state.State
	err = json.Unmarshal(obj.Content, &st)
	if err != nil {
		panic(err)
	}

	if stage.DisableLock {
		return st, stateBackend
	}

	// check if the state is locked
	_, err = stateBackend.GetObject(lockPath)
	if err == nil {
		RenderState(st)
		ctx.Logger.Fatalf("The state is locked. Please run `togomak force-unlock --workspace %s %s` to unlock the state", workspace, stage.Id)
	} else {
		// lock the state
		LockState(ctx, lockPath, stateBackend)
	}
	return st, stateBackend
}
