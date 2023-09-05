togomak {
  version = 2
}

macro "explode" {
  stage "explode" {
    script = <<-EOT
        for i in $(seq 1 10); do
          sleep 0.1
          echo "${param.eva}: Loading $i..."
        done

        echo "${param.eva}: entry plug connected! pilot ${param.pilot} synchronized! 🤖"
        EOT
  }
}


stage "entry_plug_eva01" {
  use {
    macro = macro.explode
    parameters = {
      pilot = "Shinji Ikari 🙅‍♂️"
      eva   = "01"
    }
  }
}

stage "entry_plug_eva02" {
  use {
    macro = macro.explode
    parameters = {
      pilot = "Asuka Langley Soryu 🙅‍♀️"
      eva   = "02"
    }
  }
}


