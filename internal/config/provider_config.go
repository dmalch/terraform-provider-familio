package config

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/dmalch/terraform-provider-familio/internal/familio"
)

// FamilioProviderConfig is the decoded provider block.
type FamilioProviderConfig struct {
	Cookie       types.String `tfsdk:"cookie"`
	SessionToken types.String `tfsdk:"session_token"`
	Browser      types.String `tfsdk:"browser"`
}

// ClientData is handed to every resource/data source via ProviderData.
type ClientData struct {
	Client *familio.Client
}
