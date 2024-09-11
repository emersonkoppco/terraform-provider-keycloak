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

func resourceKeycloakOpenidClientOptionalScopes() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakOpenidClientOptionalScopesReconcile,
		ReadContext:   resourceKeycloakOpenidClientOptionalScopesRead,
		DeleteContext: resourceKeycloakOpenidClientOptionalScopesDelete,
		UpdateContext: resourceKeycloakOpenidClientOptionalScopesReconcile,
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakOpenidClientOptionalScopesImport,
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
			"optional_scopes": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
			},
		},
	}
}

func openidClientOptionalScopesId(realmId string, clientId string) string {
	return fmt.Sprintf("%s/%s", realmId, clientId)
}

func resourceKeycloakOpenidClientOptionalScopesRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)

	clientScopes, err := keycloakClient.GetOpenidClientOptionalScopes(ctx, realmId, clientId)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	var optionalScopes []string
	for _, clientScope := range clientScopes {
		optionalScopes = append(optionalScopes, clientScope.Name)
	}

	err = data.Set("optional_scopes", optionalScopes)
	if err != nil {
		return diag.FromErr(err)
	}
	data.SetId(openidClientOptionalScopesId(realmId, clientId))

	return nil
}

func resourceKeycloakOpenidClientOptionalScopesReconcile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)
	tfOpenidClientOptionalScopes := interfaceSliceToStringSlice(data.Get("optional_scopes").([]any))

	keycloakOpenidClientOptionalScopes, err := keycloakClient.GetOpenidClientOptionalScopes(ctx, realmId, clientId)
	if err != nil {
		if keycloak.ErrorIs404(err) {
			return diag.FromErr(fmt.Errorf("validation error: client with id %s does not exist", clientId))
		}
		return diag.FromErr(err)
	}

	var openidClientOptionalScopesToDetach []string
	for _, keycloakOpenidClientOptionalScope := range keycloakOpenidClientOptionalScopes {
		// if this scope is attached in keycloak and tf state, no update is required
		// remove it from the set so we can look at scopes that need to be attached later
		if slices.Contains(tfOpenidClientOptionalScopes, keycloakOpenidClientOptionalScope.Name) {
			tfOpenidClientOptionalScopes = slices.DeleteFunc(tfOpenidClientOptionalScopes, func(e string) bool {
				return e == keycloakOpenidClientOptionalScope.Name
			})
		} else {
			// if this scope is attached in keycloak but not in tf state, add them to a slice containing all scopes to detach
			openidClientOptionalScopesToDetach = append(openidClientOptionalScopesToDetach, keycloakOpenidClientOptionalScope.Name)
		}
	}

	// detach scopes that aren't in tf state
	err = keycloakClient.DetachOpenidClientOptionalScopes(ctx, realmId, clientId, openidClientOptionalScopesToDetach)
	if err != nil {
		return diag.FromErr(err)
	}

	// attach scopes that exist in tf state but not in keycloak
	err = keycloakClient.AttachOpenidClientOptionalScopes(ctx, realmId, clientId, tfOpenidClientOptionalScopes)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(openidClientOptionalScopesId(realmId, clientId))

	return resourceKeycloakOpenidClientOptionalScopesRead(ctx, data, meta)
}

func resourceKeycloakOpenidClientOptionalScopesDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)
	optionalScopes := interfaceSliceToStringSlice(data.Get("optional_scopes").([]any))

	return diag.FromErr(keycloakClient.DetachOpenidClientOptionalScopes(ctx, realmId, clientId, optionalScopes))
}

func resourceKeycloakOpenidClientOptionalScopesImport(ctx context.Context, data *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(data.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid import. Supported import formats: {{realmId}}/{{openidClientId}}")
	}
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := parts[0]
	clientId := parts[1]

	keycloakOpenidClientOptionalScopes, err := keycloakClient.GetOpenidClientOptionalScopes(ctx, realmId, clientId)
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
	var optionalScopes []string
	for _, clientScope := range keycloakOpenidClientOptionalScopes {
		optionalScopes = append(optionalScopes, clientScope.Name)
	}
	err = data.Set("optional_scopes", optionalScopes)
	if err != nil {
		return nil, err
	}

	diagnostics := resourceKeycloakOpenidClientOptionalScopesRead(ctx, data, meta)
	if diagnostics.HasError() {
		return nil, errors.New(diagnostics[0].Summary)
	}

	return []*schema.ResourceData{data}, nil
}
