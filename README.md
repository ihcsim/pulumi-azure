# Pulumi Azure

This project is a Pulumi program that knows how to provision the following Azure resources:

* A new Azure resource group
* Two
[availability sets](https://docs.microsoft.com/en-us/azure/virtual-machines/windows/tutorial-availability-sets)
; `frontend` or `backend`
* Two
[application security groups](https://docs.microsoft.com/en-us/azure/virtual-network/application-security-groups)
; `web-servers` or `admin-servers`
* A new virtual network with 3 subnets
* 3 `frontend` and 3 `backend` Ubuntu VMs with the following properties:
  * Deployed to private subnets
  * Grouped by availability sets
  * Assigned to application security groups
  * 30GB OS disk
* A network security group with rules to allow HTTP/HTTPS and SSH access to the
VMs
* Two application security groups that are used as the `Destination` of  the
network security rules
* A bastion host that allows external access to the private VMs

## Prerequisites

The following is a list of required software:

1. Pulumi v1.14.0
1. az-client 2.3.1

## Getting Started

Create a new
[service principal](https://docs.microsoft.com/en-us/azure/active-directory/develop/app-objects-and-service-principals)
for Pulumi:

```
AZURE_SERVICE_PRINCIPAL=<name> make service-principal
```

Configure Pulumi to target your Azure account:

```
AZURE_CLIENT_ID=<client_id> \
AZURE_CLIENT_SECRET=<client_secret> \
AZURE_TENANT_ID=<tenant_id> \
AZURE_SUBSCRIPTION_ID=<subscription_id> \
make pulumi-config
```

See the Pulumi
[doc](https://www.pulumi.com/docs/intro/cloud-providers/azure/setup/#service-principal-authentication)
for information on how to use a service principal to connect Pulumi to Azure.

To get a preview of what this Pulumi program does:

```
pulumi preview
```

To run this Pulumi program:

```
pulumi up
```
