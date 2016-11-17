# terraform-provisioner-psinvoke
> Provide terraform resources with ways of remotely running commands with Powershell
> Useful when domain authentication is required

## Overview

**[Terraform](https://github.com/hashicorp/terraform)** is a tool for automating infrastructure. Terraform includes the ability to provision resources at creation time through a plugin api. Currently, some builtin [provisioners](https://www.terraform.io/docs/provisioners/) such as **chef** and standard scripts are provided; this provisioner introduces the ability to run remote commands through **powershell**.

This provisioner provides the ability to run a command remotely on a host at provision time. The command on the remote machine can be a Powershell script, or anything else. It accepts parameters to be passed to the remote command.

**terraform-provisioner-psinvoke** is shipped as a **Terraform** [module](https://www.terraform.io/docs/modules/create.html). To include it, simply download the binary and enable it as a terraform plugin in your **terraformrc**.

## Why not just use remote-exec?

The provisioner remote-exec works just fine in the vast majority of cases. WinRM is a great, cross platform way to communicate remotely with Windows boxes. There are a few scenarios where it falls down:

* If WinRM is filtered by firewalls or disabled by group policy.
* If domain authentication is required (the WinRM implementation used by Terraform does not support domain authentication).

Psinvoke supports domain authentication and creates a temporary Powershell script that runs the remote commands as **Invoke-Command** with Powershell. If you are affected by either of the scenarios above psinvoke might work for you.

## Installation

**terraform-provisioner-psinvoke** ships as a single binary and is compatible with **terraform**'s plugin interface. Behind the scenes, terraform plugins use https://github.com/hashicorp/go-plugin and communicate with the parent terraform process via RPC.

To install, download and un-archive the binary and place it on your path.

Once installed, an `%APPDATA%\terraform.rc` file is used to _enable_ the plugin.

```bash
providers {
    psinvoke = "C:\\terraform-provisioner-psinvoke.exe"
}
```

## Usage

Once installed, you can provision resources by including a `psinvoke` provisioner block.

```
{
  resource "aws_instance" "terraform-provisioner-psinvoke-example" {
    ami = "ami-408c7f28"
    instance_type = "t1.micro"

    provisioner "psinvoke" {
      host = "server"
      username = "domain\user"
      password = "pass"
      command = "d:\run.ps1"
      params = "-t"
    }
  }
}
```

