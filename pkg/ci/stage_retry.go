package ci

func (s *Stage) CanRetry() bool {
	if s.Retry != nil {
		return s.Retry.Enabled
	}
	return false
}

func (s *Stage) MaxRetries() int {
	return s.Retry.Attempts
}

func (s *Stage) MinRetryBackoff() int {
	return s.Retry.MinBackoff
}
func (s *Stage) MaxRetryBackoff() int {
	return s.Retry.MaxBackoff
}

func (s *Stage) RetryExponentialBackoff() bool {
	return s.Retry.ExponentialBackoff

}
