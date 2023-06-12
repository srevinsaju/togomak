# togomak 

togomak is a powerful, build pipeline orchestration tool, and a drop-in 
replacement for your CI/CD execution environment, even runs on your 
laptop. 

`togomak` is inspired from Terraform lifecycle for infrastructure as code (IaC) 
to create a context free, concurrent pipeline evaluation and orchestration engine
to simplify your local builds and your CI/CD pipelines. 

> looking for `togomak v0`? Check [here](https://github.com/srevinsaju/togomak)


### Okay, enough talk. Show me the code-

`togomak` uses [HCL (Hashicorp Language)][hcl] to define pipelines 
declaratively. If you are already familiar with 

```hcl 
togomak {
  version = 1
}

stage "hello" {
    script = "echo hello world"
}
```

simple, isn't it?




