package ci

import "testing"

func TestData_CanRetry(t *testing.T) {
	data := Data{}
	if data.CanRetry() {
		t.Error("CanRetry() should return false")
	}
}

func TestData_Description(t *testing.T) {
	data := Data{}
	if data.Description() != "" {
		t.Error("Description() should return empty string")
	}
}

func TestData_MaxRetries(t *testing.T) {
	data := Data{}
	if data.MaxRetries() != 0 {
		t.Error("MaxRetries() should return 0")
	}
}

func TestData_MaxRetryBackoff(t *testing.T) {
	data := Data{}
	if data.MaxRetryBackoff() != 0 {
		t.Error("MaxRetryBackoff() should return 0")
	}
}

func TestData_MinRetryBackoff(t *testing.T) {
	data := Data{}
	if data.MinRetryBackoff() != 0 {
		t.Error("MinRetryBackoff() should return 0")
	}
}

func TestData_RetryExponentialBackoff(t *testing.T) {
	data := Data{}
	if data.RetryExponentialBackoff() {
		t.Error("RetryExponentialBackoff() should return false")
	}
}
