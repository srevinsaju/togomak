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

stage "another" {
  env {
    name = "MY_NEW_HOME"
    value = data.env.HOME.value
  }
  script = "echo My new home is $MY_NEW_HOME"
}
