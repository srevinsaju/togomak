# Stage

A `stage` is considered as an atomic runnable
unit. Multiple stages may be executed 
concurrently. 

A stage can accept a script, a command and a set 
of arguments, or a macro. 

### Stage with Script 
```hcl 
~togomak {
~  version = 1
~}
~
stage "script" {
  script = "echo 'Hello World'"
}
```

### Stage with Command and Arguments
```hcl
~togomak {
~  version = 1
~}
~
stage "command" {
  args = ["echo", "Hello World"]
}
```

### Stage with Macro
```hcl
~togomak {
~  version = 1
~}
~
macro "echo" {
  stage "echo" {
    args = ["echo", "Hello World"]
  }
}

stage "macro" {
  use {
    macro = macro.echo     
  }
}
```

### Stage with Dependencies
```hcl
~togomak {
~  version = 1
~}
~
stage "build" {
  script = "echo 'Building'"
}

stage "install" {
  depends_on = [stage.build]
  script = "echo 'Installing'"
}
```

### Stage with Retry
```hcl
~togomak {
~  version = 1
~}
~
stage "build" {
  script = <<-EOT
  echo this script will fail
  exit 1
  EOT
  retry {
    enabled = true
    attempts = 3
    exponential_backoff = true
    min_backoff = 1
    max_backoff = 10
  }
}
```

## Stage with Containers 
```hcl
{{#include ../../../examples/docker/togomak.hcl}}
```



## Arguments Reference
* [`name`](#name) - The name of the stage
* [`if`](#if) - The condition to be evaluated before running the stage
* [`use`](#use) - Macro, or a provider that could be used. Structure is [documented below](#use)
* [`depends_on`](#depends_on) - The stages which this stage depends on.
* [`container`](#container) - The container to be used to run the stage. Structure is [documented below](#container)

* [`script`](#script) - The script to be executed
* [`args`](#args) - The command and arguments to be executed
* [`retry`](#retry) - Stage retry configuration. Structure is [documented below](#retry)
* [`daemon`](#daemon) - Daemon specific configuration. Structure is [documented below](#daemon)
---
<a id="use"></a>
The `use` block supports:
* [`macro`](#macro) - The macro to be used
---
<a id="retry"></a>
The `retry` block supports:

* [`enabled`](#enabled) - Whether the stage should be retried, defaults to `false`
* [`attempts`](#attempts) - The number of times the stage should be retried
* [`exponential_backoff`](#exponential_backoff) - Whether the backoff should be exponential
* [`min_backoff`](#min_backoff) - The minimum backoff time (in seconds)
* [`max_backoff`](#max_backoff) - The maximum backoff time (in seconds)
---
<a id="container"></a>
The `container` block supports:
* [`image`](#image) - The container image to be used
---
<a id="daemon"></a>
> Daemonization is still wip, see [daemonization](../features/daemonizing.md) for more information on availability 

The `daemon` block supports:
* [`enabled`](#enabled) - Whether the stage should be run as a daemon, defaults to `false`
* [`timeout`](#timeout) - Time to wait until the stage is terminated, in seconds. Defaults to 0 (no timeout).
