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

stage "tests" {
  pre_hook {
    stage {
      script = "echo ${ansi.fg.green}${each.key}${ansi.reset}: full"
    }
  }

  depends_on = [stage.build, stage.coverage_prepare]
  for_each   = fileset(cwd, "../examples/*/togomak.hcl")
  args = [
    "./togomak_coverage",
    "-C", dirname(each.key),
    "--ci", "-v", "-v", "-v",
  ]

  env {
    name  = "GOCOVERDIR"
    value = local.coverage_data_dir
  }
  env {
    name  = "TOGOMAK_VAR_name"
    value = "bot"
  }
}


stage "tests_dry_run" {
  pre_hook {
    stage {
      script = "echo ${ansi.fg.green}${each.key}${ansi.reset}: dry"
    }
  }

  depends_on = [stage.build, stage.coverage_prepare]
  for_each   = fileset(cwd, "../examples/*/togomak.hcl")
  args = [
    "./togomak_coverage",
    "-C", dirname(each.key),
    "--ci", "-v", "-v", "-v", "-n",
  ]

  env {
    name  = "GOCOVERDIR"
    value = local.coverage_data_dir
  }

  env {
    name  = "TOGOMAK_VAR_name"
    value = "bot"
  }
}

stage "fmt" {
  depends_on = [stage.build, stage.coverage_prepare]
  script     = "./togomak_coverage fmt --check --recursive"
}

stage "cache" {
  depends_on = [stage.fmt, stage.tests, stage.tests_dry_run]
  script     = "./togomak_coverage cache clean --recursive"
}


stage "failing" {
  depends_on = [stage.cache]
  for_each   = fileset(cwd, "tests/failing/*/togomak.hcl")
  script     = <<-EOT
  set +e
  ./togomak_coverage -C "${dirname(each.key)}" --ci -v -v -v
  result=$?
  if [ $result -eq 0 ]; then 
      set -e
      echo "$i completed successfully when it was supposed to fail"
      exit 1
  fi 
  EOT 
  env {
    name  = "GOCOVERDIR"
    value = local.coverage_data_dir
  }
}

stage "coverage_raw" {
  depends_on = [stage.tests]
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

