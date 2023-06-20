togomak {
  version = 1
}


stage "build" {
  name = "build"
  dir = ".."
  script = "go build -cover -o tests/togomak_coverage ./cmd/togomak"
}

locals {
  // where we will be writing out the profile
  profiles = {
    unit_tests = "${local.coverage.dir}/coverage-unit_tests.out"
  }

  // coverge directory where we will be writing all tests
  coverage = {
    dir = "${cwd}/.coverage"
  }

  // tests that are executed using Netflix/go-expect
  i9n_tests = {
    prompt = "${local.coverage.dir}/i9n/prompt"
    generic = "${local.coverage.dir}/i9n/generic"
  }
  
  tests = {
    must_succeed = fileset(cwd, "../examples/*/togomak.hcl")
    must_fail = fileset(cwd, "/tests/failing/*")
  }
}

stage "echo" {
  for_each = local.tests.must_succeed 
  script = "echo ${each.value}"
}

stage "unit_tests" {
  depends_on = [stage.build, stage.mkdir]
  dir = ".."
  script = "go test ./... -coverprofile=${local.profiles.unit_tests}"
  env {
    name = "PROMPT_GOCOVERDIR"
    value = local.i9n_tests.prompt
  }
}

stage "mkdir" {
  for_each = merge(
    {for s, v in local.i9n_tests: s => dirname(v)},
    {for s, v in local.profiles: s => dirname(v)},
  )
    
  script = "mkdir -p ${each.value}"
}

macro "i9n_tests" {
  stage "this" {
    depends_on = [stage.build]
    for_each = local.tests.must_succeed
    script = "./togomak_coverage -C ${dirname(each.value)} ${param.flags} --ci"
    env {
      name = "GOCOVERDIR"
      value = local.i9n_tests.generic
    }
  }
}

stage "i9n_tests" {
  use {
    macro = macro.i9n_tests
    parameters = {
      flags = "-v"
    }
  }
}

stage "i9n_tests_dry_run" {
  use {
    macro = macro.i9n_tests
    parameters = {
      flags = "-n"
    }
  }
}
