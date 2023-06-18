togomak {
  version = 1
}
macro "rei" {
  stage "this" {
    script = "hi, im rei"
  }
}

macro "gendo" {
  stage "this" {
    script = "hi, im gendo"
  }
}

stage "example" {
  use {
    macro = [macro.rei, macro.gendo]
  }
}
