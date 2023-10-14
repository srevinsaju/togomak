togomak {
  version = 2
}

data "prompt" "repo_name" {
  prompt  = "enter repo name: "
  default = "username/repo"
}

locals {
  repo        = "srevinsaju/togomak"
  lint_tools  = ["misspell", "golangci-lint", "abcgo"]
  build_types = ["amd64", "i386", "arm64"]
}

stage "lint" {
  script = <<-EOT
  echo 💅 running style checks for repo ${local.repo}
  %{for tool in local.lint_tools}
  echo "* running linter: ${tool}"
  sleep 1
  %{endfor}
  EOT
}


stage "build" {
  script = <<-EOT
  echo 👷 running ${ansifmt("green", "build")}
  %{for arch in local.build_types}
  echo "* building ${local.repo} for ${arch}..."
  sleep 1
  %{endfor}
  EOT
}

stage "deploy" {
  if         = data.prompt.repo_name.value == "srevinsaju/togomak"
  depends_on = [stage.build]
  container {
    image = "hashicorp/terraform"
  }
  args = ["version"]
}
