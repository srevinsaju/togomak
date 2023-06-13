# Concurrency 

Togomak uses concurrency, by default. The number of stages
run in parallel, is theoritically infinity (see [goroutines][so_goroutines]).
The number of stages run in parallel depends on the [graph][depgraph]
derived from the pipeline. 

Say, four stages are completely independent of each of other.
The number of stages run concurrently would be 4.
Similarly, if stage A runs independently, but B, C, D depends 
on A, the maximum number of stages that will run in parallel would 
be 3. 

When `togomak -n` or `togomak --dry-run` is used, concurrency 
is disabled, so that the command output remains readable. 

[so_goroutines]: https://stackoverflow.com/questions/8509152/max-number-of-goroutines
[depgraph]: ./dependency.md




