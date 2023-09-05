togomak {
  version = 2
}

import {
  source = "./module"
}

stage "main" {
  script = "echo script from the main file"
}

import {
  source = "git::https://github.com/srevinsaju/togomak-first-module"
}
