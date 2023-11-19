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
func TestRawSimpleList(t *testing.T) {
	byteData := listRawData()
	tmp := Raw{}
	t.Log(string(byteData))

	decoder := yaml.NewDecoder(bytes.NewReader(byteData))
	if err := decoder.Decode(&tmp); err != nil {
		t.Fatalf("Failed to decode YAML: %v", err)
	}

	byteDataNew, err := yaml.Marshal(&tmp)
	if err != nil {
		t.Fatalf("Failed to marshal YAML: %v", err)
	}
	checkByteSlicesEqual(t, byteData, byteDataNew)
}

// TestRawSimple checks if a list of Raws can be marshaled and unmarshaled
func TestRawSimpleMap(t *testing.T) {
	byteData := mapRawData()
	tmp := Raw{}
	t.Log(string(byteData))

	decoder := yaml.NewDecoder(bytes.NewReader(byteData))
	if err := decoder.Decode(&tmp); err != nil {
		t.Fatalf("Failed to decode YAML: %v", err)
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
- !Find [me, me2]
- some random string
`
	return []byte(yamlString)
}

func mapRawData() []byte {
	yamlString := `
bvv: another random string
aaa: !Find me
`
	return []byte(yamlString)
}
