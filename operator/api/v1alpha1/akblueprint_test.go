package v1alpha1

import (
	"fmt"
	"testing"

	yaml_v3 "gopkg.in/yaml.v3"
	yaml_k8s "sigs.k8s.io/yaml"
)

func TestAkBpApiYaml(t *testing.T) {
	bytes := exampleAkBlueprintYaml()
	fmt.Printf("%s\n", bytes)
	akbp := &AkBlueprint{}
	err := yaml_v3.Unmarshal(bytes, akbp)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}
	fmt.Printf("%+v\n", akbp)
	remarshalled, err := yaml_v3.Marshal(akbp)
	if err != nil {
		t.Fatalf("Failed to marshal YAML: %v", err)
	}
	fmt.Printf("%s\n", remarshalled)
	t.Fatalf("Not implemented")
}

func TestAkBpApiK8sYaml(t *testing.T) {
	bytes := exampleAkBlueprintYaml()
	fmt.Printf("%s\n", bytes)
	akbp := &AkBlueprint{}
	err := yaml_k8s.Unmarshal(bytes, akbp)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}
	fmt.Printf("%+v\n", akbp)
	remarshalled, err := yaml_k8s.Marshal(akbp)
	if err != nil {
		t.Fatalf("Failed to marshal YAML: %v", err)
	}
	fmt.Printf("%s\n", remarshalled)
	t.Fatalf("Not implemented")
}

func exampleAkBlueprintYaml() []byte {
	y := `apiVersion: akm.goauthentik.io/v1alpha1
kind: AkBlueprint
metadata:
  labels:
    app.kubernetes.io/name: akblueprint
    app.kubernetes.io/instance: akm
    app.kubernetes.io/part-of: operator
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: operator
  name: tenant
  namespace: auth
spec:
  file: /blueprints/default/default-tenant.yaml
  blueprint:
    metadata:
      name: Default - Tenant
    version: 1
    entries:
    - model: authentik_blueprints.metaapplyblueprint
      attrs:
        identifiers:
          name: Default - Authentication flow
        required: false
    - model: authentik_blueprints.metaapplyblueprint
      attrs:
        identifiers:
          name: Default - Invalidation flow
        required: false
    - model: authentik_blueprints.metaapplyblueprint
      attrs:
        identifiers:
          name: Default - User settings flow
        required: false
    - attrs:
        flow_authentication: !Find [authentik_flows.flow, [slug, default-authentication-flow]]
        flow_invalidation: !Find [authentik_flows.flow, [slug, default-invalidation-flow]]
        flow_user_settings: !Find [authentik_flows.flow, [slug, default-user-settings-flow]]
        # # TODO: (END USER) some of your customisations will likely be here
        # branding_favicon: https://URL_TO_YOUR_FAVICON.ico
        # branding_logo: https://URL_TO_YOUR_LOGO.png
        # branding_title: Your Organisation Name
        branding_title: YourBranding
      identifiers:
        domain: authentik-default
        default: True
      state: create
      model: authentik_tenants.tenant
`
	return []byte(y)
}
