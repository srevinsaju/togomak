# Usage

## Basic Usage
By default, `togomak` runs all stages 
which evaluates their condition `stage.xxx.if` to 
`true`. By default, all stages evaluate their
condition to true, unless explicitly 
specified. 

To simply run all stages which meet the criteria,
just do
```bash
togomak 
```

and you should be good to go.

If your togomak pipeline lives in a different
directory, you could:
```bash
togomak -C path/to/different/dir
```

Similary, you can also explicitly specify 
the path to `togomak.hcl` using the `-f` or 
the `--file` parameter.

## Running specific stages

If you would like to run only specific 
stages, instead of the entire pipeline,
you could do something like this:

```bash 
togomak stage1 stage2 
```

This would run both `stage1`, `stage2` and 
all the dependencies of `stage1` and `stage2`.
That means, if there were a stage `stage3`
which depends on `stage1`, then `stage3` would
also be included in the same pipeline.

Now, you can also blacklist and whitelist 
certain stages. Let us take the specific
example of `togomak.hcl` which is used
to build togomak itself, at the root of 
this repository:

```hcl
{{#include ../../../togomak.hcl}}
```

In the above example, doing `togomak build`
would run both `stage.build` and `stage.install`.

However, if you would like to run only `stage.build`
and not `stage.install`, you could do:

```bash
togomak build ^install
```



The `^` operator is used as a blacklist operator.
Similarly, if you would like to add the `stage.fmt`
along with the stages which run, you would do:

```bash
togomak build ^install +fmt
```

Here, the `+` operator, is used as a whitelist operator.
You can add multiple `+{stage_name}` and `^{stage_name}`
and togomak would run all the stages which meet the
criteria.

### Running a specific stage alone
If you strictly want to run a single stage,
and do not want to include its dependencies,
or if you do not want to manually blacklist 
all its dependencies, `togomak` has a 
special stage called the `root` stage,
which will run regardless of the whitelist or the 
blacklist.

So, if you would like to run only `stage.build`
and not `stage.install`, you could do:

```bash
togomak root +build
```
This translates to:
* Run only the `root` stage (and its dependencies, which are `nil`)
* Whitelist `build` stage

> Adding multiple whitelist and blacklist entries
> for the same stage will take no effect. The first
> entry will be considered and the rest will be ignored.

#### Whitlisting stages from a macro
If you would like to whitelist stages from a macro,
you could do so by using the `+` operator.

```bash
togomak macro_name.build
```

or 

```bash
togomak macro_name.root +macro_name.build
``` 

and so on.


## Daemonize stages
> Experimental feature, use with caution.

If you would like to run a stage in the background,
you could do so by adding the `&` operator to the
stage name.

```bash 
togomak &build
```

This would run the `build` stage in the background,
and will not wait for the build stage to complete
before moving on to the next stage. However, 
the `build` stage will be terminated once the last
dependent of the `build` stage completes.

Under the hood, the `build` stage will receive
a `SIGTERM` signal, and will be given a grace period
of 10 seconds to complete. If the stage does not
terminate within the grace period, it will be
forcefully killed (`SIGKILL`).

The above feature is particularly useful for 
running long running stages, such as a docker 
service for performing integration tests.

For extensive information about the daemonizing
feature, refer to the [daemonizing](../features/daemonizing.md) 
section.
