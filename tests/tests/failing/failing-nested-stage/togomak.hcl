togomak {
  version = 1
}

macro "nested" {
  source = "./nested/togomak.hcl"
}


stage "root_function" {
  use {
    macro = macro.nested
  }
}

stage "another_function" {
  depends_on = [stage.root_function]
  script = "echo done"
}
