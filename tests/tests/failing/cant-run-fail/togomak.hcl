togomak {
  version = 2
}
stage "example" {
  if     = this.what
  name   = "example"
  script = "echo hello world"
}
