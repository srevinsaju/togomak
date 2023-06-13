# `data.env`

The `env` data provider fetches the variable from the environment
before the execution of the pipeline happens. If the specified 
environment variable is undefined, falls back to the `default`
value specified in the `data.env` block, else returns an empty 
string.

## Reading an environment variable 


```hcl 
data "env" "user" {
  key = "USER"
  default = "me"
}
```

## Arguments Reference

* [`key`](#key): The environment variable key which needs to be fetched from `os.environ`
* [`default`](#default): Fallback definition, which will be returned if `os.environ[key]` is undefined. 

## Attributes Reference 
* [`value`]: The value from the environment variable defined in `key`


