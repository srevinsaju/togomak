togomak
=======

> This repository is under active development ⚠️
> Please check back regularly for updates.

A CI/CD which works everywhere, even on your local environment. 

`togomak` is a customizable, extensible CI/CD system which multiple 
backends. It has a plugin based system which allows developers to write 
their custom code for deploying 

`togomak` doesn't aim to be a competitor to other CI/CD systems like 
GitHub Actions or Jenkins, but in fact extend on them, helping to create 
a unified place to track builds across all infrastructure, and to make 
local developers' build and deployment process much easier.

Roadmap
-------
- [x] Locally executable CI/CD 
- [x] Concurrency 
- [x] YAML Configuration (`togomak.yaml`)
- [x] Docker support (`.stages[].container`)
- [x] Matrix Builds (`.matrix`)
- [x] Dependency tree resolution (`.stages[].depends-on`)
- [x] Plugins (`.providers[]`)
- [x] Pipeline Templating (Django like templating, see [`pongo2`](https://github.com/flosch/pongo2))
- [ ] Dry Run (`-n`, `--dry-run`)
- [ ] Git 
  - [ ] Private Git Repositories (`.togomak.git`)
  - [ ] Public Git Repositories (`.stages[].source`)
- [ ] External CI/CD Integration
    - [ ] GitLab CI
    - [ ] GitHub Actions
    - [ ] Jenkins 
- [ ] CI User Interface (`togomak serve`)
- [ ] Tracking Server, helps to track local builds on developer machines
- [ ] Deep Merge (`.stages[].extends`), to inherit properties from other stages
- [ ] Override Build, Stages (`.stages[].overrides.script`, `.stages[].overrides.container`, `.stages[].overrides.args`)
- [ ] Artifacts collection
- [ ] Documentation
- [ ] Releases
- [ ] Parameters 
    - [ ] Prompt
    - [ ] Web User Interface
    - [ ] CLI User Interface
    - [ ] Environment Variables
- [ ] Secrets, although its technically possible to use plugins
    - [ ] Filtering Secrets in output
- [ ] Build Backends
    - [x] Local 
    - [ ] Google Cloud Build
- [ ] Logstash logging
- [ ] Cross stage variable injection
- [ ] Support for templating loops, before evaluation of stages
- [ ] Reusable stages from third party services
- [ ] Logo

    
Building from source 
--------------------
You will require Go 1.18+ to build this project.
```bash
cd cmd/togomak
go build .
```

Or, if you already have `togomak`, just run `togomak` and it can figure out its way 
```bash
togomak
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


### Build Options
Some specific `togomak` configuration can be overriden in the `togomak` section

```yaml 
togomak: 
    chdir: false  # do not automatically change directory to the root where .togomak is stored
    debug: false  # do not show debug logs by default
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