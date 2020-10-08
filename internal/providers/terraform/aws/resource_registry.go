package aws

import "github.com/infracost/infracost/internal/schema"

var ResourceRegistry []*schema.RegistryItem = []*schema.RegistryItem{
	GetAutoscalingGroupRegistryItem(),
	GetLaunchConfigurationRegistryItem(),
	GetLaunchTemplateRegistryItem(),
	GetDBInstanceRegistryItem(),
	GetDMSRegistryItem(),
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
