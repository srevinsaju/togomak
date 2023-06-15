togomak {
  version = 1
}

stage "echo" {
  script = "echo bye ${param.target}"
}
