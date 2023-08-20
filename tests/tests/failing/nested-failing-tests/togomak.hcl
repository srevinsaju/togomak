togomak {
  version = 1
}

macro "nested" {
  files = {
    "togomak.hcl": file("./nested.hcl")
  }
}

stage "example" {
  use {
    macro = macro.nested 
  }
}

stage "another" {
  depends_on = [stage.example]
  script = "echo done"
}