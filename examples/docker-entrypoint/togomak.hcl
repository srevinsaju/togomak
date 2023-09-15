togomak {
  version = 2
}

stage "apt" {
  container {
    image = "ubuntu:latest"
    entrypoint = ["apt"]
  }
  args = ["install"]
}
