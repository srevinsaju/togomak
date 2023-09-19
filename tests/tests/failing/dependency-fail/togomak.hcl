togomak {
  version = 2
}

stage "test" {
  script = "exit 1"
}

stage "example" {
  depends_on = [stage.test]
  script = "exit 0"
}
