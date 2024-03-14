 togomak {
  version = 2
}

locals {
  a = 99
  b = 1
}

stage "paths" {
  script = <<-EOT
  echo path.module: ${path.module}
  echo path.cwd: ${path.cwd}
  echo path.root: ${path.root}
  EOT
}

module "add" {
  source = "./calculator"
  a = local.a
  b = local.b
  operation = "add"

  lifecycle {
    phase = ["add"]
  }
}

