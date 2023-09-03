togomak {
  version = 2
}

data "env" "HOME" {
  key = "HOME"
}

stage "example" {
  name   = "example"
  script = "echo ${data.env.HOME.value}"
}
