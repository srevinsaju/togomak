togomak {
  version = 2
}


stage "fmt" {
  script = "go fmt github.com/srevinsaju/togomak/v1/..."
  lifecycle {
    phase = ["default", "validate"]
  }
}

stage "vet" {
  script = "go vet github.com/srevinsaju/togomak/v1/..."
  lifecycle {
    phase = ["default", "validate"]
  }
}

stage "build" {
  depends_on = [stage.fmt, stage.vet]
  script     = "go build -v -o ./cmd/togomak/togomak github.com/srevinsaju/togomak/v1/cmd/togomak"
  lifecycle {
    phase = ["default", "build"]
  }
}

stage "install" {
  depends_on = [stage.build]
  script     = "go install github.com/srevinsaju/togomak/v1/cmd/togomak"
  lifecycle {
    phase = ["default", "install"]
  }
}

stage "docs_serve" {
  daemon {
    enabled = true
  }
  if     = false
  script = "cd docs && mdbook serve"
}
