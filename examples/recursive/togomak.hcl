togomak {
  version = 1
}


macro "hello" {
  source = "./hello.hcl"
}
macro "bye" {
  source = "./bye.hcl"
}

stage "hello_phase" {
  use {
    macro = macro.hello
    parameters = {
      target = "eva-01"
    }
  }
}

stage "bye_phase" {
  depends_on = [
    stage.hello_phase
  ]
  use {
    macro = macro.bye
    parameters = {
      target = "world"
    }
  }
}
