togomak {
  version = 2
}


stage "example" {
  script = "echo hello world"
  
  pre_hook {
    stage "echo" {
      script = "echo before the script for stage ${this.id} runs"
    }
  }

  post_hook {
    stage "echo" {
      script = "echo the script for ${this.id} done with status ${this.status}"
    }
  }
}

stage "example_2" {
  script = "echo bye world"
  
  pre_hook {
    stage "echo" {
      script = "echo before the script for stage ${this.id} runs"
    }
  }

  post_hook {
    stage "echo" {
      script = "echo the script for ${this.id} done with status ${this.status}"
    }
  }
}
