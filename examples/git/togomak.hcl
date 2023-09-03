togomak {
  version = 2
}

data "git" "repo" {
  url   = "https://github.com/srevinsaju/togomak"
  files = ["togomak.hcl"]
}

stage "example" {
  name   = "example"
  script = <<-EOT
  echo '${data.git.repo.files["togomak.hcl"]}'
  EOT

}
