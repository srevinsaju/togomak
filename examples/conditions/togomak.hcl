togomak {
  version = 2
}

data "env" "home" {
  key     = "HOME"
  default = "@"
}

stage "example" {
  if     = data.env.home.value != "@"
  name   = "example"
  script = "echo hello world"
}
