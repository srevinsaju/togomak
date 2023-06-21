togomak {
  version = 1
}
stage "example" {
  if     = this.what
  name   = "example"
  script = "echo hello world"
}
