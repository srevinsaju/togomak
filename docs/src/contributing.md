# Contributing

Contributions are welcome, and encouraged. 
All contributions are licensed under the [MPL License v2.0](https://github.com/srevinsaju/togomak/blob/v1/LICENSE)

## Table of Contents

- [Contribution Workflow](#contribution-workflow)
- [Commit Guidelines](#commit-guidelines)
- [Documentation](#documentation)
- [Code Formatting](#code-formatting)
- [Testing](#testing)
- [Pull Requests](#pull-requests)
- [Code of Conduct](#code-of-conduct)
- [License](#license)

## Contribution Workflow
* Fork the repository to your GitHub account.
* Clone the forked repository to your local machine.
* Create a new branch for your feature or bug fix: `git checkout -b feature-name`.
* Make your changes, following the project's coding style and guidelines.
* Commit your changes using Conventional Commit standards (see guidelines below).
* Push your branch to your forked repository: `git push origin feature-name`.
* Open a pull request against the main repository's `v1` branch.

## Commit Guidelines

We follow the Conventional Commit standard for our commit messages. 
This standard helps us maintain a clean and informative commit history. 
Each commit message should have the following format:

```
<type>: <description>

[optional body]

[optional footer]
```


Here are some examples of commit message types:

- **feat**: A new feature implementation.
- **fix**: A bug fix.
- **docs**: Documentation changes.
- **refactor**: Code refactoring.
- **test**: Adding or modifying tests.
- **chore**: General maintenance tasks (build system updates, dependency management, etc.).

Please make sure to provide a clear and concise description in the commit message. If additional information is necessary, feel free to include a body section or refer to relevant issues or pull requests.

## Documentation
It is important to keep the project's documentation up-to-date. Any changes or additions made to the code should be reflected in the relevant documentation files. If you make any modifications that affect the project's usage or behavior, please update the documentation accordingly.
We use [`mdbook`](https://github.com/rust-lang/mdBook) to generate the documentation.
You can start a live server to preview your changes by running `mdbook serve` in the `docs` directory, or
by using `togomak root +docs_serve` command.

## Code Formatting

To maintain a consistent code style, please run `go fmt` on your Go code before committing. This command automatically formats your code according to the standard Go formatting guidelines.
Or just run `togomak` from time to time.

## Testing
Before submitting your changes, please ensure that all existing tests pass and add new tests when appropriate. Running the test suite helps to verify the correctness and stability of the codebase.

## Pull Requests

When submitting a pull request, please adhere to the following guidelines:

1. Provide a clear and descriptive title for the pull request.
2. Include a summary of the changes made and the motivation behind them.
3. Reference any relevant issues or pull requests in the description.
4. Make sure your code is properly formatted and documented.
5. Verify that all tests pass successfully.
6. Assign the pull request to the appropriate reviewer(s).
7. Be prepared to address feedback and make any necessary changes.

## Code of Conduct
Please note that by participating in this project, you are expected to abide by [GitHub's Code of Conduct](https://docs.github.com/en/site-policy/github-terms/github-community-code-of-conduct).
Be respectful and considerate towards others, and help create a welcoming and inclusive environment for everyone involved.

## License

By contributing to this project, you agree that your contributions will be licensed under the project's [LICENSE](https://github.com/srevinsaju/togomak/blob/v1/LICENSE).






