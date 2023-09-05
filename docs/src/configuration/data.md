![since v1.0.0](https://img.shields.io/badge/since-v1.0.0-green)
![config v1](https://img.shields.io/badge/config-%201-green)

# Data

The `data` block allows fetching static data from different sources known 
as `data_providers`. Common providers include examples like environment variables,
file sources, or an external secrets manager, like Hashicorp Vault or Google 
Secret Manager. 

A basic `data` block looks like this:
```hcl 
data "provider_name" "id" {
    ...
    default = "???"
}
```

Here, the data block uses the provider, `provider_name` to retrieve information.
The retrieved value will be stored, and can be accessed `data.provider_name.id.value`.

## Data Providers
### Built-in providers
Parameters for the `data` block depends on the type of the provider. 
Supported buit-in providers are:
* [`data.env`](./data_env.md): Environment Variable Provider 
* [`data.prompt`](./data_prompt.md): Interactive user CLI prompt 
* [`data.file`](./data_file.md): Read data from a file

Refer to the documentation by the provider type, for more information
on the supported attributes.
