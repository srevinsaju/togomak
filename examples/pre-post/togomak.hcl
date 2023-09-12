togomak {
  version = 2
}

pre {
  script = "echo first stage to be executed"
}

post {
  script = "echo last stage to be executed"
}

stage "example" {
  script = "echo hello world"
}

stage "example_2" {
  depends_on = [stage.example]
  script = "echo hello world 2"
}
