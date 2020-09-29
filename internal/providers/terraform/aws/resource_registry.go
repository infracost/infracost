package aws

import "github.com/infracost/infracost/pkg/schema"

var ResourceRegistry []*schema.RegistryItem = []*schema.RegistryItem{
	GetAutoscalingGroupRegistryItem(),
	GetLaunchConfigurationRegistryItem(),
	GetLaunchTemplateRegistryItem(),
	GetDBInstanceRegistryItem(),
	GetDocDBClusterInstanceRegistryItem(),
	GetDynamoDBTableRegistryItem(),
	GetEBSSnapshotCopyRegistryItem(),
	GetEBSSnapshotRegistryItem(),
	GetEBSVolumeRegistryItem(),
	GetECSServiceRegistryItem(),
	GetECSClusterRegistryItem(),
	GetECSTaskDefinitionRegistryItem(),
	GetElasticsearchDomainRegistryItem(),
	GetELBRegistryItem(),
	GetInstanceRegistryItem(),
	GetLambdaFunctionRegistryItem(),
	GetLBRegistryItem(),
	GetALBRegistryItem(),
	GetNATGatewayRegistryItem(),
	GetRDSClusterInstanceRegistryItem(),
	GetRDSClusterRegistryItem(),
}
