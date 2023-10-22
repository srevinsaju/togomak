togomak {
  version = 2
}

variable "name" {
  description = "Name of the individual executing this script (or bot)"
}

locals {
  numbers = toset([
    {
      a = 30
      b = 40
    },
    {
      a = 10
      b = 9
    },
    {
      a = 99
      b = 1
    }
  ])
} 

stage "example" {
  name   = "example"
  script = "echo hello world"
}

stage "welcome" {
  script = "echo hello ${var.name}"
}

stage "test" {
  for_each = local.numbers 
  script = "echo ${each.value.a}, ${each.value.b}"
}

module "sum" {
  depends_on = [stage.example, stage.welcome]
  for_each = local.numbers 
  source = "./calculator"

  a = each.value.a
  b = each.value.b 
  operation = "add"
}
