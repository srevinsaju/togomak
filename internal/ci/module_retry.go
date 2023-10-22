package ci

func (m *Module) CanRetry() bool {
	if m.Retry != nil {
		return m.Retry.Enabled
	}
	return false
}

func (m *Module) MaxRetries() int {
	return m.Retry.Attempts
}

func (m *Module) MinRetryBackoff() int {
	return m.Retry.MinBackoff
}
func (m *Module) MaxRetryBackoff() int {
	return m.Retry.MaxBackoff
}

func (m *Module) RetryExponentialBackoff() bool {
	return m.Retry.ExponentialBackoff
}
