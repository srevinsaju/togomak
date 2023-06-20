togomak {
  version = 1
}

stage "tests" {
  if = false
  script = "togomak -C tests --ci"
}

stage "fmt" {
  script = "go fmt github.com/srevinsaju/togomak/v1/..."
}
stage "vet" {
  script = "go vet github.com/srevinsaju/togomak/v1/..."
}
stage "build" {
  depends_on = [stage.fmt, stage.vet]
  script = "go build -v -o ./cmd/togomak/togomak github.com/srevinsaju/togomak/v1/cmd/togomak"
}
stage "install" {
  depends_on = [stage.build]
  script = "go install github.com/srevinsaju/togomak/v1/cmd/togomak"
}

stage "docs_serve" {
  daemon {
    enabled = true
  }
  if = false
  script = "cd docs && mdbook serve"
}
