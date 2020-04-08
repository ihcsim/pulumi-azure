# Pulumi Azure

A Pulumi program to provision Azure resources.

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
