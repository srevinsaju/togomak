togomak {
  version = 1
}

macro "explode" {
    stage "explode" {
        script = "echo This is an exploding stage! BOOM! ${param.item}"
    }
}


stage "explode_party_poppers" {
    use {
        macro = macro.explode 
        parameters = {
            item = "Party! 🎉"
        }
    }
}

stage "explode_water_balloon" {
    use { 
        macro = macro.explode 
        parameters = {
            item = "Balloon! 🎈"
        }
    }
}


