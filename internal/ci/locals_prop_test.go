package ci

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLocal_Override(t *testing.T) {
	assert.Equal(t, (&Local{}).Override(), false)
}

func TestLocals_CheckIfDistinct(t *testing.T) {
	local1 := LocalGroup{
		&Local{
			Key: "local1",
		},
		&Local{
			Key: "local2",
		},
	}

	local2 := LocalGroup{
		&Local{
			Key: "local3",
		},
		&Local{
			Key: "local4",
		},
	}

	assert.Equal(t, local1.CheckIfDistinct(local2).HasErrors(), false)
	assert.Equal(t, local1.CheckIfDistinct(local1).HasErrors(), true)
}
