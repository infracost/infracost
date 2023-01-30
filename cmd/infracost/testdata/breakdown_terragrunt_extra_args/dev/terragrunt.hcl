terraform {
  extra_arguments "variables" {
    commands = get_terraform_commands_that_need_vars()

    optional_var_files = [
      find_in_parent_folders("dev.tfvars")
    ]
  }
}
