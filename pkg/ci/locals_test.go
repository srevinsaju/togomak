package ci

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLocals_Description(t *testing.T) {
	locals := Locals{}
	if locals.Description() != "" {
		t.Error("Description() should return empty string")
	}
}

func TestLocal_Kill(t *testing.T) {
	local := Local{}
	if local.Kill() != nil {
		t.Error("Kill() should return nil")
	}
}

func TestLocal_Terminate(t *testing.T) {
	local := Local{}
	if local.Terminate(false) != nil {
		t.Error("Terminate() should return nil")
	}
}

func TestLocal_IsDaemon(t *testing.T) {
	local := Local{}
	if local.IsDaemon() {
		t.Error("IsDaemon() should return false")
	}
}

func TestLocal_Identifier(t *testing.T) {
	local := Local{Key: "test"}
	if local.Identifier() != "local.test" {
		t.Error("Identifier() should return 'local.test'")
	}
}

func TestLocal_Type(t *testing.T) {
	local := Local{}
	if local.Type() != LocalBlock {
		t.Error("Type() should return 'local'")
	}
}

func TestLocal_Set(t *testing.T) {
	data := Local{}
	data.Set("key", "value")
}
func TestLocal_Get(t *testing.T) {
	data := Local{}
	assert.Equal(t, data.Get("key"), nil)
}
