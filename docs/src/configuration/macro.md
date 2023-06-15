# Macro

Macros are reusable stages. If you would like to reuse the same stage multiple times 
in the same pipeline, with optionally different parameters, `macros` are the right 
thing for you.

### Example usage 

```tf
{{#include ../../../examples/macros/togomak.hcl}}
```

## Argument reference 
* [`stage`](#stage) - The stage that will be reused, optional. Structure [documented below](#stage)
* [`source`](#source) - Path to a different togomak file, which will be recursively invoked.

---
<a href="stage"></a>
The `stage` block supports:
* All features under the [`stage`](../stage.md), except `id`, `name`, `description`



