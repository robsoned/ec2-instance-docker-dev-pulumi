package config

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func GetAvailabilityZone(ctx *pulumi.Context) pulumi.String {
	availabilityZone := config.New(ctx, "aws").Require("region")
	return pulumi.String(availabilityZone)
}
