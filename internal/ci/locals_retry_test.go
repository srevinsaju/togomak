package ci

import "testing"

func TestLocal_CanRetry(t *testing.T) {
	local := Local{}
	if local.CanRetry() {
		t.Error("CanRetry() should return false")
	}
}

func TestLocal_Description(t *testing.T) {
	local := Local{}
	if local.Description().Description != "" {
		t.Error("Description() should return empty string")
	}
}

func TestLocal_MaxRetries(t *testing.T) {
	local := Local{}
	if local.MaxRetries() != 0 {
		t.Error("MaxRetries() should return 0")
	}
}

func TestLocal_MaxRetryBackoff(t *testing.T) {
	local := Local{}
	if local.MaxRetryBackoff() != 0 {
		t.Error("MaxRetryBackoff() should return 0")
	}
}

func TestLocal_MinRetryBackoff(t *testing.T) {
	local := Local{}
	if local.MinRetryBackoff() != 0 {
		t.Error("MinRetryBackoff() should return 0")
	}
}

func TestLocal_RetryExponentialBackoff(t *testing.T) {
	local := Local{}
	if local.RetryExponentialBackoff() {
		t.Error("RetryExponentialBackoff() should return false")
	}
}
