togomak {
  version = 1
}

data "git" "eva01_source" {
  url = "https://github.com/srevinsaju/togomak"
  depth = 1
  files = ["togomak.hcl"]
}

macro "gendo_brain" {
  files = data.git.eva01_source.files 
}

stage "build_eva01" {
  name   = "Building eva unit"
  use {
    macro = macro.gendo_brain
  }

}
