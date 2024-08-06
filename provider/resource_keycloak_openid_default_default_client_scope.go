package provider

import (
	"context"
	"fmt"
	"time"

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

	diagnostics := detachDeletedDefaultScopes(ctx, keycloakOpenidDefaultDefaultScopes, tfOpenidDefaultDefaultScopes, err, keycloakClient, realmId)
	if diagnostics != nil {
		return diagnostics
	}

	if tfOpenidDefaultDefaultScopes.Len() > 0 {
		diagnostics = attachNewDefaultScopes(ctx, keycloakClient, realmId, tfOpenidDefaultDefaultScopes)
		if diagnostics != nil {
			return diagnostics
		}
	}

	return waitForDefaultUpdates(ctx, keycloakClient, realmId, tfOpenidDefaultDefaultScopes, 5)
}

func waitForDefaultUpdates(ctx context.Context, keycloakClient *keycloak.KeycloakClient, realmId string, scopes *schema.Set, times int) diag.Diagnostics {
	if times == 0 {
		return nil
	}
	keycloakOpenidDefaultDefaultScopes, err := keycloakClient.GetOpenidRealmDefaultDefaultClientScopes(ctx, realmId)
	if err != nil {
		if keycloak.ErrorIs404(err) {
			return diag.FromErr(fmt.Errorf("validation error: realm with id %s does not exist", realmId))
		}
		return diag.FromErr(err)
	}

	if len(keycloakOpenidDefaultDefaultScopes) != scopes.Len() {
		time.Sleep(1 * time.Second)
		return waitForOptionalUpdates(ctx, keycloakClient, realmId, scopes, times-1)
	}
	for _, keycloakOpenidDefaultDefaultScope := range keycloakOpenidDefaultDefaultScopes {
		if !scopes.Contains(keycloakOpenidDefaultDefaultScope.Name) {
			time.Sleep(1 * time.Second)
			return waitForOptionalUpdates(ctx, keycloakClient, realmId, scopes, times-1)
		}
	}
	return nil
}

func detachDeletedDefaultScopes(ctx context.Context, keycloakOpenidDefaultDefaultScopes []*keycloak.OpenidClientScope, tfOpenidDefaultDefaultScopes *schema.Set, err error, keycloakClient *keycloak.KeycloakClient, realmId string) diag.Diagnostics {
	for _, keycloakOpenidDefaultDefaultScope := range keycloakOpenidDefaultDefaultScopes {
		if !tfOpenidDefaultDefaultScopes.Contains(keycloakOpenidDefaultDefaultScope.Name) {
			err = keycloakClient.DeleteOpenidRealmDefaultDefaultClientScope(ctx, realmId, keycloakOpenidDefaultDefaultScope.Id)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
	return nil
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

	keycloakOpenidDefaultOptionalScopes, err := keycloakClient.GetOpenidRealmDefaultDefaultClientScopes(ctx, realmId)
	if err != nil {
		if keycloak.ErrorIs404(err) {
			return diag.FromErr(fmt.Errorf("validation error: realm with id %s does not exist", realmId))
		}
		return diag.FromErr(err)
	}

	for _, keycloakClientScope := range keycloakOpenidDefaultOptionalScopes {
		err = keycloakClient.DeleteOpenidRealmDefaultDefaultClientScope(ctx, realmId, keycloakClientScope.Id)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}
