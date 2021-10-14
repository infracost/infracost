package aws_test

import (
	"fmt"
	"testing"

	resources "github.com/infracost/infracost/internal/resources/aws"
	"github.com/stretchr/testify/assert"
)

func stubEC2DescribeImages(stub *stubbedAWS, ami string, usageOp string) {
	body := fmt.Sprintf(`Action=DescribeImages&ImageId.1=%s`, ami)
	response := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
		<DescribeImagesResponse xmlns="http://ec2.amazonaws.com/doc/2016-11-15/">
				<requestId>00000000-0000-0000-0000-000000000000</requestId>
				<imagesSet>
						<item>
								<imageId>%s</imageId>
								<imageLocation>012345678901/stuff/and/things</imageLocation>
								<imageState>available</imageState>
								<imageOwnerId>012345678901</imageOwnerId>
								<creationDate>1970-01-01T00:00:00Z</creationDate>
								<isPublic>true</isPublic>
								<architecture>x86_64</architecture>
								<imageType>machine</imageType>
								<sriovNetSupport>simple</sriovNetSupport>
								<name>stuff/and/things</name>
								<description>Stuff and things</description>
								<rootDeviceType>ebs</rootDeviceType>
								<rootDeviceName>/dev/sda1</rootDeviceName>
								<blockDeviceMapping>
								</blockDeviceMapping>
								<virtualizationType>hvm</virtualizationType>
								<hypervisor>xen</hypervisor>
								<enaSupport>true</enaSupport>
								<platformDetails>Stubbed</platformDetails>
								<usageOperation>%s</usageOperation>
						</item>
				</imagesSet>
		</DescribeImagesResponse>`, ami, usageOp)
	stub.WhenBody(body).Then(200, response)
}

func TestInstanceOS(t *testing.T) {
	stub := stubAWS(t)
	defer stub.Close()

	tests := [][]string{
		{"ami-0ceee60bcb94f60cd", "RunInstances", "linux"},
		{"ami-0227c65b90645ae0c", "RunInstances:0002", "windows"},
		{"ami-0d5d1eef89f668c58", "RunInstances:0010", "rhel"},
		{"ami-0d9b2196bf4232757", "RunInstances:000g", "suse"},
	}

	for _, test := range tests {
		ami, op, os := test[0], test[1], test[2]
		stubEC2DescribeImages(stub, ami, op)

		args := resources.Instance{AMI: ami}
		resource := args.BuildResource()
		estimates := newEstimates(stub.ctx, t, resource)
		assert.Equal(t, os, estimates.usage["operating_system"])
	}
}
