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
      target = "world"
    }
  }
}

stage "bye_phase" {
  use {
    macro = macro.bye
    parameters = {
      target = "world"
    }
  }
}
