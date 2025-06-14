package ebs

import (
	"ec2-instance-docker-dev/config"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ebs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const pulumiProjectTag = "PulumiProject"
const pulumiStackTag = "PulumiStack"
const pulumiEBSName = "ec2-instane-ebs-volume"

// todo, return idString when already exists, return ID when creted
type CreateVolumeResult struct {
	VolumeIDString *string
	VolumeIDOutput *pulumi.IDOutput
}

func CreateVolume(ctx *pulumi.Context, availabilityZone pulumi.StringInput) (*ebs.Volume, error) {

	volumeSize, err := config.GetEBSVolumeSize(ctx)

	if err != nil {
		return nil, err
	}

	volume, err := ebs.NewVolume(ctx, pulumiEBSName, &ebs.VolumeArgs{
		AvailabilityZone: availabilityZone,
		Size:             volumeSize,
		Type:             pulumi.String("gp3"),
		Tags: pulumi.StringMap{
			"Name":           pulumi.String(pulumiEBSName),
			pulumiProjectTag: pulumi.String(ctx.Project()),
			pulumiStackTag:   pulumi.String(ctx.Stack()),
		},
	}, pulumi.RetainOnDelete(true))

	if err != nil {
		return nil, err
	}

	return volume, nil
}
