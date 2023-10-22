package ci

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStage_Description(t *testing.T) {
	stage := Stage{}
	assert.Equal(t, stage.Description().Description, "")
}

func TestStage_Set(t *testing.T) {
	stage := Stage{}
	stage.Set("key", "value")
	assert.Equal(t, stage.Get("key"), "value")
}

func TestStage_Get(t *testing.T) {
	stage := Stage{}
	assert.Equal(t, stage.Get("key"), nil)
}
