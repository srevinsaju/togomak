togomak {
  version = 1
}
locals {
  planets = toset(["mercury", "venus", "earth", "mars", "jupiter", "saturn", "uranus", "neptune"])
  #planets = toset(["mercury", "venus", "earth", "mars", "jupiter", "saturn", "uranus", "neptune"])
}

stage "example" {
  for_each = local.planets
  name   = "example"
  script = "echo hello world ${each.value}"
}
