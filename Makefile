-include .env

AZURE_SERVICE_PRINCIPAL ?= pulumi-azure

AZURE_CLIENT_ID ?=
AZURE_CLIENT_SECRET ?=
AZURE_TENANT_ID ?=
AZURE_SUBSCRIPTION_ID ?=

service-principal:
	az ad sp create-for-rbac --name "$(AZURE_SERVICE_PRINCIPAL)"

pulumi-config:
	pulumi config set azure:clientId "$(AZURE_CLIENT_ID)"
	pulumi config set azure:clientSecret "$(AZURE_CLIENT_SECRET)" --secret
	pulumi config set azure:tenantId "$(AZURE_TENANT_ID)"
	pulumi config set azure:subscriptionId "$(AZURE_SUBSCRIPTION_ID)"
