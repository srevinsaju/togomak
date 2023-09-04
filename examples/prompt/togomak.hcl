togomak {
  version = 2
}

data "env" "quit_if_not_shinji" {
  key     = "QUIT_IF_NOT_SHINJI"
  default = "false"
}

data "prompt" "name" {
  prompt  = "What is your name?"
  default = "Pen Pen"
}

stage "example" {
  name   = "example"
  script = "echo hello ${data.prompt.name.value}"
}

stage "quit" {
  if     = data.env.quit_if_not_shinji.value != "false"
  script = <<-EOT
  #!/usr/bin/env bash
  set -eux
  if [[ "${data.prompt.name.value}" != "Shinji Ikari" ]]; then
    echo "${data.prompt.name.value}" >> /tmp/quit
    exit 1
  fi
  EOT

}
