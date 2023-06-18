togomak {
  version = 1
}

data "prompt" "name" {
  prompt = "What is your name?"
  default = "Pen Pen"
}

stage "example" {
  name   = "example"
  script = "echo hello ${data.prompt.name.value}"
}
