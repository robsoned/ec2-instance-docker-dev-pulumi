package ec2

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/vpc"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func getSecurityGroup(ctx *pulumi.Context) (*ec2.SecurityGroup, error) {

	getSecurityGroupCidrIpv4 := getSecurityGroupCidrIpv4(ctx)

	securityGroup, err := ec2.NewSecurityGroup(ctx, "securityGroup", &ec2.SecurityGroupArgs{
		Tags: pulumi.StringMap{
			"Name": pulumi.String("ec2-instance-docker-dev"),
		},
	})

	if err != nil {
		return nil, err
	}

	_, err = vpc.NewSecurityGroupIngressRule(ctx, "securityGroupSSHIngressRule", &vpc.SecurityGroupIngressRuleArgs{
		SecurityGroupId: securityGroup.ID(),
		CidrIpv4:        getSecurityGroupCidrIpv4,
		FromPort:        pulumi.Int(22),
		ToPort:          pulumi.Int(22),
		IpProtocol:      pulumi.String("tcp"),
		Description:     pulumi.String("Allow SSH access"),
	})

	if err != nil {
		return nil, err
	}

	_, err = vpc.NewSecurityGroupIngressRule(ctx, "securityGroupHTTPIngressRule", &vpc.SecurityGroupIngressRuleArgs{
		SecurityGroupId: securityGroup.ID(),
		CidrIpv4:        getSecurityGroupCidrIpv4,
		FromPort:        pulumi.Int(80),
		ToPort:          pulumi.Int(80),
		IpProtocol:      pulumi.String("tcp"),
		Description:     pulumi.String("Allow HTTP access"),
	})

	if err != nil {
		return nil, err
	}

	_, err = vpc.NewSecurityGroupIngressRule(ctx, "securityGroupHTTPSIngressRule3", &vpc.SecurityGroupIngressRuleArgs{
		SecurityGroupId: securityGroup.ID(),
		CidrIpv4:        getSecurityGroupCidrIpv4,
		FromPort:        pulumi.Int(443),
		ToPort:          pulumi.Int(443),
		IpProtocol:      pulumi.String("tcp"),
		Description:     pulumi.String("Allow HTTPS access"),
	})

	if err != nil {
		return nil, err
	}

	_, err = vpc.NewSecurityGroupIngressRule(ctx, "securityGroupPHPMYAdminIngressRule", &vpc.SecurityGroupIngressRuleArgs{
		SecurityGroupId: securityGroup.ID(),
		CidrIpv4:        getSecurityGroupCidrIpv4,
		FromPort:        pulumi.Int(9000),
		ToPort:          pulumi.Int(9000),
		IpProtocol:      pulumi.String("tcp"),
		Description:     pulumi.String("Allow PHPMYAdmin access"),
	})

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
