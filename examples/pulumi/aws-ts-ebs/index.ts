import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";
import * as awsx from "@pulumi/awsx";

const bucket = new aws.s3.Bucket("mybucket");
const namePrefix = 'example'
const vpc = new awsx.ec2.Vpc(`${namePrefix}-vpc`, {
    cidrBlock: '10.0.0.0/22',
    numberOfAvailabilityZones: 3,
    numberOfNatGateways: 2,
  });

const ebsVolume = new aws.ebs.Volume(`${namePrefix}-ebs-volume`, {
    availabilityZone: "us-west-2a",
    size: 40,
    tags: {
        Name: "HelloWorld",
    },
});

const ebsVolume1 = new aws.ebs.Volume(`${namePrefix}-ebs-volume-1`, {
    availabilityZone: "us-west-2a",
    size: 40,
    tags: {
        Name: "HelloWorld",
    },
});

const ubuntu = aws.ec2.getAmi({
    mostRecent: true,
    filters: [
        {
            name: "name",
            values: ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"],
        },
        {
            name: "virtualization-type",
            values: ["hvm"],
        },
    ],
    owners: ["099720109477"],
});

const eip = new aws.ec2.Eip(`${namePrefix}-elastic-ip`, {
    vpc: true,
});

const web = new aws.ec2.Instance(`${namePrefix}-ec2-instance`, {
    ami: ubuntu.then(ubuntu => ubuntu.id),
    instanceType: "t3.micro",
    tags: {
        Name: "HelloWorld",
    },
    rootBlockDevice: {
        volumeSize: 40,
        volumeType: "gp3"
    },
    creditSpecification: {
        cpuCredits: "unlimited",
    },
    ebsBlockDevices: [
        { deviceName: '/dev/xvde', volumeId: ebsVolume.id},
        { deviceName: '/dev/xvdf', volumeId: ebsVolume1.id}
    ]
});

const eipAssoc = new aws.ec2.EipAssociation(`${namePrefix}-eipAssoc`, {
    instanceId: web.id,
    allocationId: eip.id,
});


const dbAppSecurityGroup = new aws.ec2.SecurityGroup(`${namePrefix}-dbAppAccessGroup`, {
    vpcId: vpc.id
})

const rdsInstance = new aws.rds.Instance(`${namePrefix}-rds`, {
    allocatedStorage: 40,
    dbSubnetGroupName: dbAppSecurityGroup.name,
    engine: "mysql",
    engineVersion: "8.0.28",
    instanceClass: "db.t3.small",
    iops: 0,
    backupRetentionPeriod: 7,
    backupWindow: "00:00-01:00",
    maintenanceWindow:  "Mon:02:00-Mon:04:00",
    monitoringInterval: 0,
    monitoringRoleArn: "",
    optionGroupName: "",
    parameterGroupName:"",
    password: "example1234!",
    username: "dbAdmin",
    dbName: "example",
    storageType: "gp2",
    skipFinalSnapshot: true,
    vpcSecurityGroupIds: [],
}, );

var azs = pulumi.output(vpc.privateSubnets).apply((subnets) => subnets.map((s) => s.subnet.availabilityZone))

const rdsCluster = new aws.rds.Cluster(`${namePrefix}-rds-cluster`, {
    availabilityZones: azs,
    backupRetentionPeriod: 5,
    clusterIdentifier: "aurora-cluster-demo",
    databaseName: "mydb",
    engine: "aurora-mysql",
    engineVersion: "5.7.mysql_aurora.2.03.2",
    masterPassword: "example1234!",
    masterUsername: "foo",
    preferredBackupWindow: "07:00-09:00",
    dbClusterInstanceClass: "db.r6gd.xlarge",
    iops: 1000,
    allocatedStorage: 1000
});

const clusterInstances: aws.rds.ClusterInstance[] = [];
for (const range = {value: 0}; range.value < 2; range.value++) {
    clusterInstances.push(new aws.rds.ClusterInstance(`${namePrefix}-rds-clusterInstances-${range.value}`, {
        identifier: `aurora-cluster-demo-${range.value}`,
        clusterIdentifier: rdsCluster.id,
        instanceClass: "db.r4.large",
        engine: "aurora-mysql",
        engineVersion: rdsCluster.engineVersion,
    }));
}

const exampleRole = new aws.iam.Role(`${namePrefix}-iam-role`, {assumeRolePolicy: `{
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
        "Principal": {
          "Service": "eks.amazonaws.com"
        },
        "Action": "sts:AssumeRole"
      }
    ]
  }
  `});
  const example_AmazonEKSClusterPolicy = new aws.iam.RolePolicyAttachment(`${namePrefix}-AmazonEKSClusterPolicy`, {
      policyArn: "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy",
      role: exampleRole.name,
  });
  // Optionally, enable Security Groups for Pods
  // Reference: https://docs.aws.amazon.com/eks/latest/userguide/security-groups-for-pods.html
  const example_AmazonEKSVPCResourceController = new aws.iam.RolePolicyAttachment(`${namePrefix}AmazonEKSVPCResourceController`, {
      policyArn: "arn:aws:iam::aws:policy/AmazonEKSVPCResourceController",
      role: exampleRole.name,
  });

const eksCluster = new aws.eks.Cluster(`${namePrefix}-eks-cluster`, {
    roleArn: exampleRole.arn,
    vpcConfig: {
        subnetIds: vpc.privateSubnetIds,
    },
}, {
    dependsOn: [
        example_AmazonEKSClusterPolicy,
        example_AmazonEKSVPCResourceController,
    ],
});

const example_AmazonEKSWorkerNodePolicy = new aws.iam.RolePolicyAttachment("example-AmazonEKSWorkerNodePolicy", {
    policyArn: "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy",
    role: exampleRole.name,
});
const example_AmazonEKS_CNI_Policy = new aws.iam.RolePolicyAttachment("example-AmazonEKSCNIPolicy", {
    policyArn: "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy",
    role: exampleRole.name,
});
const example_AmazonEC2ContainerRegistryReadOnly = new aws.iam.RolePolicyAttachment("example-AmazonEC2ContainerRegistryReadOnly", {
    policyArn: "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly",
    role: exampleRole.name,
});

const eksNodeGroup = new aws.eks.NodeGroup("example", {
    clusterName: eksCluster.name,
    nodeRoleArn: exampleRole.arn,
    subnetIds: vpc.privateSubnetIds,
    scalingConfig: {
        desiredSize: 1,
        maxSize: 1,
        minSize: 1,
    },
    updateConfig: {
        maxUnavailable: 2,
    },
    instanceTypes: ["t3.large"]
}, {
    dependsOn: [
        example_AmazonEKSWorkerNodePolicy,
        example_AmazonEKS_CNI_Policy,
        example_AmazonEC2ContainerRegistryReadOnly,
    ],
});