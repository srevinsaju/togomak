# Outputs 

Sometimes, you would want to generate some outputs 
from your build. Sometimes, you might want to store 
the version you parsed from a file, or a list of files 
that you would want to share with another independent 
stage.

To do so, you can write as environment variables to `$TOGOMAK_OUTPUTS`

for example:
```hcl
{{#include ../../../examples/output/togomak.hcl}}
```

As a limitation, you can only share data within 
the same pipeline scope. i.e, data is not 
implicitly shared between pipelines run using
an remote source, or external file, in the case
of using `macros` with external files. However,
you can still pass them using `macro`'s parameters.




