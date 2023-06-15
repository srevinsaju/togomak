package ci

func (l *Local) CanRetry() bool {
	return false
}

func (l *Local) MaxRetries() int {
	return 0
}

func (l *Local) MinRetryBackoff() int {
	return 0
}
func (l *Local) MaxRetryBackoff() int {
	return 0
}

func (l *Local) RetryExponentialBackoff() bool {
	return false
}
