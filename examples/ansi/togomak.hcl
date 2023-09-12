togomak {
  version = 2
}
stage "example" {
  name   = "example"
  script = "echo ${ansifmt("green", "hello world in green")}"
}
