# Upgrade from v2.x to v3.x

If you have any questions regarding this upgrade process, please consult the `examples` directory:

- [Complete](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/tree/master/examples/complete)
- [Volume Attachment](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/tree/master/examples/volume-attachment)

If you find a bug, please open an issue with supporting configuration to reproduce.

## List of backwards incompatible changes

- Terraform v0.13.1 is now minimum supported version to take advantage of `count` and `for_each` arguments at module level

### Variable and output changes

1. Removed variables:

   - `instance_count`
   - `subnet_ids` (only need to use `subnet_id` now)
   - `private_ips` (only need to use `private_ip` now)
   - `use_num_suffix`
   - `num_suffix_format`

2. Renamed variables:

   - `tags` -> `tags_all`

3. Removed outputs:

   - `availability_zone`
   - `placement_group`
   - `key_name`
   - `ipv6_addresses`
   - `private_ip`
   - `security_groups`
   - `vpc_security_group_ids`
   - `subnet_id`
   - `credit_specification`
   - `metadata_options`
   - `root_block_device_volume_ids`
   - `ebs_block_device_volume_ids`
   - `volume_tags`
   - `instance_count`

4. Renamed outputs:

   :info: All outputs used to be lists, and are now singular outputs due to the removal of `count`

## Upgrade State Migrations

### Before 2.x Example

```hcl
module "ec2_upgrade" {
  source  = "terraform-aws-modules/ec2-instance/aws"
  version = "2.21.0"

  instance_count = 3

  name                        = local.name
  ami                         = data.aws_ami.amazon_linux.id
  instance_type               = "c5.large"
  subnet_ids                  = module.vpc.private_subnets
  vpc_security_group_ids      = [module.security_group.security_group_id]
  associate_public_ip_address = true

  tags = local.tags
}
```

### After 3.x Example

```hcl
locals {
  num_suffix_format = "-%d"
  multiple_instances = {
    0 = {
      num_suffix    = 1
      instance_type = "c5.large"
      subnet_id     = element(module.vpc.private_subnets, 0)
    }
    1 = {
      num_suffix    = 2
      instance_type = "c5.large"
      subnet_id     = element(module.vpc.private_subnets, 1)
    }
    2 = {
      num_suffix    = 3
      instance_type = "c5.large"
      subnet_id     = element(module.vpc.private_subnets, 2)
    }
  }
}

module "ec2_upgrade" {
  source = "../../"

  for_each = local.multiple_instances

  name = format("%s${local.num_suffix_format}", local.name, each.value.num_suffix)

  ami                         = data.aws_ami.amazon_linux.id
  instance_type               = each.value.instance_type
  subnet_id                   = each.value.subnet_id
  vpc_security_group_ids      = [module.security_group.security_group_id]
  associate_public_ip_address = true

  tags = local.tags
}
```

To migrate from the `v2.x` version to `v3.x` version example shown above, the following state move commands can be performed to maintain the current resources without modification:

```bash
terraform state mv 'module.ec2_upgrade.aws_instance.this[0]' 'module.ec2_upgrade["0"].aws_instance.this[0]'
terraform state mv 'module.ec2_upgrade.aws_instance.this[1]' 'module.ec2_upgrade["1"].aws_instance.this[0]'
terraform state mv 'module.ec2_upgrade.aws_instance.this[2]' 'module.ec2_upgrade["2"].aws_instance.this[0]'
```

:info: Notes

- In the `v2.x` example we use `subnet_ids` which is an array of subnets. These are mapped to the respective instance based on their index location; therefore in the `v3.x` example we are doing a similar index lookup to map back to the existing subnet used for that instance. This would also be the case for `private_ips`
- In the `v3.x` example we have shown how users can continue to use the same naming scheme that is currently in use by the `v2.x` module. By moving the `num_suffix_format` into the module name itself inside a format function, users can continue to customize the names generated in a similar manner as that of the `v2.x` module.
