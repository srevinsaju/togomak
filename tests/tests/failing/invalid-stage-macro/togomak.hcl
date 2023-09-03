togomak {
  version = 2
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
