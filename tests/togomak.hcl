togomak {
  version = 2
}


stage "build" {
  name   = "build"
  dir    = ".."
  script = "go build -cover -o tests/togomak_coverage ./cmd/togomak"
}

locals {
  coverage_data_dir             = "${cwd}/coverage_data_files"
  coverage_merge_dir            = "${cwd}/coverage_merge_dir"
  coverage_data_interactive_dir = "${cwd}/coverage_data_interactive_dir"
}

stage "coverage_prepare" {
  script = <<-EOT
  set -e
  rm -rf ${local.coverage_data_dir} && mkdir ${local.coverage_data_dir}
  rm -rf ${local.coverage_data_interactive_dir} && mkdir ${local.coverage_data_interactive_dir}
  rm -rf ${local.coverage_merge_dir} && mkdir ${local.coverage_merge_dir}
  EOT
}

stage "integration_tests" {
  depends_on = [stage.build, stage.coverage_prepare]
  script     = <<-EOT
  #!/usr/bin/env bash
  set -e
  ls ../examples
  for i in ../examples/*; do
    echo ${ansi.fg.green}$i${ansi.reset}
    ./togomak_coverage -C "$i" --ci -v
    ./togomak_coverage -C "$i" --ci -v root
    ./togomak_coverage -C "$i" --ci -v -n
  done
  ./togomak_coverage cache clean --recursive
  ./togomak_coverage fmt --check --recursive

  for i in tests/failing/*; do 
    set +e
    echo ${ansi.fg.green}$i${ansi.reset}
    ./togomak_coverage -C "$i" --ci -v
    result=$?
    if [ $result -eq 0 ]; then 
      set -e
      echo "$i completed successfully when it was supposed to fail"
      exit 1
    fi
  done
  EOT

  env {
    name  = "GOCOVERDIR"
    value = local.coverage_data_dir
  }
}

stage "coverage_raw" {
  depends_on = [stage.integration_tests]
  script     = "go tool covdata percent -i=${local.coverage_data_dir}"
}
stage "coverage_merge" {
  depends_on = [stage.coverage_raw, stage.coverage_unit_tests]
  script     = "go tool covdata merge -i=${local.coverage_data_dir},${local.coverage_data_interactive_dir} -o=${local.coverage_merge_dir}"
}
stage "coverage" {
  depends_on = [stage.coverage_merge]
  script     = "go tool covdata textfmt -i=${local.coverage_merge_dir} -o=coverage.out"
}
stage "coverage_unit_tests" {
  depends_on = [stage.build]
  dir        = ".."
  script     = "go test ./... -coverprofile=coverage_unit_tests.out"
  env {
    name  = "PROMPT_GOCOVERDIR"
    value = local.coverage_data_interactive_dir
  }
}

