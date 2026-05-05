package iam

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

const ec2TrustPolicy = `{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": { "Service": "ec2.amazonaws.com" },
    "Action": "sts:AssumeRole"
  }]
}`

var nonAlphanumeric = regexp.MustCompile(`[^a-zA-Z0-9]`)

type CreateRoleOutput struct {
	InstanceProfileName pulumi.StringOutput
}

func CreateRole(ctx *pulumi.Context) (*CreateRoleOutput, error) {
	cfg := config.New(ctx, "ec2")
	raw := cfg.Get("iamPolicies")
	if raw == "" {
		return nil, nil
	}

	var policyNames []string
	if err := json.Unmarshal([]byte(raw), &policyNames); err != nil {
		return nil, fmt.Errorf("ec2:iamPolicies must be a JSON array of policy names: %w", err)
	}

	role, err := iam.NewRole(ctx, "ec2-dev-role", &iam.RoleArgs{
		AssumeRolePolicy: pulumi.String(ec2TrustPolicy),
		Tags:             pulumi.StringMap{"Name": pulumi.String("ec2-dev-role")},
	})
	if err != nil {
		return nil, err
	}

	instanceProfile, err := iam.NewInstanceProfile(ctx, "ec2-dev-instance-profile", &iam.InstanceProfileArgs{
		Role: role.Name,
		Tags: pulumi.StringMap{"Name": pulumi.String("ec2-dev-instance-profile")},
	})
	if err != nil {
		return nil, err
	}

	for _, name := range policyNames {
		arn := fmt.Sprintf("arn:aws:iam::aws:policy/%s", name)
		resourceName := "ec2-dev-policy-" + nonAlphanumeric.ReplaceAllString(name, "-")
		_, err := iam.NewRolePolicyAttachment(ctx, resourceName, &iam.RolePolicyAttachmentArgs{
			Role:      role.Name,
			PolicyArn: pulumi.String(arn),
		})
		if err != nil {
			return nil, err
		}
	}

	return &CreateRoleOutput{InstanceProfileName: instanceProfile.Name}, nil
}
