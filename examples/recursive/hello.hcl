togomak {
  version = 1
}

stage "echo" {
  script = "echo hello ${param.target}"
}
