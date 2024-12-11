package ec2

import (
	"ec2-instance-docker-dev/ec2/securitygroup"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/vpc"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func getSecurityGroup(ctx *pulumi.Context) (*ec2.SecurityGroup, error) {

	securityGroupCidrIpv4 := getSecurityGroupCidrIpv4(ctx)

	securityGroup, err := ec2.NewSecurityGroup(ctx, "securityGroup", &ec2.SecurityGroupArgs{
		Tags: pulumi.StringMap{
			"Name": pulumi.String("ec2-instance-docker-dev"),
		},
	})

	if err != nil {
		return nil, err
	}

	err = createSecurityGroupIngresses(ctx, securityGroup.ID(), securityGroupCidrIpv4)

	if err != nil {
		return nil, err
	}

	_, err = vpc.NewSecurityGroupEgressRule(
		ctx,
		"securityGroupEgressRule",
		&vpc.SecurityGroupEgressRuleArgs{
			SecurityGroupId: securityGroup.ID(),
			CidrIpv4:        pulumi.String("0.0.0.0/0"),
			FromPort:        pulumi.Int(0),
			ToPort:          pulumi.Int(0),
			IpProtocol:      pulumi.String("-1"),
			Description:     pulumi.String("Allow all traffic"),
		})

	if err != nil {
		return nil, err
	}

	return securityGroup, nil

}

func createSecurityGroupIngresses(ctx *pulumi.Context, securityGroupId pulumi.StringInput, cidrIpv4 pulumi.StringInput) error {

	mappingPorts := securitygroup.GetMappingPorts(ctx)

	for _, mapping := range mappingPorts {

		ingressRuleName := mapping.Name + "-IngressRule"

		_, err := vpc.NewSecurityGroupIngressRule(ctx, ingressRuleName, &vpc.SecurityGroupIngressRuleArgs{

			SecurityGroupId: securityGroupId,
			CidrIpv4:        cidrIpv4,
			FromPort:        pulumi.Int(mapping.FromPort),
			ToPort:          pulumi.Int(mapping.ToPort),
			IpProtocol:      pulumi.String(mapping.Protocol),
			Tags: pulumi.StringMap{
				"Name": pulumi.String(ingressRuleName),
			},
		})

		if err != nil {
			return err
		}

	}

	return nil

}
