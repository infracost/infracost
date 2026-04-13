# Cost-Optimized AWS Infrastructure Example

This example demonstrates cost-optimized AWS infrastructure configurations using Infracost. It shows how small changes can lead to significant cost savings while maintaining performance.

## 💰 Cost Optimization Strategies Demonstrated

| Resource | Original | Optimized | Savings |
|----------|----------|-----------|---------|
| EC2 Instance | m5.4xlarge | m5.2xlarge | ~50% |
| EBS Volume | io1 1000GB | gp3 500GB | ~75% |
| Lambda Memory | 1024 MB | 512 MB | ~50% |

## 🚀 Usage

```bash
# View cost breakdown for optimized configuration
infracost breakdown --path .

# Compare with non-optimized configuration
infracost diff --path . --compare-to ../terraform
```

## 📊 Expected Costs (us-east-1)

### Original Configuration
- EC2 (m5.4xlarge): ~$560/month
- EBS (io1 1000GB): ~$1,250/month
- Lambda: Usage-dependent

### Optimized Configuration
- EC2 (m5.2xlarge): ~$280/month
- EBS (gp3 500GB): ~$40/month
- Lambda: Usage-dependent

**Total Monthly Savings: ~$1,490 (~65%)**

## 📝 Optimization Notes

1. **Right-sizing EC2**: Changed from m5.4xlarge to m5.2xlarge. Monitor CPU utilization to ensure the smaller instance meets your needs.

2. **EBS gp3 instead of io1**: gp3 provides better price/performance for most workloads. Only use io1/io2 for IOPS-intensive applications.

3. **Lambda memory tuning**: Lower memory reduces cost per invocation. Test your Lambda functions to find the optimal memory setting.

4. **Reserved Instances**: For predictable workloads, consider Reserved Instances or Savings Plans for additional 30-60% savings.

## 🔍 Validation

Always validate optimizations with performance testing before applying to production.
