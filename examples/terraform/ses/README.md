# AWS SES (Simple Email Service) Example

This example demonstrates how to use Infracost with AWS SES resources.

## Overview

AWS Simple Email Service (SES) is a cost-effective, flexible, and scalable email service that enables developers to send mail from within any application.

## Resources Included

- `aws_ses_configuration_set` - Configuration sets for email sending
- `aws_ses_domain_identity` - Domain identity for SES
- `aws_ses_email_identity` - Email address identity for SES  
- `aws_ses_template` - Email templates

## Pricing Components

| Component | Description | Billing |
|-----------|-------------|---------|
| Outbound Email | Emails sent through SES | Per 1,000 emails |
| Inbound Email | Emails received | Per 1,000 emails |
| Attachments | Email attachments | Per GB |
| Dedicated IP | Dedicated IP address | Monthly |

## Usage

```bash
# Run Infracost breakdown
infracost breakdown --path .

# Run with usage file
infracost breakdown --path . --usage-file infracost-usage.yml
```

## Example Output

```
 Name                                          Monthly Qty  Unit            Monthly Cost

 aws_ses_configuration_set.main
 └─ Outbound emails                                    10  1,000 emails           $1.00

 aws_ses_dedicated_ip_pool.main
 └─ Dedicated IP address                                1  hours                 $24.95

 OVERALL TOTAL                                                                            $25.95
```

## References

- [AWS SES Pricing](https://aws.amazon.com/ses/pricing/)
- [Terraform AWS SES Resources](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ses_configuration_set)
- [Infracost Issue #2998 - Add support for some SES and SESv2 resources](https://github.com/infracost/infracost/issues/2998)
