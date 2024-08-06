package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mrparkers/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakOpenidDefaultOptionalClientScopes() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakOpenidDefaultOptionalClientScopeReconcile,
		ReadContext:   resourceKeycloakOpenidDefaultOptionalClientScopesRead,
		UpdateContext: resourceKeycloakOpenidDefaultOptionalClientScopeReconcile,
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"optional_scopes": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
				Set:      schema.HashString,
			},
		},
	}
}

func resourceKeycloakOpenidDefaultOptionalClientScopesRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)

	clientScopes, err := keycloakClient.GetOpenidRealmDefaultOptionalClientScopes(ctx, realmId)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	var defaultScopes []string
	for _, clientScope := range clientScopes {
		defaultScopes = append(defaultScopes, clientScope.Name)
	}

	err = data.Set("optional_scopes", defaultScopes)
	if err != nil {
		return diag.FromErr(err)
	}
	data.SetId(realmId)

	return nil
}

func resourceKeycloakOpenidDefaultOptionalClientScopeReconcile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	tfOpenidDefaultOptionalScopes := data.Get("optional_scopes").(*schema.Set)

	keycloakOpenidDefaultOptionalScopes, err := keycloakClient.GetOpenidRealmDefaultOptionalClientScopes(ctx, realmId)
	if err != nil {
		if keycloak.ErrorIs404(err) {
			return diag.FromErr(fmt.Errorf("validation error: realm with id %s does not exist", realmId))
		}
		return diag.FromErr(err)
	}

	diagnostics, done := detachDeletedOptionalScopes(ctx, keycloakOpenidDefaultOptionalScopes, tfOpenidDefaultOptionalScopes, err, keycloakClient, realmId)
	if done {
		return diagnostics
	}

	if tfOpenidDefaultOptionalScopes.Len() > 0 {
		return attachNewOptionalScopes(ctx, keycloakClient, realmId, tfOpenidDefaultOptionalScopes)
	}

	return nil
}

func detachDeletedOptionalScopes(ctx context.Context, keycloakOpenidDefaultOptionalScopes []*keycloak.OpenidClientScope, tfOpenidDefaultOptionalScopes *schema.Set, err error, keycloakClient *keycloak.KeycloakClient, realmId string) (diag.Diagnostics, bool) {
	for _, keycloakOpenidDefaultOptionalScope := range keycloakOpenidDefaultOptionalScopes {
		if tfOpenidDefaultOptionalScopes.Contains(keycloakOpenidDefaultOptionalScope.Name) {
			tfOpenidDefaultOptionalScopes.Remove(keycloakOpenidDefaultOptionalScope.Name)
		} else {
			err = keycloakClient.DeleteOpenidRealmDefaultOptionalClientScope(ctx, realmId, keycloakOpenidDefaultOptionalScope.Id)
			if err != nil {
				return diag.FromErr(err), true
			}
		}
	}
	return nil, false
}

func attachNewOptionalScopes(ctx context.Context, keycloakClient *keycloak.KeycloakClient, realmId string, tfOpenidDefaultOptionalScopes *schema.Set) diag.Diagnostics {
	keycloakClientScopes, err := keycloakClient.GetRealmClientScopes(ctx, realmId)
	if err != nil {
		return diag.FromErr(err)
	}
	for _, keycloakClientScope := range keycloakClientScopes {
		if tfOpenidDefaultOptionalScopes.Contains(keycloakClientScope.Name) {
			err = keycloakClient.PutOpenidRealmDefaultOptionalClientScope(ctx, realmId, keycloakClientScope.Id)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
	return nil
}
