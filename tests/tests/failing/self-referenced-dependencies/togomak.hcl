togomak {
  version = 2
}
stage "example" {
  depends_on = [stage.example]
  name       = "example"
  script     = "echo hello world"
}
