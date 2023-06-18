togomak {
  version = 1
}
stage "example" {
  depends_on = [stage.example]
  name   = "example"
  script = "echo hello world"
}
