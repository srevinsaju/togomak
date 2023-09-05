togomak {
  version = 2
}

data "env" "hello" {
}

stage "example" {
  name   = "example"
  script = "echo hello world ${data.env.hello.value}"
}
