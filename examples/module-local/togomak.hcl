 togomak {
  version = 2
}

locals {
  a = 99
  b = 1
}

module "add" {
  source = "./calculator"
  a = local.a
  b = local.b
  operation = "add"
}

