# Stages

An atomic runnable unit may be defined as a stage. GitHub calls it Jobs, Jenkins calls it a stage.
Stages may run in parallel, unless explicit dependencies are provided. 

A simple `stage` definition may be described below 

```yaml
stages:
  - id: hello
    args: ["python", "--version"]

  - id: world
    depends-on:
      - hello
    script: | 
      for i in {1..3}; do
        echo "Concurrency Test $i"
        sleep 1
      done
```

The above example uses two stages which executes one after the other. 
The stage `world` waits for the first stage to complete since it has a `.depends-on` parameter 
set.

## `.stages[].id`
Stage ID is used to uniquely identify the stage in the pipeline. Stage IDs should strictly 
conform to the following regex validation key `^([a-zA-Z0-9_/:.\-]+)$`. Note that `.` and `:`
are shorthand annotations for [`.stages[].extends`](##stages---extends). See `.extends`
documentation on how to use them. 

Stage IDs may be referred to, in `.depends-on` parameter, or even within the pipleine script. 
By default, the running stage will have the parameter `id` available within it's context. For
example:
```yaml
version: 1

...

stages:
    - id: world
      script: echo "hello {{ id }}"
```

This is useful when writing dynamic stages which may be reused several times. See `.extends` for 
detailed usage.

## `.stages[].if`
`.if` accepts a boolean value, or a `pongo2` expression which determines if 
the stage will run. The default value for the `.if` parameter is `true`

```yaml
version: 1

...
parameters:
    - name: ENV
      default: dev

stages:
    - id: world
      if: {{ param.ENV == "dev" }}
      script: echo "DEBUG=true" >> .env
```

## `.stages[].name`
Name of the stage. Will be pretty printed, and replaced in output before the stage 
starts. Useful if the viewers of the build logs need more context on what's happening
and if the `.id` of stage is not self-explanatory.

## `.stages[].description`
Description of the stage. A short blob of text which explains what the pipeline does. 

> **Planned**: `.stages[].description` and `.stages[].name` may be used for `togomak-docs` 
> which can be used for automatically generating Markdown docs explaining the pipeline workflow. 

## `.stages[].container`
Container name under which the `.stages[].script` or `.stages[].args` should be executed under. 
If container name is unspecified, the code will run on the host, without containerization.
In the case of backend providers other than the localhost provider, an applicable `busybox` 
container image will be replaced. 

```yaml
stages:
    - id: world
      container: python:latest
      script: |
        python --version
```


## `.stages[].script`
Blob text, which gets executed under the default shell of the container if defined under
`.stages[].container` or, on the host's default shell. 

```yaml
stages: 
    - id: world
      script: |
        echo "hello"
```

> **Note**: It is recommended to set `set -euo pipefail` on bash, or `set -eu` to not deal with 
> ambigious build statuses. Read why `set -euo pipefail` might be necessary [here](https://coderwall.com/p/fkfaqq/safer-bash-scripts-with-set-euxo-pipefail)

## `.stages[].args`
A yaml array of parameters which needs to be passed to the docker container, or
the command will be run on shell. 

```yaml
stages:
    - id: world
      args: ["echo", "hello", "world"]
```
or 
```yaml
stages:
    - id: world
      container: "python:latest"
      args:
      - "--version"
```




