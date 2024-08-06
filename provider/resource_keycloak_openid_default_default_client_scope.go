package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mrparkers/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakOpenidDefaultDefaultClientScope() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakOpenidDefaultDefaultClientScopeCreate,
		ReadContext:   resourceKeycloakOpenidDefaultDefaultClientScopesRead,
		DeleteContext: resourceKeycloakOpenidDefaultDefaultClientScopeDelete,
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"client_scope_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"client_scope_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceKeycloakOpenidDefaultDefaultClientScopeCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientScopeId := data.Get("client_scope_id").(string)

	return diag.FromErr(keycloakClient.PutOpenidRealmDefaultDefaultClientScope(ctx, realmId, clientScopeId))
}

func resourceKeycloakOpenidDefaultDefaultClientScopesRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientScopeId := data.Get("client_scope_id").(string)

	clientScope, err := keycloakClient.GetOpenidRealmDefaultDefaultClientScope(ctx, realmId, clientScopeId)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	err = data.Set("client_scope_id", clientScope.Id)
	if err != nil {
		return diag.FromErr(err)
	}
	err = data.Set("client_scope_name", clientScope.Name)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakOpenidDefaultDefaultClientScopeDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientScopeId := data.Get("client_scope_id").(string)

	return diag.FromErr(keycloakClient.DeleteOpenidRealmDefaultDefaultClientScope(ctx, realmId, clientScopeId))
}
