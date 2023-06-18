package ci

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStage_Description(t *testing.T) {
	stage := Stage{}
	assert.Equal(t, stage.Description(), "")
}
