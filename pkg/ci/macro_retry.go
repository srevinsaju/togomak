package ci

func (m *Macro) CanRetry() bool {
	return false
}

func (m *Macro) MaxRetries() int {
	return 0
}

func (m *Macro) MinRetryBackoff() int {
	return 0
}
func (m *Macro) MaxRetryBackoff() int {
	return 0
}

func (m *Macro) RetryExponentialBackoff() bool {
	return false

}
