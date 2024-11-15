package ec2

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func getKeyPairName(ctx *pulumi.Context) pulumi.String {

	keyPair := config.New(ctx, "ec2").Require("keyPairName")

	return pulumi.String(keyPair)

}
