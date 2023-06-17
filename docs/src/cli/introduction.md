# Command Line Interface

- [Usage](./cli/usage.md)

Togomak, internally features three modes of execution:
* `ci` (`--ci`)
* `unattended` (`--unattended`)
* Normal execution

In `unattended` mode, all prompts and interactive options will be disabled. 
`ci` mode is same as unattended mode, but however, a separate variable will
be defined in the pipeline `togomak.ci`, which can be used later to run 
specific stages on CI, and only run certain stages on the user side. 

The `ci` mode is implicit on popular CICD providers like Jenkins, GitHub or 
GitLab CI, etc. It requires one of the environment variables: `CI`, `TOGOMAK_CI`
or the CLI flag `--ci` for it to be enabled.

In the normal execution mode, prompts will pass through the pipeline execution, 
and it will wait indefinitely until the user enters the value.
Similarly, Interrupts will be enabled on the normal execution mode.

On the first `Ctrl + C` recevied from the user in the normal execution mode, 
it will send `SIGTERM` to child stages, or will stop the docker container, in the 
case of docker container engines. The deadline in most cases would be 10 seconds.

If a second `Ctrl + C` is received within the above deadline time, the child processes
will be sent `SIGKILL` signal. In the case of docker containers, no action will be
taken. This means that, it is possible that containers might be left dangling.
Similarly, it is possible to have zombie processes.


