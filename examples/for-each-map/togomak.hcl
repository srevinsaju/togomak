togomak {
  version = 2
}

locals {
  m = {
    part1   = "You Are (Not) Alone."
    part2   = "You Can (Not) Advance."
    part3   = "You Can (Not) Redo."
    part3-1 = "Thrice Upon a Time."
  }
}


stage "movie" {
  for_each = local.m
  name     = "example"
  script   = <<-EOT
  echo "Evangelion ${each.key}: ${each.value}"
  EOT
}
