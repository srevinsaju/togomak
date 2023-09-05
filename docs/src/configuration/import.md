
![since v1.3.9](https://img.shields.io/badge/since-v1.3.0-green)
![config v1](https://img.shields.io/badge/config-%201-green)

# Import


The `import` block allows you to merge multiple pipelines together,
into a flatter pipeline. `import` blocks are evaluated before the 
dependency tree is generated, and hence, variables or expressions 
are not permitted. If you need to statisfy a more dynamic evaluation, 
use [`macro`](./macros.md) instead. 

`import` block accepts a single parameter without any id. The block
specifies the source, which should be a constant value, and could 
be a local file, 

## Example usage
```hcl
{{#include ../../../examples/import/togomak.hcl}}
```

## Argument reference 

<a id="source"></a>
* [`source`](#source) - URL to fetch the pipeline from, could be local file path, relative file path, absolute file path, a Git URL, a HTTP URL, GCS bucket, AWS s3 bucket or mercurial. 

The URL for `source` may be link to the required pipeline inclusive of the 
protocol, or you may omit in certain cases, where detectors would 
be run on top of it. 

For example:
* `github.com/srevinsaju/togomak` would automatically evaluate to a SSH URL
* `./file.hcl` would evaluate a relative file in the Togomak current working directory.
* `git::https://github.com/srevinsaju/togomak//examples/docker` would specifically checkout the `examples/docker` folder from the GitHub repository `https://github.com/srevinsaju/togomak.git` over `https` using `git`.

Additionally, you may add query strings like `?ref=abcdefg` to checkout a specific commit of the repository. Check out the documentation of the underlying library, `go-getter` which powers terraform modules here: [hashicorp/go-getter](https://github.com/hashicorp/go-getter)


## What's the difference between `macro` and `import`?

Great question, the following diagram may help you filter out the 
differences between them. 

![macro execution](./img/macros.png)

![import execution](./img/imports.png)
