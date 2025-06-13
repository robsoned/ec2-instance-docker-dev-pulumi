package ec2

import (
	"ec2-instance-docker-dev/ebs"
	"ec2-instance-docker-dev/ec2/elasticIP"
	"ec2-instance-docker-dev/ec2/userdata"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateInstance(ctx *pulumi.Context, volume *ebs.SearchVolumeOutput) (*ec2.Instance, error) {

	elasticIp, err := elasticIP.Create(ctx, &elasticIP.CreateElasticIPArgs{})

	if err != nil {
		return nil, err
	}

	securityGroup, err := getSecurityGroup(ctx, elasticIp.PublicIp)

	if err != nil {
		return nil, err
	}

	ec2InstanceArgs := &ec2.InstanceArgs{
		Ami:                 getAmi(ctx),
		InstanceType:        getInstancetype(ctx),
		KeyName:             getKeyPairName(ctx),
		VpcSecurityGroupIds: pulumi.StringArray{securityGroup.ID()},
		UserData:            userdata.GetInstanceUserData(ctx),
		RootBlockDevice: &ec2.InstanceRootBlockDeviceArgs{
			VolumeSize: pulumi.Int(100),
		},
	}

	if volume != nil {
		ctx.Log.Info("Creating instance in the same availability zone as the volume", nil)
		ec2InstanceArgs.AvailabilityZone = pulumi.String(volume.AvailabilityZone)
	}

	ec2Instance, err := ec2.NewInstance(ctx, "ec2-instance", ec2InstanceArgs)

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
