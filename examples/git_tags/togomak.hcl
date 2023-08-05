togomak {
  version = 1
}

data "git" "repo" {
  url   = "https://github.com/srevinsaju/togomak"
  files = ["togomak.hcl"]
  tag = "v1.2.0"
}

stage "example" {
  name   = "example"
  script = <<-EOT
  echo '${data.git.repo.files["togomak.hcl"]}'
  EOT

}