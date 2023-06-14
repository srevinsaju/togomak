# togomak 

![GitHub go.mod Go version (branch)](https://img.shields.io/github/go-mod/go-version/srevinsaju/togomak/v1)


togomak is a powerful, build pipeline orchestration tool, and a drop-in 
replacement for your CI/CD execution environment, even runs on your 
laptop. 

`togomak` is inspired from Terraform lifecycle for infrastructure as code (IaC) 
to create a context free, concurrent pipeline evaluation and orchestration engine
to simplify your local builds and your CI/CD pipelines. 

> looking for `togomak v0`? Check [here][v0]


Okay, enough talk, let's see some code.

## Getting Started

`togomak` uses [HCL (Hashicorp Language)][hcl] to define pipelines 
declaratively. If you are already familiar with Terraform, this becomes
a piece of cake. 

```hcl 
togomak {
  version = 1
}

stage "hello" {
    script = "echo hello world"
}
```

simple, isn't it?

### Documentation
* We have a WIP [documentation](https://togomak.srev.in/v1) (also available over [docs](./docs) directory)
* Check out the [examples](./examples) directory for examples
* Check out the [tests](./tests) directory for more bizarre examples.

### Features (in a nutshell)
* Declarative pipeline definition
* HCL based configuration
* Native dependency resolution
* Concurrency
* Plugins (wip, [use v0][v0] for plugin support)
* Macros (reusable stages)
* Terraform-like data sources

## Installation 
Check out the [releases](https://github.com/srevinsaju/togomak/releases) page
for the `v1.x.y` release binaries, and other pre-built packages for your 
desired platform.

### Building from Source
```bash
cd cmd/togomak 
go build
```
### Building using `togomak` (what!)
```bash 
togomak
```

## Contributing
Contributions are welcome, and encouraged. Please check out the
[contributing](CONTRIBUTING.md) guide for more information.

## License
`togomak` is licensed under the [MPL License v2.0](LICENSE)

[hcl]: https://github.com/hashicorp/hcl
[v0]: https://github.com/srevinsaju/togomak/tree/main



