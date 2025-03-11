package elasticIP

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type CreateElasticIPArgs struct {
	InstanceId pulumi.StringInput
}

func Create(ctx *pulumi.Context, args *CreateElasticIPArgs) (*ec2.Eip, error) {

	elasticIP, err := ec2.NewEip(ctx, "elasticIP", &ec2.EipArgs{

		Tags: pulumi.StringMap{
			"Name": pulumi.String("elasticIP"),
		},
	})

	if err != nil {
		return nil, err
	}

	return elasticIP, nil

}
