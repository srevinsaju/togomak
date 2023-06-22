package ci

func (s Data) CanRetry() bool {
	return false
}

func (s Data) MaxRetries() int {
	return 0
}

func (s Data) MinRetryBackoff() int {
	return 0
}
func (s Data) MaxRetryBackoff() int {
	return 0
}

func (s Data) RetryExponentialBackoff() bool {
	return false

}
