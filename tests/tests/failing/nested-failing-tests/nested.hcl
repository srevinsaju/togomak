togomak {
  version = 2
}

stage "test" {
  script = "echo hello world"
}

stage "failing" {
  depends_on = [
    stage.test
  ]
  script = "echo failed && exit 1"
}