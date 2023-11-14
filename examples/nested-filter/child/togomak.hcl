togomak {
  version = 2
}

stage "limited" {
  script = "echo limited"
  
  lifecycle {
    phase = ["build_phase"]
  }
}

stage "example" {
  name   = "example"
  script = "echo hello world"
}
