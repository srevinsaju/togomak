# Dependency Resolution

togomak uses native dependency resolution, based on references within their attributes. 
Let's take a look at this example:


```hcl
stage "alice" {
    script = "echo hello world"
}

stage "bob" {
    script = "echo waiting for ${stage.alice.id}"
}
```

In the above example, `stage.alice` naturally becomes a dependency of `stage.bob`.
However, you can also explicitly specify dependencies if you would like to:


```hcl 

stage "alice" {
    script = "hello world"
}

stage "bob" {
    depends_on = [stage.alice]
    script = "hello bob"
}
```

## Data evaluation
[Data blocks](../configuration/data.md) are lazily evaluated. 
Data blocks will be only evaluated before the stage requiring them, gets
executed. 

```hcl
stage "bob" {
    script = "echo Hello World"
}

data "env" "bob_name" {
    default = "Bob Ross"
}

stage "alice" {
        script = "echo This is an environment variable: ${data.env.bob_name.value}"
}
```

In the above example, the order of execution would be `stage.bob` and `env.bob_name` in parallel, and then `stage.alice`.

If you would like to disable concurrency, and let all the execution happen synchronously and linearly,
you can disable concurrency on [`togomak.pipeline.concurrency`](../configuration/togomak.md#concurrency) options. 




