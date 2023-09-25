togomak {
  version = 2
}

stage "another_stage" {
  script = "echo this is coming from module"
}

import {
  source = "./child"
}
