package ci

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStage_Override(t *testing.T) {
	assert.Equal(t, (&Stage{}).Override(), false)
}

func TestStages_CheckIfDistinct(t *testing.T) {
	stage1 := Stages{
		Stage{
			Id: "stage1",
		},
		Stage{
			Id: "stage2",
		},
	}
	stage2 := Stages{
		Stage{
			Id: "stage3",
		},
		Stage{
			Id: "stage4",
		},
	}

	assert.Equal(t, stage1.CheckIfDistinct(stage2).HasErrors(), false)
	assert.Equal(t, stage1.CheckIfDistinct(stage1).HasErrors(), true)
}

func TestStages_Override(t *testing.T) {
	assert.Equal(t, (&Stages{}).Override(), false)
}
