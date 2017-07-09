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
      script = "d:\run.ps1"
      params = "-t"
    }
  }
}
```

## Build Gotchas

When building, make sure to build against a specific version of Terraform, roughly matching the version you are currently using. Building against the latest commit can produce issues with the API versions. To avoid this, pull down a specific tag in git, for example to build against v0.8.4:

```
> cd C:\go\src\github.com\hashicorp\terraform
> git ls-remote --tags
...
# b09283b76c6371340659bebfea72e617cd13ad4e        refs/tags/v0.8.3
# b845cb7093c24c49ece854bb38b55aff3783f195        refs/tags/v0.8.3^{}
# 1737cfff270d479e31533e19609442c40344872d        refs/tags/v0.8.4
# a791ff09b29d063dd4b6da0cac04ad3b83c836f5        refs/tags/v0.8.4^{}
# 7ef08156c201953f758ab8bb9da2471d37c07f5f        refs/tags/v0.8.5
# b4d477660b5abd20f2a70175460c9603797fada0        refs/tags/v0.8.5^{}
# aa0a7fa2f6216fe12cdb2acd672fa5d9b5b92a12        refs/tags/v0.8.6
# df4bcf64828598a25bd41f00470b2ab3a66f3169        refs/tags/v0.8.6^{}

> git checkout a791ff09b29d063dd4b6da0cac04ad3b83c836f5
> cd C:\go\src\github.com\brapbrapbrap\terraform-provisioner-psinvoke
> go build
```
