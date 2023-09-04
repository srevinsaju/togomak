package ci

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestData_Override(t *testing.T) {
	assert.Equal(t, (&Data{}).Override(), false)
}

func TestDatas_CheckIfDistinct(t *testing.T) {
	data1 := Datas{
		Data{
			Id: "data1",
		},
		Data{
			Id: "data2",
		},
	}
	data2 := Datas{
		Data{
			Id: "data3",
		},
		Data{
			Id: "data4",
		},
	}

	assert.Equal(t, data1.CheckIfDistinct(data2).HasErrors(), false)
	assert.Equal(t, data1.CheckIfDistinct(data1).HasErrors(), true)
}
