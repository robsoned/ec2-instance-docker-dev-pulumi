package ec2

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func getInstancetype(ctx *pulumi.Context) pulumi.String {
	instanceType := config.New(ctx, "ec2").Require("instanceType")
	return pulumi.String(instanceType)
}
