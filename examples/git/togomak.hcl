togomak {
  version = 1
}

data "git" "repo" {
  url = "https://github.com/srevinsaju/togomak"
  files = ["togomak.hcl"]
  depth = 1
}

stage "example" {
  name   = "example"
  script = <<-EOT
  echo '${data.git.repo.files["togomak.hcl"]}'
  EOT
}
