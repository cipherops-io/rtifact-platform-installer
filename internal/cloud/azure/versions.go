package azure

// Pinned versions and naming constants for the Azure provider.
const (
	CloudName = "azure"

	OpenTofuRequiredVersion = ">= 1.6.0"

	ProviderSource         = "opentofu/azurerm"
	ProviderVersion        = "5.14.0"
	AzureRMProviderVersion = ProviderVersion
	RandomProviderVersion  = "3.9.1"
	TLSProviderVersion     = "4.4.2"

	AzureCLI        = "az"
	CLIVersion      = "2.93.0"
	CLIName         = AzureCLI
	StateContainerName = "tfstate"

	StateStorageAccountSuffix = "tfstate"

	StorageAccountNameMaxLen = 24
)