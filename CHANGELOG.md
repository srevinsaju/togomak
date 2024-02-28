# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v2.0.0-alpha.16]
- Add `stage.*.container.skip_workspace` boolean parameter to skip mounting the current working directory when using the docker plugin

## [v2.0.0-alpha.14]
- Fixes `depends_on` not being respected on modules.

## [v2.0.0-alpha.13]
- Changes the behavior of module lifecycles. By default all modules will be run if lifecycle.phase is unspecified. 

## [v2.0.0-alpha.12]
- Add `TOGOMAK_ARGS` environment variable.

## [v2.0.0-alpha.11]
- Add `--logging.local.file` and `--logging.local.file.path` for writing logs to file.

## [v2.0.0-alpha.10]
- Fix `path.module` incorrectly being populated for togomak modules

## [v2.0.0-alpha.9]
- Add support for `modules`
- Add JSON logger
- Add `for_each` to modules
- Add `variable` block

## [v2.0.0-alpha.4]
- Add support for `for_each` meta argument

## [Older releases]
> see `git log` for previous release history, or use GitHub releases page
