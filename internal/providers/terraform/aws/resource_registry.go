package aws

import "github.com/infracost/infracost/internal/schema"

var ResourceRegistry []*schema.RegistryItem = []*schema.RegistryItem{
	GetAutoscalingGroupRegistryItem(),
	GetDBInstanceRegistryItem(),
	GetDMSRegistryItem(),
	GetDocDBClusterInstanceRegistryItem(),
	GetDynamoDBTableRegistryItem(),
	GetEBSSnapshotCopyRegistryItem(),
	GetEBSSnapshotRegistryItem(),
	GetEBSVolumeRegistryItem(),
	GetECSServiceRegistryItem(),
	GetElasticsearchDomainRegistryItem(),
	GetELBRegistryItem(),
	GetInstanceRegistryItem(),
	GetLambdaFunctionRegistryItem(),
	GetLBRegistryItem(),
	GetLightsailInstanceRegistryItem(),
	GetALBRegistryItem(),
	GetNATGatewayRegistryItem(),
	GetRDSClusterInstanceRegistryItem(),
	GetRoute53RecordRegistryItem(),
	GetRoute53ZoneRegistryItem(),
	GetS3BucketRegistryItem(),
	GetS3BucketAnalyticsConfigurationRegistryItem(),
	GetS3BucketInventoryRegistryItem(),
	GetSQSQueueRegistryItem(),
}
