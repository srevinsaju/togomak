buildsys
========

> This repository is under active development ⚠️
> Please check back regularly for updates.

A CI/CD which works everywhere, even on your local environment. 

`buildsys` is a customizable, extensible CI/CD system which multiple 
backends. It has a plugin based system which allows developers to write 
their custom code for deploying 

Building from source 
--------------------
You will require Go 1.18+ to build this project.
```bash
cd cmd/togomak
go build .
```

Running the binary
------------------
```bash
./togomak ./config.yaml
```

## Concepts 

### Providers
Providers are plugins, which can help in 
* Gathering information 
* Checking if all the preconditions of task are met
* Running a task

The data from the providers from the "Gather information" step can be used 
in other stages 

A provider can be defined in the config file as follows:

```yaml
providers:
  - id: git
    path: plugins/git/git
```
> Remote plugins are still a WIP.

### Stages 
Stages are jobs which happen concurrently by default.  
They can be run in parallel or sequentially.
If you need stages to execute sequentially, you can specify the 
`stages[].depends-on` parameters

A sample stage can be defined like this 
```yaml
stages:
    id: myuniqueid # id needs to be unique
    container: python
    args: 
        - "-c"
        - "print('Hello World')"
```

The above stage snipped uses Docker (or Podman) to pull the `python` image
from container registry, and executes the snippet as mentioned in the args. 

You can also run a shell script within a container.
```yaml 
stages:
    id: helloworld
    container: python
    # the above container will run the following script in 'sh'
    script: |
        echo "Hello World"
```

To specify dependencies between stages, you can use the `depends-on` parameter.
`buildsys` will wait for the stages to finish before running the next stage.

```yaml
stages:
    id: myuniqueid
    container: python
    args: 
        - "-c"
        - "print('Hello World')"
    depends-on:
        - helloworld
```



### Configuration 
Configuration is in `yaml`. Some fields currently supported are, some fields 
support `pongo` syntax, which is very similar to django's template syntax. 

Currently supported fields are:
* `.steps[].condition`
* `.steps[].script`
* `.steps[].args`


```yaml
```