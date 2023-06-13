# Reusable Stages

Sometimes, you might have run into a scenario where you would like to 
repeat the same boring task multiple times in the same pipeline. 

Your first thought might be a shell script, with custom arguments 
as parameters, but maintaining these pipelines over multiple places
would be a hassle. A single pipeline is already complicated enough in my opinion. 

`togomak` provides a feature called `macros` which are re-usable stages.
You can write a stage once, with a set of parameters, and re-use them 
in multiple stages later. Let's see how:

```tf
{{#include ../../../examples/macros/togomak.hcl}}

```
