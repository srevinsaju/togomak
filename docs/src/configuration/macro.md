# Macro

Macros are reusable stages. If you would like to reuse the same stage multiple times 
in the same pipeline, with optionally different parameters, `macros` are the right 
thing for you.

You may use `macro` in three ways:
* **Inline stages** (`stage` block): The stage is defined in the macro block, and re-used multiple times
* **External files** (`source` argument): Path to an external single pipeline located on the filesystem, which will
be passed to an internal togomak child process, which will independently run the single file as a separate stage. 
See [Reusable Stages section](../features/macros.md) for more information.
* **Pipeline content** (`files` argument): A map containing files inline, which will be used by togomak, to create 
another child process which will run the contents of the file, as an independant stage. 

### Example usage (Inline stages)

```tf
{{#include ../../../examples/macros/togomak.hcl}}
```

## Argument reference 
* [`stage`](#stage) - The stage that will be reused, optional. Structure [documented below](#stage)
* [`source`](#source) - Path to a different togomak file, which will be recursively invoked.
* [`files`](#files) - a map containing key value pairs of file paths to file content. Map [documented below](#files)

---
<a href="stage"></a>
The `stage` block supports:
* All features under the [`stage`](../stage.md), except `id`, `name`, `description`
---
<a href="files"></a>
The `files` is a map, which accepts data in the format of key-value pairs, where the "key" is the path to the file 
and the "value" is the content of the file. 

For example,
```hcl
files = {
    "togomak.hcl" = <<-EOT
    togomak {
        version = 1
    }
    stage "hello" {
        script = "echo hello world"
    }
    EOT,
    ...
}
```






