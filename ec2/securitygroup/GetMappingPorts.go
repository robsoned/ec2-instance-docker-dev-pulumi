package securitygroup

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type PortMapping struct {
	Name     string
	Protocol string
	FromPort int
	ToPort   int
}

func GetMappingPorts(ctx *pulumi.Context) []PortMapping {

	var mappings []PortMapping

	cfg := config.New(ctx, "ec2")

	cfg.RequireObject("ingressSecurityGroups", &mappings)

	return mappings
}
