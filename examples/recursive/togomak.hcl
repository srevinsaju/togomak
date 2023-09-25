togomak {
  version = 2
}


macro "hello" {
  source = "./hello"
}
macro "bye" {
  source = "./bye"
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
