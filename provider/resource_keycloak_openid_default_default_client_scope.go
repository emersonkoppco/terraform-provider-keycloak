package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mrparkers/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakOpenidDefaultDefaultClientScopes() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakOpenidDefaultDefaultClientScopeReconcile,
		ReadContext:   resourceKeycloakOpenidDefaultDefaultClientScopesRead,
		DeleteContext: resourceKeycloakOpenidDefaultDefaultClientScopeDelete,
		UpdateContext: resourceKeycloakOpenidDefaultDefaultClientScopeReconcile,
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"default_scopes": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
				Set:      schema.HashString,
			},
		},
	}
}

func resourceKeycloakOpenidDefaultDefaultClientScopesRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)

	clientScopes, err := keycloakClient.GetOpenidRealmDefaultDefaultClientScopes(ctx, realmId)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	var defaultScopes []string
	for _, clientScope := range clientScopes {
		defaultScopes = append(defaultScopes, clientScope.Name)
	}

	err = data.Set("default_scopes", defaultScopes)
	if err != nil {
		return diag.FromErr(err)
	}
	data.SetId(realmId)

	return nil
}

func resourceKeycloakOpenidDefaultDefaultClientScopeReconcile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	tfOpenidDefaultDefaultScopes := data.Get("default_scopes").(*schema.Set)

	keycloakOpenidDefaultDefaultScopes, err := keycloakClient.GetOpenidRealmDefaultDefaultClientScopes(ctx, realmId)
	if err != nil {
		if keycloak.ErrorIs404(err) {
			return diag.FromErr(fmt.Errorf("validation error: realm with id %s does not exist", realmId))
		}
		return diag.FromErr(err)
	}

	diagnostics, done := detachDeletedDefaultScopes(ctx, keycloakOpenidDefaultDefaultScopes, tfOpenidDefaultDefaultScopes, err, keycloakClient, realmId)
	if done {
		return diagnostics
	}

	if tfOpenidDefaultDefaultScopes.Len() > 0 {
		return attachNewDefaultScopes(ctx, keycloakClient, realmId, tfOpenidDefaultDefaultScopes)
	}

	return nil
}

func detachDeletedDefaultScopes(ctx context.Context, keycloakOpenidDefaultDefaultScopes []*keycloak.OpenidClientScope, tfOpenidDefaultDefaultScopes *schema.Set, err error, keycloakClient *keycloak.KeycloakClient, realmId string) (diag.Diagnostics, bool) {
	for _, keycloakOpenidDefaultDefaultScope := range keycloakOpenidDefaultDefaultScopes {
		if tfOpenidDefaultDefaultScopes.Contains(keycloakOpenidDefaultDefaultScope.Name) {
			tfOpenidDefaultDefaultScopes.Remove(keycloakOpenidDefaultDefaultScope.Name)
		} else {
			err = keycloakClient.DeleteOpenidRealmDefaultDefaultClientScope(ctx, realmId, keycloakOpenidDefaultDefaultScope.Id)
			if err != nil {
				return diag.FromErr(err), true
			}
		}
	}
	return nil, false
}

func attachNewDefaultScopes(ctx context.Context, keycloakClient *keycloak.KeycloakClient, realmId string, tfOpenidDefaultDefaultScopes *schema.Set) diag.Diagnostics {
	keycloakClientScopes, err := keycloakClient.GetRealmClientScopes(ctx, realmId)
	if err != nil {
		return diag.FromErr(err)
	}
	for _, keycloakClientScope := range keycloakClientScopes {
		if tfOpenidDefaultDefaultScopes.Contains(keycloakClientScope.Name) {
			err = keycloakClient.PutOpenidRealmDefaultDefaultClientScope(ctx, realmId, keycloakClientScope.Id)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
	return nil
}

func resourceKeycloakOpenidDefaultDefaultClientScopeDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientScopeId := data.Get("client_scope_id").(string)

	return diag.FromErr(keycloakClient.DeleteOpenidRealmDefaultDefaultClientScope(ctx, realmId, clientScopeId))
}
