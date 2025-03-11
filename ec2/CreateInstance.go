package ec2

import (
	"ec2-instance-docker-dev/ec2/elasticIP"
	"ec2-instance-docker-dev/ec2/userdata"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateInstance(ctx *pulumi.Context) error {

	elasticIp, err := elasticIP.Create(ctx, &elasticIP.CreateElasticIPArgs{})

	if err != nil {
		return err
	}

	securityGroup, err := getSecurityGroup(ctx, elasticIp.PublicIp)

	if err != nil {
		return err
	}

	ec2Instance, err := ec2.NewInstance(ctx, "ec2-instance", &ec2.InstanceArgs{
		Ami:                 getAmi(ctx),
		InstanceType:        getInstancetype(ctx),
		KeyName:             getKeyPairName(ctx),
		VpcSecurityGroupIds: pulumi.StringArray{securityGroup.ID()},
		UserData:            userdata.GetInstanceUserData(ctx),
		RootBlockDevice: &ec2.InstanceRootBlockDeviceArgs{
			VolumeSize: pulumi.Int(100),
		},
	})

	if err != nil {
		return err
	}

	err = elasticIP.CreateEipAssociation(ctx, ec2Instance.ID().ToStringOutput(), elasticIp.AllocationId)

	if err != nil {
		return err
	}

	ctx.Export("elastic-ip", elasticIp.PublicIp)

	return nil

}
