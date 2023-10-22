package ci

func (v *Variable) CanRetry() bool {
	return false
}

func (v *Variable) MinRetryBackoff() int {
	return 0
}

func (v *Variable) MaxRetryBackoff() int {
	return 0
}

func (v *Variable) RetryExponentialBackoff() bool {
	return false
}

func (v *Variable) MaxRetries() int {
	return 0
}
