togomak {
  version = 2
}

stage "normal" {
  script = "echo hi"
}

stage "dont_execute" {
  if = false
  script = "echo this shouldnt be executed && exit 1"
}

stage "docker" {
  lifecycle {
    phase = ["build"]
  }
  script = "echo docker run ..."
}

stage "terraform_fmt_check" {
  lifecycle {
    phase = ["deploy", "default"]
  }
  script = "echo terraform ..."
}

stage "terraform" {
  depends_on = [stage.docker]
  lifecycle {
    phase = ["deploy"]
  }
  script = "echo terraform fmt -check ..."
}

