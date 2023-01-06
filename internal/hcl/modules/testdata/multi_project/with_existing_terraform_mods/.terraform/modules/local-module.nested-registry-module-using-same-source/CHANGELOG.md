# Changelog

All notable changes to this project will be documented in this file.

## [3.4.0](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v3.3.0...v3.4.0) (2022-01-14)


### Features

* Add `instance_metadata_tags` attribute and bump AWS provider to support ([#254](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/254)) ([b2fb960](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/commit/b2fb9604b32aa23d14a8a6e3cff761d6c69661b7))

# [3.3.0](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v3.2.1...v3.3.0) (2021-11-22)


### Features

* Add instance IPv6 addresses to the outputs ([#249](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/249)) ([08bdf6a](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/commit/08bdf6af910f665149d8d64a19175b89a8fac447))

## [3.2.1](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v3.2.0...v3.2.1) (2021-11-22)


### Bug Fixes

* update CI/CD process to enable auto-release workflow ([#250](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/250)) ([1508c9e](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/commit/1508c9ec45ba954828c734326366143a17434a0f))

<a name="v3.2.0"></a>
## [v3.2.0] - 2021-10-07

- feat: Add instance private IP to the outputs ([#241](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/241))


<a name="v3.1.0"></a>
## [v3.1.0] - 2021-08-27

- feat: add support for spot instances via spot instance requests ([#236](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/236))
- chore: update `README.md` example for making an encrypted AMI ([#235](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/235))


<a name="v3.0.0"></a>
## [v3.0.0] - 2021-08-26

- BREAKING CHANGE: update module to include latest features and remove use of `count` for module `count`/`for_each` ([#234](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/234))


<a name="v2.21.0"></a>
## [v2.21.0] - 2021-08-26

- feat: Add support for EBS GP3 throughput ([#233](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/233))


<a name="v2.20.0"></a>
## [v2.20.0] - 2021-08-25

- feat: Add cpu optimization options ([#181](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/181))


<a name="v2.19.0"></a>
## [v2.19.0] - 2021-05-12

- fix: root_block_device tags conflicts ([#220](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/220))


<a name="v2.18.0"></a>
## [v2.18.0] - 2021-05-11

- feat: add tags support to root block device ([#218](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/218))
- chore: update CI/CD to use stable `terraform-docs` release artifact and discoverable Apache2.0 license ([#217](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/217))
- chore: Updated versions&comments in examples
- chore: update documentation and pin `terraform_docs` version to avoid future changes ([#212](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/212))
- chore: align ci-cd static checks to use individual minimum Terraform versions ([#207](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/207))
- chore: add ci-cd workflow for pre-commit checks ([#201](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/201))


<a name="v2.17.0"></a>
## [v2.17.0] - 2021-02-20

- chore: update documentation based on latest `terraform-docs` which includes module and resource sections ([#200](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/200))


<a name="v2.16.0"></a>
## [v2.16.0] - 2020-12-10

- feat: Add support for metadata_options argument ([#193](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/193))


<a name="v2.15.0"></a>
## [v2.15.0] - 2020-06-10

- feat: Add "num_suffix_format" variable for instance naming ([#147](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/147))


<a name="v2.14.0"></a>
## [v2.14.0] - 2020-06-10

- Updated README
- Updated t instance type check to include t3a type ([#145](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/145))
- Updated README
- Instance count as output ([#121](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/121))
- Added user_data_base64 (fixed [#117](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/117)) ([#137](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/137))
- Added support for network_interface and arn ([#136](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/136))
- Update outputs to remove unneeded function wrappers ([#135](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/135))
- Track all changes (remove ignore_changes lifecycle) ([#125](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/125))
- Add encrypted and kms_key_id arguments to the ebs_* and root_* block ([#124](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/124))
- Remove T2 specifics to unify Terraform object names inside TF State ([#111](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/111))
- Fixed output of placement_group (fixed [#104](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/104)) ([#110](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/110))
- Add get_password_data ([#105](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/105))
- Fixed when private_ips is empty (fixed [#103](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/103))
- Added support for the list of private_ips (fixes [#102](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/102)) ([#103](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/103))
- Added support for placement group and volume tags ([#96](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/96))
- Terraform 0.12 update ([#93](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/93))


<a name="v1.25.0"></a>
## [v1.25.0] - 2020-03-05

- Updated t instance type check to include t3a type ([#146](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/146))
- Added placement groups ([#94](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/94))
- Revert example
- Added placement groups
- Add volume tags naming and output ([#82](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/82))


<a name="v2.13.0"></a>
## [v2.13.0] - 2020-03-05

- Updated t instance type check to include t3a type ([#145](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/145))


<a name="v2.12.0"></a>
## [v2.12.0] - 2019-11-19

- Updated README
- Instance count as output ([#121](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/121))


<a name="v2.11.0"></a>
## [v2.11.0] - 2019-11-19

- Added user_data_base64 (fixed [#117](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/117)) ([#137](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/137))


<a name="v2.10.0"></a>
## [v2.10.0] - 2019-11-19

- Added support for network_interface and arn ([#136](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/136))


<a name="v2.9.0"></a>
## [v2.9.0] - 2019-11-19

- Update outputs to remove unneeded function wrappers ([#135](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/135))


<a name="v2.8.0"></a>
## [v2.8.0] - 2019-08-27

- Track all changes (remove ignore_changes lifecycle) ([#125](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/125))


<a name="v2.7.0"></a>
## [v2.7.0] - 2019-08-27

- Add encrypted and kms_key_id arguments to the ebs_* and root_* block ([#124](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/124))


<a name="v2.6.0"></a>
## [v2.6.0] - 2019-07-21

- Remove T2 specifics to unify Terraform object names inside TF State ([#111](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/111))


<a name="v2.5.0"></a>
## [v2.5.0] - 2019-07-08

- Fixed output of placement_group (fixed [#104](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/104)) ([#110](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/110))


<a name="v2.4.0"></a>
## [v2.4.0] - 2019-06-24

- Add get_password_data ([#105](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/105))


<a name="v2.3.0"></a>
## [v2.3.0] - 2019-06-15

- Fixed when private_ips is empty (fixed [#103](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/103))


<a name="v2.2.0"></a>
## [v2.2.0] - 2019-06-14

- Added support for the list of private_ips (fixes [#102](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/102)) ([#103](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/103))


<a name="v2.1.0"></a>
## [v2.1.0] - 2019-06-08

- Added support for placement group and volume tags ([#96](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/96))
- Terraform 0.12 update ([#93](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/93))


<a name="v1.24.0"></a>
## [v1.24.0] - 2019-06-06

- Added placement groups ([#94](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/94))
- Revert example
- Added placement groups


<a name="v1.23.0"></a>
## [v1.23.0] - 2019-06-06

- Add volume tags naming and output ([#82](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/82))


<a name="v2.0.0"></a>
## [v2.0.0] - 2019-06-06

- Terraform 0.12 update ([#93](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/93))


<a name="v1.22.0"></a>
## [v1.22.0] - 2019-06-06

- Update module to the current version ([#88](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/88))


<a name="v1.21.0"></a>
## [v1.21.0] - 2019-03-22

- Fix formatting
- examples/basic/main.tf: Add usage of "root_block_device" ([#18](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/18)) ([#65](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/65))


<a name="v1.20.0"></a>
## [v1.20.0] - 2019-03-22

- Fix formatting
- main.tf: Make number of instances created configurable, defaulting to 1 ([#64](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/64))
- Add missing required field ([#81](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/81))


<a name="v1.19.0"></a>
## [v1.19.0] - 2019-03-01

- Fixed readme after [#76](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/76)
- network_interface_id Attribute Removal ([#76](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/76))


<a name="v1.18.0"></a>
## [v1.18.0] - 2019-02-27

- fix count variables are only valid within resources ([#72](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/72))


<a name="v1.17.0"></a>
## [v1.17.0] - 2019-02-25

- fix call to local.instance_name ([#71](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/71))


<a name="v1.16.0"></a>
## [v1.16.0] - 2019-02-25

- Fixed readme
- Ability to append numerical suffix even to 1 instance ([#70](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/70))


<a name="v1.15.0"></a>
## [v1.15.0] - 2019-02-21

- Allow multiple subnet ids ([#67](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/67))


<a name="v1.14.0"></a>
## [v1.14.0] - 2019-01-26

- Tags should be possible to override (fixed [#53](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/53)) ([#66](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/66))
- fix typo ([#61](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/61))


<a name="v1.13.0"></a>
## [v1.13.0] - 2018-10-31

- Include the module version and some code formatting ([#52](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/52))


<a name="v1.12.0"></a>
## [v1.12.0] - 2018-10-06

- Fixed [#51](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/51). t2 and t3 instances can be unlimited


<a name="v1.11.0"></a>
## [v1.11.0] - 2018-09-04

- Added example of EBS volume attachment (related to [#46](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/46)) ([#47](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/47))
- Ignore changes in the ebs_block_device ([#46](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/46))


<a name="v1.10.0"></a>
## [v1.10.0] - 2018-08-18

- [master]: Narrow t2 selection criteria. ([#44](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/44))


<a name="v1.9.0"></a>
## [v1.9.0] - 2018-06-08

- Fixed t2-unlimited bug (related issue [#35](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/35)) ([#37](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/37))


<a name="v1.8.0"></a>
## [v1.8.0] - 2018-06-04

- Added support for CPU credits ([#35](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/35))


<a name="v1.7.0"></a>
## [v1.7.0] - 2018-06-02

- Added encrypted AMI info ([#34](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/34))


<a name="v1.6.0"></a>
## [v1.6.0] - 2018-05-16

- Added pre-commit hook to autogenerate terraform-docs ([#33](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/33))


<a name="v1.5.0"></a>
## [v1.5.0] - 2018-04-04

- Minor formatting fix
- Modify tag name management to add -%d suffixe only if instance_count > 1


<a name="v1.4.0"></a>
## [v1.4.0] - 2018-04-04

- Stop ignoring changes in vpc_security_group_ids


<a name="v1.3.0"></a>
## [v1.3.0] - 2018-03-06

- Renamed count to instance_count (fixes [#23](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/23))
- Fix: add missing variable to the usage example


<a name="v1.2.1"></a>
## [v1.2.1] - 2018-03-01

- Added aws_eip to example and pre-commit hooks


<a name="v1.2.0"></a>
## [v1.2.0] - 2018-01-19

- Add tags to output variables


<a name="v1.1.0"></a>
## [v1.1.0] - 2017-12-08

- Make module idempotent by requiring subnet_id and ignore changes in several arguments (fixes [#10](https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/issues/10))


<a name="v1.0.4"></a>
## [v1.0.4] - 2017-11-21

- Removed placement_group from outputs


<a name="v1.0.3"></a>
## [v1.0.3] - 2017-11-15

- Fix incorrect subnet_id output expression
- Updated example with all-icmp security group rule


<a name="v1.0.2"></a>
## [v1.0.2] - 2017-10-13

- Added example with security-group module


<a name="v1.0.1"></a>
## [v1.0.1] - 2017-09-14

- Updated example and made security group required


<a name="v1.0.0"></a>
## v1.0.0 - 2017-09-12

- Updated repo name
- Updated README
- Added complete code with example and READMEs
- Initial commit


[Unreleased]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v3.2.0...HEAD
[v3.2.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v3.1.0...v3.2.0
[v3.1.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v3.0.0...v3.1.0
[v3.0.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v2.21.0...v3.0.0
[v2.21.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v2.20.0...v2.21.0
[v2.20.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v2.19.0...v2.20.0
[v2.19.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v2.18.0...v2.19.0
[v2.18.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v2.17.0...v2.18.0
[v2.17.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v2.16.0...v2.17.0
[v2.16.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v2.15.0...v2.16.0
[v2.15.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v2.14.0...v2.15.0
[v2.14.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v1.25.0...v2.14.0
[v1.25.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v2.13.0...v1.25.0
[v2.13.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v2.12.0...v2.13.0
[v2.12.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v2.11.0...v2.12.0
[v2.11.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v2.10.0...v2.11.0
[v2.10.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v2.9.0...v2.10.0
[v2.9.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v2.8.0...v2.9.0
[v2.8.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v2.7.0...v2.8.0
[v2.7.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v2.6.0...v2.7.0
[v2.6.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v2.5.0...v2.6.0
[v2.5.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v2.4.0...v2.5.0
[v2.4.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v2.3.0...v2.4.0
[v2.3.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v2.2.0...v2.3.0
[v2.2.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v2.1.0...v2.2.0
[v2.1.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v1.24.0...v2.1.0
[v1.24.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v1.23.0...v1.24.0
[v1.23.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v2.0.0...v1.23.0
[v2.0.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v1.22.0...v2.0.0
[v1.22.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v1.21.0...v1.22.0
[v1.21.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v1.20.0...v1.21.0
[v1.20.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v1.19.0...v1.20.0
[v1.19.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v1.18.0...v1.19.0
[v1.18.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v1.17.0...v1.18.0
[v1.17.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v1.16.0...v1.17.0
[v1.16.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v1.15.0...v1.16.0
[v1.15.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v1.14.0...v1.15.0
[v1.14.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v1.13.0...v1.14.0
[v1.13.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v1.12.0...v1.13.0
[v1.12.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v1.11.0...v1.12.0
[v1.11.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v1.10.0...v1.11.0
[v1.10.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v1.9.0...v1.10.0
[v1.9.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v1.8.0...v1.9.0
[v1.8.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v1.7.0...v1.8.0
[v1.7.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v1.6.0...v1.7.0
[v1.6.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v1.5.0...v1.6.0
[v1.5.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v1.4.0...v1.5.0
[v1.4.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v1.3.0...v1.4.0
[v1.3.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v1.2.1...v1.3.0
[v1.2.1]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v1.2.0...v1.2.1
[v1.2.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v1.1.0...v1.2.0
[v1.1.0]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v1.0.4...v1.1.0
[v1.0.4]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v1.0.3...v1.0.4
[v1.0.3]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v1.0.2...v1.0.3
[v1.0.2]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v1.0.1...v1.0.2
[v1.0.1]: https://github.com/terraform-aws-modules/terraform-aws-ec2-instance/compare/v1.0.0...v1.0.1
