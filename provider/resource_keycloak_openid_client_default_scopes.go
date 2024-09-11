package provider

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mrparkers/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakOpenidClientDefaultScopes() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakOpenidClientDefaultScopesReconcile,
		ReadContext:   resourceKeycloakOpenidClientDefaultScopesRead,
		DeleteContext: resourceKeycloakOpenidClientDefaultScopesDelete,
		UpdateContext: resourceKeycloakOpenidClientDefaultScopesReconcile,
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakOpenidClientDefaultScopesImport,
		},
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"client_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"default_scopes": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
			},
		},
	}
}

func openidClientDefaultScopesId(realmId string, clientId string) string {
	return fmt.Sprintf("%s/%s", realmId, clientId)
}

func resourceKeycloakOpenidClientDefaultScopesRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)

	clientScopes, err := keycloakClient.GetOpenidClientDefaultScopes(ctx, realmId, clientId)
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
	data.SetId(openidClientDefaultScopesId(realmId, clientId))

	return nil
}

func resourceKeycloakOpenidClientDefaultScopesReconcile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)
	tfOpenidClientDefaultScopes := interfaceSliceToStringSlice(data.Get("default_scopes").([]any))

	keycloakOpenidClientDefaultScopes, err := keycloakClient.GetOpenidClientDefaultScopes(ctx, realmId, clientId)
	if err != nil {
		if keycloak.ErrorIs404(err) {
			return diag.FromErr(fmt.Errorf("validation error: client with id %s does not exist", clientId))
		}
		return diag.FromErr(err)
	}

	var openidClientDefaultScopesToDetach []string
	for _, keycloakOpenidClientDefaultScope := range keycloakOpenidClientDefaultScopes {
		// if this scope is attached in keycloak and tf state, no update is required
		// remove it from the set so we can look at scopes that need to be attached later
		if slices.Contains(tfOpenidClientDefaultScopes, keycloakOpenidClientDefaultScope.Name) {
			tfOpenidClientDefaultScopes = slices.DeleteFunc(tfOpenidClientDefaultScopes, func(e string) bool {
				return e == keycloakOpenidClientDefaultScope.Name
			})
		} else {
			// if this scope is attached in keycloak but not in tf state, add them to a slice containing all scopes to detach
			openidClientDefaultScopesToDetach = append(openidClientDefaultScopesToDetach, keycloakOpenidClientDefaultScope.Name)
		}
	}

	// detach scopes that aren't in tf state
	err = keycloakClient.DetachOpenidClientDefaultScopes(ctx, realmId, clientId, openidClientDefaultScopesToDetach)
	if err != nil {
		return diag.FromErr(err)
	}

	// attach scopes that exist in tf state but not in keycloak
	err = keycloakClient.AttachOpenidClientDefaultScopes(ctx, realmId, clientId, tfOpenidClientDefaultScopes)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(openidClientDefaultScopesId(realmId, clientId))

	return resourceKeycloakOpenidClientDefaultScopesRead(ctx, data, meta)
}

func resourceKeycloakOpenidClientDefaultScopesDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)
	defaultScopes := interfaceSliceToStringSlice(data.Get("default_scopes").([]any))

	return diag.FromErr(keycloakClient.DetachOpenidClientDefaultScopes(ctx, realmId, clientId, defaultScopes))
}

func resourceKeycloakOpenidClientDefaultScopesImport(ctx context.Context, data *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(data.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("Invalid import. Supported import formats: {{realmId}}/{{openidClientId}}")
	}
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := parts[0]
	clientId := parts[1]

	keycloakOpenidClientDefaultScopes, err := keycloakClient.GetOpenidClientDefaultScopes(ctx, realmId, clientId)
	if err != nil {
		return nil, err
	}

	err = data.Set("realm_id", realmId)
	if err != nil {
		return nil, err
	}
	err = data.Set("client_id", clientId)
	if err != nil {
		return nil, err
	}
	var defaultScopes []string
	for _, clientScope := range keycloakOpenidClientDefaultScopes {
		defaultScopes = append(defaultScopes, clientScope.Name)
	}
	err = data.Set("default_scopes", defaultScopes)
	if err != nil {
		return nil, err
	}

	diagnostics := resourceKeycloakOpenidClientDefaultScopesRead(ctx, data, meta)
	if diagnostics.HasError() {
		return nil, errors.New(diagnostics[0].Summary)
	}

	return []*schema.ResourceData{data}, nil
}
