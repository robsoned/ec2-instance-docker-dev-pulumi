package ec2

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func getAmi(ctx *pulumi.Context) pulumi.String {
	ami := config.New(ctx, "ec2").Require("ami")
	return pulumi.String(ami)
}
