package common

type Config struct {
	// AzureTenantID is the azure tenant ID
	AzureTenantID string `json:"azure.tenantId" validate:"required"`
	// AzureClientID is your azure client ID
	AzureClientID string `json:"azure.clientId" validate:"required"`
	// AzureClientSecret is your azure client secret
	AzureClientSecret string `json:"azure.clientSecret" validate:"required"`

	// EventHubNameSpace is your FQNS
	EventHubNameSpace string `json:"eventHubNamespace" validate:"required"`
	// EventHubName is your event hub name
	EventHubName string `json:"eventHubName" validate:"required"`
}
