togomak {
  version = 1
}

stage "example_a" {
  depends_on = [stage.example_b]
  script = "echo hello world"
}

stage "example_b" {
  depends_on = [stage.example_a]
  script = "echo hello again"
}
