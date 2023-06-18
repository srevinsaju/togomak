togomak {
  version = 1
}

locals {
  world = "hello"
}

stage "example" {
  use {
    macro = local.world
  }
  script = "echo hello world"
}
