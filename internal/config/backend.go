package config

import "fmt"

// EffectiveBackend returns remote backend config for a stack layer.
func EffectiveBackend(customer *CustomerConfig, profileID string, layer string) *BackendBlock {
	if customer.State != nil && customer.State.Backend == "local" {
		return nil
	}
	if customer.Backend != nil && customer.Backend.Key != "" {
		b := *customer.Backend
		if b.Key == "" {
			b.Key = fmt.Sprintf("%s/%s/%s.tfstate", profileID, customer.NamePrefix(), layer)
		}
		return &b
	}
	if customer.State == nil || customer.State.Backend != "azurerm" {
		return nil
	}
	return &BackendBlock{
		Type:          "azurerm",
		ContainerName: "tfstate",
		Key:           fmt.Sprintf("%s/%s/%s.tfstate", profileID, customer.NamePrefix(), layer),
	}
}