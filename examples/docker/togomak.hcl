togomak {
  version = 1
}

stage "example" {
  container {
    image = "ubuntu"
  }
  script = <<-EOT
  #!/usr/bin/env bash
  apt update && apt install -y neofetch
  neofetch
  EOT
}
