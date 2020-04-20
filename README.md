# Pulumi Azure

This project is a Pulumi program that knows how to provision the following Azure resources:

* A new virtual network with 3 subnets
* An application security group
* A network security group with 2 rules to allow HTTP and SSH accesses
* 3 `frontend` Ubuntu VMs deployed to different subnets
  * All VMs belong to the same availability set
  * Each VM has a 30GB OS disk attached to it
* 3 `backend` Ubuntu VMs deployed to different subnets
  * All VMs belong to the same availability set
  * Each VM has a 30GB OS disk attached to it

## Prerequisites

The following is a list of required software:

1. Pulumi v1.14.0

## Getting Started

To create a new
[service principal](https://docs.microsoft.com/en-us/azure/active-directory/develop/app-objects-and-service-principals)
for Pulumi:

```
AZURE_SERVICE_PRINCIPAL=<name> make service-principal
```

To configure Pulumi to target your Azure account:

```
AZURE_CLIENT_ID=<client_id> AZURE_CLIENT_SECRET=<client_secret> AZURE_TENANT_ID=<tenant_id> AZURE_SUBSCRIPTION_ID=<subscription_id> make pulumi-config
```

See the Pulumi
[doc](https://www.pulumi.com/docs/intro/cloud-providers/azure/setup/#service-principal-authentication)
for information on how to use a service principal to connect Pulumi to Azure.

To run the Pulumi program:

```
pulumi up
```
