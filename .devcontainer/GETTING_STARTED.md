Hi Xpiriters 👋

This devcontainer has been included for your convenience and includes the following development environment:

- Go Toolchain
- Terraform
- Azure CLI

Any additional features, vscode extensions or other development tools can be added to the `devcontainer.json` file so we as a community can benefit from it.

Infracost's dependencies have been installed and Infracost has been built from source upon launching this container, a log of this installation can be found at `post-create.log`.

You should now run: `infracost auth login` to set up your Infracost account and configure an API key. After that you should be all set to use and develop on Infracost.

Try running: `infracost breakdown --path examples/terraform --usage-file=examples/terraform/infracost-usage.yml` to get a breakdown of the costs of `../examples/main.tf`.

Isn't that cool? Now that you know how Infracost works, read `../CONTRIBUTING.md` for full details on how to contribute to the codebase.
