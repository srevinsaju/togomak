togomak {
  version = 1
}


stage "example" {
  retry {
    enabled = true
    exponential_backoff = true
    attempts = 3
    min_backoff = 1
    max_backoff = 3
  }
  script = "echo hello world && exit 1"
}
