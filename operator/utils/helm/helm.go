package helm

import (
	akmv1a1 "gitlab.com/GeorgeRaven/authentik-manager/operator/api/v1alpha1"
	yaml_v3 "gopkg.in/yaml.v3"
)

// GetAkFQDN returns the fully qualified domain name of the given Ak resource
// which is .Values.global.domain.full equivelant in helm
func GetAkFQDN(ak *akmv1a1.Ak) (string, error) {
	// Get the json raw message values
	var vals map[string]interface{}
	err := yaml_v3.Unmarshal([]byte(ak.Spec.Values), &vals)
	if err != nil {
		return "", err
	}
	// get the fqdn from the values
	fqdn := vals["global"].(map[string]interface{})["domain"].(map[string]interface{})["full"].(string)
	return fqdn, nil
}
