package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/hashicorp/terraform/terraform"
	"github.com/brapbrapbrap/terraform-provisioner-psinvoke/psinvoke"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProvisionerFunc: func() terraform.ResourceProvisioner {
			return new(psinvoke.ResourceProvisioner)
		},
	})
}
