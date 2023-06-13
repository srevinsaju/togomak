# `data.prompt`

If you interactively need to ask user for an input, you may use 
the `prompt` data source. 
The `prompt` data source interactively asks for the user 
on the Command line interface, if the TTY supports it, or
asks the user on the Web UI (not implemented yet).

It is very likely that, depending on the execution engine, 
`data.prompt` resources are placed at the beginning of the 
topological sorted layer to prevent interference from other 
running stages (in the case of CLIs), or in the case of 
Web UIs to collect all data in the beginning of pipeline 
execution, so that it can happen asynchronously. 

## Prompting a user for response 
```hcl 
data "prompt" "name" {
    prompt = "what is your name?"
    default = "John Doe"
}
```

## Argument Reference 
* [`prompt`](#prompt) - The data that the user will be prompted for (optional)
* [`default`](#default) - Fallback data which will be returned if the user did not enter anything, or if the 
TTY is absent, or if any other UI provider is missing. (optional)


## Attributes Reference 

* [`value`](#value) - The response from the user, otherwise the `default` value will be returned.

