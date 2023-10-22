togomak {
  version = 2
}

variable "operation" {
  type = string
  description = "Possible operations: [add, subtract, multiply, divide]"
}
variable "a" {
  type = number
}
variable "b" {
  type = number
}
stage "add" {
  if = var.operation == "add"
  script = "echo sum of ${var.a} and ${var.b} is ${var.a + var.b}"
}
