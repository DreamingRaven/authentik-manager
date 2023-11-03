package utils

import (
	"bytes"
	"testing"

	yamlv3 "gopkg.in/yaml.v3"
)

type SimpleSchema struct {
	Value yamlv3.Node `json:"value" yaml:"value"`
}

func TestSimpleYamlNodeTag(t *testing.T) {
	simpleData := []byte("value: !Find [hello world]\n")
	t.Logf("old: %v", string(simpleData))
	ss := SimpleSchema{}
	if err := yamlv3.Unmarshal([]byte(simpleData), &ss); err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}
	simpleDataNew, err := yamlv3.Marshal(&ss)
	if err != nil {
		t.Fatalf("Failed to marshal YAML: %v", err)
	}
	t.Logf("new: %v", string(simpleDataNew))
	// compare both bytestrings to see if they are equal
	//equal := reflect.DeepEqual([]byte(simpleData), simpleDataNew)
	checkByteSlicesEqual(t, simpleData, simpleDataNew)
}

// function to check if two byte slices are equal or t.fatalf
func checkByteSlicesEqual(t *testing.T, expected, actual []byte) {
	if !bytes.Equal(expected, actual) {
		t.Fatalf("Byte slices are not equal. Expected: %v, Actual: %v", expected, actual)
	}
}

//func TestTaggable(t *testing.T) {
//	yamlData := genYamlData()
//	bp := akmv1a1.BP{}
//	if err := yaml.Unmarshal(yamlData, &bp); err != nil {
//		t.Fatalf("Failed to unmarshal YAML: %v", err)
//	}
//	yamlDataNew, err := yaml.Marshal(&bp)
//	if err != nil {
//		t.Fatalf("Failed to marshal YAML: %v", err)
//	}
//
//	if string(yamlData) != string(yamlDataNew) {
//		t.Logf("old: %v", string(yamlData))
//		t.Logf("new: %v", string(yamlDataNew))
//		t.Fatalf("Original and new YAML does not match")
//	}
//
//}

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

// getExampleBlueprint deterministically generates an example blueprint to use.
func getExampleBlueprint() string {
	return `
version: 1
metadata:
  name: Default - Authentication flow
entries: []
- model: authentik_blueprints.metaapplyblueprint
  attrs:
    identifiers:
      name: Default - Password change flow
    required: false
- attrs:
    designation: authentication
    name: Welcome to authentik!
    title: Welcome to authentik!
    authentication: none
  identifiers:
    slug: default-authentication-flow
  model: authentik_flows.flow
  id: flow
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
- identifiers:
    name: default-authentication-mfa-validation
  id: default-authentication-mfa-validation
  model: authentik_stages_authenticator_validate.authenticatorvalidatestage
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
- identifiers:
    order: 20
    stage: !KeyOf default-authentication-password
    target: !KeyOf flow
  model: authentik_flows.flowstagebinding
- identifiers:
    order: 30
    stage: !KeyOf default-authentication-mfa-validation
    target: !KeyOf flow
  model: authentik_flows.flowstagebinding
- identifiers:
    order: 100
    stage: !KeyOf default-authentication-login
    target: !KeyOf flow
  model: authentik_flows.flowstagebinding
`
}
