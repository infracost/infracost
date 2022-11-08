This is an attempt to reproduce the "Azure resources cost breakdown $0.00" reported in https://github.com/infracost/infracost/issues/2108.

It can be executed with the infracost CLI:
```shell
cd examples/terraform_azure_redis_cache

infracost breakdown --path .
```

So far I haven't been able to reproduce the reported issue.  An example of the output I get with Infracost v0.10.13 is saved in `breakdown.txt`.  It was generated with:
```shell
infracost breakdown --path . --out-file=breakdown.txt --no-color
```