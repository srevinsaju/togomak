# Parameters

The `.parameters` block allow you to specify key-value pairs which
can be used in the CICD script. They only accept string values at the moment, 
although, it is possible to add multiline strings, or even JSON, YAML values
which can parsed at a later stage of the pipeline.

The `.parameters` block has two parameters
* `.parameters[].name`
* `.parameters[].default`

A normal definition of the parameters block on the `togomak.yaml` would look like this:
```yaml
# togomak.yaml 


version: 1

parameters:
  - name: YOUR_NAME
    default: "John Doe"
```

## `.parameters[].name` 
The unique ID or name of the parameter. 

`.name` is required attribute for a parameter for be defined. 

This name will be used throughout the configuration if you would like to refer to the same.
If you would like to use them in a [stage](./stages.md), you would do something like, for example:

```yaml
# togomak.yaml
version: 1 

parameters:
  - name: YOUR_NAME
    default: "John Doe"

stages:
  - id: hello_world
    script: | 
      echo {{ param.YOUR_NAME }}
```

Note that, the above is statically rendered. [Stages](./stages.md) are lazily rendered. 
The value of `param` at the point where the stage named `hello_world` is executed will be 
templated into the script, before its executed. Unlike shell variables which are 
expanded dynamically at the point where the command is executed, stage templating happens
at the beginning of the execution of a stage. See [stages](./stages.md) for more information.

## `.parameters[].default`

`.default` is an optional attribute for the parameters block. If the parameter does 
not have a default value, it will prompt the user over CLI for the values. 


# Value resolution 

Parameter values can be passed in several ways. The following lists the precendence of 
reading terraform parameter values:

## CLI parameters
Togomak supports providing parameters inline through the CLI interface. 

```bash
togomak --parameters "YOUR_NAME=Jon Snow"
```
or 
```bash
togomak -e HELLO=world
```

## *Environment variables**: Togomak reads the value of `TOGOMAK__param__{param_name}`. For example: 
```bash
export TOGOMAK__param__HELLO=world
togomak
```

* **Interactively**: If the default value is unspecified, Togomak, by default prompts 
for the value over a CLI interface, if standard input is accessible. 










