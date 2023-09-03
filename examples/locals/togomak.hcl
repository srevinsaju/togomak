togomak {
  version = 2
}

locals {
  nerv_headquarters = "Tokyo-3"
  pilot_name        = "Shinji"
}

stage "eva01_synchronization" {
  name   = "Eva-01 Synchronization Tests"
  script = "echo ${local.pilot_name} is now running synchronization tests at ${local.nerv_headquarters}"
}
