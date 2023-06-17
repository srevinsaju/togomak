# Daemonizing

Sometimes, you might want to run a stage in the background, for example, 
an HTTP API server, and continue with multiple integrations tests on 
the same HTTP API server. In another scenario, you might just want 
two live reloading servers for two frontends running in parallel.
Or, maybe you just need a `postgres` database docker container running.
In these cases, `togomak` takes care of all the process management 
required for handling these daemon-like long-running process.

Here is a sample use case, directly from `togomak`'s `togomak.hcl`

```hcl 
{{#include ../../../togomak.hcl}}
```
In the above togomak configuration file, the `mdbook` generator, which 
is used for writing this documentation is allowed to run as a daemon 
process. Similarly, it is possible to have multiple processes running 
as well.


