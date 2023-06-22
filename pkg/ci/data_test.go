package ci

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestData_Kill(t *testing.T) {
	data := Data{}
	if data.Kill() != nil {
		t.Error("Kill() should do nothing")
	}
}

func TestData_Terminate(t *testing.T) {
	data := Data{}
	if data.Terminate(false) != nil {
		t.Error("Terminate() should do nothing")
	}
}

func TestData_Type(t *testing.T) {
	data := Data{}
	if data.Type() != DataBlock {
		t.Error("Type() should return data")
	}
}

func TestData_IsDaemon(t *testing.T) {
	data := Data{}
	if data.IsDaemon() {
		t.Error("IsDaemon() should return false")
	}
}

func TestData_Set(t *testing.T) {
	data := Data{}
	data.Set("key", "value")
}

func TestData_Get(t *testing.T) {
	data := Data{}
	assert.Equal(t, data.Get("key"), nil)
}
