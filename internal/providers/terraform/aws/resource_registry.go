package aws

import "github.com/infracost/infracost/pkg/schema"

var ResourceRegistry []*schema.RegistryItem = []*schema.RegistryItem{
	GetAutoscalingGroupRegistryItem(),
	GetDBInstanceRegistryItem(),
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
	GetNATGatewayRegistryItem(),
	GetRDSClusterInstanceRegistryItem(),
}
