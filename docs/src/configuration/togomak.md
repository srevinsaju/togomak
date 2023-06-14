# Togomak

The `togomak` block is a mandatory 
block for the file to be recognized as a valid
`togomak` pipeline. 

```hcl 
# togomak.hcl

togomak {
    # ...
    version = 1
}
```
The above block, with `version` parameter 
is required for any pipeline to be considered
"runnable".

## Arguments Reference
* [`version`](#version) - The version of the Togomak pipeline file

## Attributes Reference

* [`version`](#version) - The version of the Togomak runner executable
* [`boot_time`](#boot_time) - The time at which the Togomak runner executable was started
* [`boot_time_unix`](#boot_time_unix) - The time at which the Togomak runner executable was started, in unix format
* [`pipeline_id`](#pipeline_id) - The unique identifier for the pipeline run (uuid)
