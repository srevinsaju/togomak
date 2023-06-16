# Locals

If you plan to use a specific variable several times in your pipeline, you may use 
a `locals {}` block. 

A simple usage may be shown below


```hcl
locals {
    var1 = 1
    var2 = "hello"
    var3 = {
        apple = 2
        orange = 3
    }
}
```

When referring to them in stages, use `local.<variable_name>`, for example:
```hcl 
{{#include ../../../examples/locals/togomak.hcl}}
```
