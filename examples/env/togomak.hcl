togomak {
  version = 1
}

data "env" "HOME" {
  key = "HOME"
}

stage "example" {
  name   = "example"
  script = "echo ${data.env.HOME.value}"
}
