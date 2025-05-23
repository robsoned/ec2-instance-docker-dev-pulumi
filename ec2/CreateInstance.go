package ec2

import (
	"ec2-instance-docker-dev/ec2/elasticIP"
	"ec2-instance-docker-dev/ec2/userdata"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateInstance(ctx *pulumi.Context, availabilityZone *pulumi.StringInput) (*ec2.Instance, error) {

	elasticIp, err := elasticIP.Create(ctx, &elasticIP.CreateElasticIPArgs{})

	if err != nil {
		return nil, err
	}

	securityGroup, err := getSecurityGroup(ctx, elasticIp.PublicIp)

	if err != nil {
		return nil, err
	}

	ec2InstaceArgs := &ec2.InstanceArgs{
		Ami:                 getAmi(ctx),
		InstanceType:        getInstancetype(ctx),
		KeyName:             getKeyPairName(ctx),
		VpcSecurityGroupIds: pulumi.StringArray{securityGroup.ID()},
		UserData:            userdata.GetInstanceUserData(ctx),
		RootBlockDevice: &ec2.InstanceRootBlockDeviceArgs{
			VolumeSize: pulumi.Int(100),
		},
	}

	if availabilityZone != nil {
		ec2InstaceArgs.AvailabilityZone = *availabilityZone
	}

	ec2Instance, err := ec2.NewInstance(ctx, "ec2-instance", ec2InstaceArgs)

	if err != nil {
		return nil, err
	}

	err = elasticIP.CreateEipAssociation(ctx, ec2Instance.ID().ToStringOutput(), elasticIp.AllocationId)

	if err != nil {
		return nil, err
	}

	ctx.Export("elastic-ip", elasticIp.PublicIp)

	return ec2Instance, nil

}
