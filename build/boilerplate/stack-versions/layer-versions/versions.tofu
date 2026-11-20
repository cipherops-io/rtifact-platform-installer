# GENERATED from vars/versions.yml — edit there, then: make bump-providers
terraform {
  required_version = ">= {{ .OpenTofuVersion }}"
  required_providers {
    {{- $cloud := index (splitList "/" .StackTarget) 0 }}
    {{- $src := index .ProviderSources $cloud }}
    {{- $short := index (splitList "/" $src) 1 }}
    {{ $short }} = {
      source  = "{{ $src }}"
      version = "{{ index .ProviderVersions $cloud }}"
    }
    random = {
      source  = "opentofu/random"
      version = "{{ index .ExtraProviderVersions "random" }}"
    }
    tls = {
      source  = "opentofu/tls"
      version = "{{ index .ExtraProviderVersions "tls" }}"
    }
  }
}
