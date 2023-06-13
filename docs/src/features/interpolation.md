# Interpolation

String interpolation is a native feature of togomak.
It uses similarly named functions from Terraform, and has a lot of helper scripts
to achieve the same.

Let's take an example:
```hcl 

stage "hello_world" {
    script = "echo ${upper('hello world')}"
}
```

The above stage prints `HELLO WORLD` to standard output, thanks to the [upper](../functions.md#upper)
helper function.

Similarly, you can even do math!

```hcl 
stage "i_can_calculate" {
    script = "echo 1 plus 2 is ${1 + 2}"
}
```

For an incomplete list of functions, take a look at the detailed [`functions`](../functions.md) page.
