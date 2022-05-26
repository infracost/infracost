import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";
import * as awsx from "@pulumi/awsx";

const bucket = new aws.s3.Bucket("mybucket");

const ebsVolume = new aws.ebs.Volume("example", {
    availabilityZone: "us-west-2a",
    size: 40,
    tags: {
        Name: "HelloWorld",
    },
});

const ebsVolume1 = new aws.ebs.Volume("example-1", {
    availabilityZone: "us-west-2a",
    size: 40,
    tags: {
        Name: "HelloWorld",
    },
});
