package utils

import (
	"testing"

	akmv1a1 "gitlab.com/GeorgeRaven/authentik-manager/operator/api/v1alpha1"
	"sigs.k8s.io/yaml"
)

func TestTaggable(t *testing.T) {
	yamlData := genYamlData()
	bp := akmv1a1.BP{}
	if err := yaml.Unmarshal(yamlData, &bp); err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}
	yamlDataNew, err := yaml.Marshal(&bp)
	if err != nil {
		t.Fatalf("Failed to marshal YAML: %v", err)
	}

	if string(yamlData) != string(yamlDataNew) {
		t.Logf("old: %v", string(yamlData))
		t.Logf("new: %v", string(yamlDataNew))
		t.Fatalf("Original and new YAML does not match")
	}

}

func genYamlData() []byte {
	yamlString := `
version: 1
metadata:
  name: Default - Authentication flow
entries:
- attrs:
    backends:
    - authentik.core.auth.InbuiltBackend
    - authentik.sources.ldap.auth.LDAPBackend
    - authentik.core.auth.TokenBackend
    configure_flow: !Find [authentik_flows.flow, [slug, default-password-change]]
  identifiers:
    name: default-authentication-password
  id: default-authentication-password
  model: authentik_stages_password.passwordstage
- attrs:
    user_fields:
    - email
    - username
  identifiers:
    name: default-authentication-identification
  id: default-authentication-identification
  model: authentik_stages_identification.identificationstage
- identifiers:
    name: default-authentication-login
  id: default-authentication-login
  model: authentik_stages_user_login.userloginstage
- identifiers:
    order: 10
    stage: !KeyOf default-authentication-identification
    target: !KeyOf flow
  model: authentik_flows.flowstagebinding
`
	return []byte(yamlString)
}
