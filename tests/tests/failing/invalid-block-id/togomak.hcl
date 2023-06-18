togomak {
  version = 1
}

data "env" "hello_world" {
  key = "HOME"
  default = "@"
}

stage "example" {
  script = "echo ${data.env.hello_world.value}"
}

stage "example_2" {
  script = "echo ${data.env.hello_hello_world.value}"
}

