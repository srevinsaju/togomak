package ci

func (d Data) CanRetry() bool {
	return false
}

func (d Data) MaxRetries() int {
	return 0
}

func (d Data) MinRetryBackoff() int {
	return 0
}
func (d Data) MaxRetryBackoff() int {
	return 0
}

func (d Data) RetryExponentialBackoff() bool {
	return false

}
