package common

type Config struct {
	AzureTenantID     string `json:"azure.tenantId" validate:"required"`
	AzureClientID     string `json:"azure.clientId" validate:"required"`
	AzureClientSecret string `json:"azure.clientSecret" validate:"required"`

	EventHubNameSpace string `json:"eventHubNamespace" validate:"required"`
	EventHubName      string `json:"eventHubName" validate:"required"`
}
