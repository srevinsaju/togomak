togomak {
  version = 1
}

locals {
  target = "from togomak"
}

stage "example" {
  name   = "example"
  script = "echo hello ${local.target}"
}
