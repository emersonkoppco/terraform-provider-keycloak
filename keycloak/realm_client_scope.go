package keycloak

import (
	"context"
	"fmt"

	"github.com/mrparkers/terraform-provider-keycloak/keycloak/types"
)

type ClientScope struct {
	Id          string `json:"id,omitempty"`
	RealmId     string `json:"-"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Protocol    string `json:"protocol"`
	Attributes  struct {
		DisplayOnConsentScreen types.KeycloakBoolQuoted `json:"display.on.consent.screen"` // boolean in string form
		ConsentScreenText      string                   `json:"consent.screen.text"`
		GuiOrder               string                   `json:"gui.order"`
		IncludeInTokenScope    types.KeycloakBoolQuoted `json:"include.in.token.scope"` // boolean in string form
	} `json:"attributes"`
}

func (keycloakClient *KeycloakClient) GetRealmClientScopes(ctx context.Context, realmId string) ([]*ClientScope, error) {
	var clientScopes []*ClientScope

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/client-scopes", realmId), &clientScopes, nil)
	if err != nil {
		return nil, err
	}

	for _, clientScope := range clientScopes {
		clientScope.RealmId = realmId
	}

	return clientScopes, nil
}
