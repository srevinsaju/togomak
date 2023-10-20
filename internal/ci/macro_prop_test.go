package ci

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMacro_Override(t *testing.T) {
	assert.Equal(t, (&Macro{}).Override(), false)
}

func TestMacros_CheckIfDistinct(t *testing.T) {
	macro1 := Macros{
		Macro{
			Id: "macro1",
		},
		Macro{
			Id: "macro2",
		},
	}
	macro2 := Macros{
		Macro{
			Id: "macro3",
		},
		Macro{
			Id: "macro4",
		},
	}

	assert.Equal(t, macro1.CheckIfDistinct(macro2).HasErrors(), false)
	assert.Equal(t, macro1.CheckIfDistinct(macro1).HasErrors(), true)
}
