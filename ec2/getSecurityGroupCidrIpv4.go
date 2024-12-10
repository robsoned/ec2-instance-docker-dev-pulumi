package ec2

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func getSecurityGroupCidrIpv4(ctx *pulumi.Context) pulumi.String {

	securityGroupCidrIpv4 := config.New(ctx, "ec2").Require("securityGroupCidrIpv4")

	return pulumi.String(securityGroupCidrIpv4)
}
