package ec2

import (
	"ec2-instance-docker-dev/ec2/userdata"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateInstance(ctx *pulumi.Context) error {

	securityGroup, err := getSecurityGroup(ctx)

	if err != nil {
		return err
	}

	ec2Instance, err := ec2.NewInstance(ctx, "ec2-instance", &ec2.InstanceArgs{
		Ami:                 getAmi(ctx),
		InstanceType:        getInstancetype(ctx),
		KeyName:             getKeyPairName(ctx),
		VpcSecurityGroupIds: pulumi.StringArray{securityGroup.ID()},
		UserData:            userdata.GetInstanceUserData(ctx),
	})

	if err != nil {
		return err
	}

	ctx.Export("intance-ip", ec2Instance.PublicIp)

	return nil

}
