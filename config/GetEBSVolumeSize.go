package config

import (
	"strconv"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func GetEBSVolumeSize(ctx *pulumi.Context) (pulumi.Int, error) {

	volumeSize := config.New(ctx, "ebs").Require("volumeSize")
	volumeSizeInt, err := strconv.Atoi(volumeSize)

	if err != nil {
		return pulumi.Int(0), err
	}

	return pulumi.Int(volumeSizeInt), nil
}
