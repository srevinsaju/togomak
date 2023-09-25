togomak {
  version = 2
}


stage "home" {
  script = "echo ${env("HOME")}"
}

stage "non_existent_env" {
  script = "echo ${env("THIS_SHOULD_NOT_EXIST", "env does not exist")}"
}
