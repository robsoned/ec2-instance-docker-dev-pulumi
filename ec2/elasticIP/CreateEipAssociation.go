package elasticIP

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateEipAssociation(ctx *pulumi.Context, instanceID pulumi.StringOutput, allocationID pulumi.StringInput) error {

	_, err := ec2.NewEipAssociation(ctx, "elasticIPAssociation", &ec2.EipAssociationArgs{
		InstanceId:   instanceID,
		AllocationId: allocationID,
	})

	return err

}
