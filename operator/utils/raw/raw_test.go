package raw

import (
	"bytes"
	"testing"

	"gopkg.in/yaml.v3"
)


/////////////////////////////////
// TESTING LIST OF RAWS WITH TAGS
/////////////////////////////////

// TestRawSimple checks if a list of Raws can be marshaled and unmarshaled
func TestRawSimple(t *testing.T) {
	byteData := listRawData()
	type tmpStruct struct {
		Value []Raw `json:"value" yaml:"value"`
	}
	tmp := tmpStruct{}
	if err := yaml.Unmarshal(byteData, &tmp); err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}
	byteDataNew, err := yaml.Marshal(&tmp)
	if err != nil {
		t.Fatalf("Failed to marshal YAML: %v", err)
	}
	checkByteSlicesEqual(t, byteData, byteDataNew)
}

// function to check if two byte slices are equal or t.fatalf
func checkByteSlicesEqual(t *testing.T, expected, actual []byte) {
	if !bytes.Equal(expected, actual) {
		t.Fatalf("Byte slices are not equal. Expected: %v, Actual: %v", expected, actual)
	}
}

func listRawData() []byte {
	yamlString := `
value:
- !Find [hello world]
- !Find [hello world]
- !Find [hello world]
`
	return []byte(yamlString)
}
