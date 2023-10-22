package ci

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMacro_CanRetry(t *testing.T) {
	macro := Macro{}
	if macro.CanRetry() {
		t.Error("CanRetry() should return false")
	}
}

func TestMacro_Description(t *testing.T) {
	macro := Macro{}
	if macro.Description().Description != "" {
		t.Error("Description() should return empty string")
	}
}

func TestMacro_MaxRetries(t *testing.T) {
	macro := Macro{}
	if macro.MaxRetries() != 0 {
		t.Error("MaxRetries() should return 0")
	}
}

func TestMacro_MaxRetryBackoff(t *testing.T) {
	macro := Macro{}
	if macro.MaxRetryBackoff() != 0 {
		t.Error("MaxRetryBackoff() should return 0")
	}
}

func TestMacro_MinRetryBackoff(t *testing.T) {
	macro := Macro{}
	if macro.MinRetryBackoff() != 0 {
		t.Error("MinRetryBackoff() should return 0")
	}
}

func TestMacro_RetryExponentialBackoff(t *testing.T) {
	macro := Macro{}
	if macro.RetryExponentialBackoff() {
		t.Error("RetryExponentialBackoff() should return false")
	}
}

func TestMacro_Set(t *testing.T) {
	data := Macro{}
	data.Set("key", "value")
}

func TestMacro_Get(t *testing.T) {
	data := Macro{}
	assert.Equal(t, data.Get("key"), nil)
}
