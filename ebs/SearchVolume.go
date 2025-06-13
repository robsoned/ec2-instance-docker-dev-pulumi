package ebs

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ebs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type SearchVolumeOutput struct {
	ID               string
	AvailabilityZone string
}

func SearchVolume(ctx *pulumi.Context) (*SearchVolumeOutput, error) {

	ctx.Log.Info("Searching for existing volume", nil)
	volumesSearchResult, err := ebs.GetEbsVolumes(ctx, &ebs.GetEbsVolumesArgs{
		Tags: map[string]string{
			pulumiProjectTag: ctx.Project(),
			pulumiStackTag:   ctx.Stack(),
			"Name":           pulumiEBSName,
		},
	})

	if err != nil {
		ctx.Log.Error("Error searching for volume", nil)
		return nil, err
	}

	if len(volumesSearchResult.Ids) == 0 {
		ctx.Log.Info("Volume not found", nil)
		return nil, nil
	}

	ctx.Log.Info("Volume found, getting the volume resource", nil)

	volume, err := ebs.LookupVolume(ctx, &ebs.LookupVolumeArgs{
		Filters: []ebs.GetVolumeFilter{
			{
				Name:   "tag:" + pulumiProjectTag,
				Values: []string{ctx.Project()},
			},
			{
				Name:   "tag:" + pulumiStackTag,
				Values: []string{ctx.Stack()},
			},
			{
				Name:   "tag:Name",
				Values: []string{pulumiEBSName},
			},
		},
		Tags: map[string]string{
			pulumiProjectTag: ctx.Project(),
			pulumiStackTag:   ctx.Stack(),
			"Name":           pulumiEBSName,
		},
	})

	if err != nil {
		ctx.Log.Error("Error getting volume resource", nil)
		return nil, err
	}

	ctx.Log.Info("Volume found", nil)

	return &SearchVolumeOutput{
		ID:               volume.Id,
		AvailabilityZone: volume.AvailabilityZone,
	}, nil

}
