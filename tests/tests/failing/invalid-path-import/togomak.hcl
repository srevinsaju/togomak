togomak {
  version = 2
}
stage "example" {
  name   = "example"
  script = "echo hello world"
}
import {
  source = "git::ssh://git@github.com:codespaces/thisrepo.will-never-exist"
}
