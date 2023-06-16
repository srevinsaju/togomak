togomak {
  version = 1
}

data "git" "repo" {
  url = "https://github.com/srevinsaju/togomak"
  depth = 1
  files = ["togomak.hcl"]
}

macro "togomak" {
  files = data.git.repo.files 
}

stage "example" {
  name   = "example"
  use {
    macro = macro.togomak
  }

}
