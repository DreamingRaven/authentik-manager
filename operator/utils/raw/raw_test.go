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
	byteData := listRawDataTag()
	tmp := &Raw{}
	t.Log(string(byteData))

	decoder := yaml.NewDecoder(bytes.NewReader(byteData))
	if err := decoder.Decode(tmp); err != nil {
		t.Fatalf("Failed to decode YAML: %v", err)
	}

	byteDataNew, err := yaml.Marshal(tmp)
	if err != nil {
		t.Fatalf("Failed to marshal YAML: %v", err)
	}
	checkByteSlicesEqual(t, byteData, byteDataNew)
}

// TestRawSimple checks if a list of Raws can be marshaled and unmarshaled
func TestRawSimpleMap(t *testing.T) {
	byteData := mapRawDataTag()
	tmp := &Raw{}
	t.Log(string(byteData))

	decoder := yaml.NewDecoder(bytes.NewReader(byteData))
	if err := decoder.Decode(tmp); err != nil {
		t.Fatalf("Failed to decode YAML: %v", err)
	}

	byteDataNew, err := yaml.Marshal(tmp)
	if err != nil {
		t.Fatalf("Failed to marshal YAML: %v", err)
	}
	checkByteSlicesEqual(t, byteData, byteDataNew)
}

// function to check if two byte slices are equal or t.fatalf
func checkByteSlicesEqual(t *testing.T, expected, actual []byte) {
	if !bytes.Equal(expected, actual) {
		t.Logf("old: %v", string(expected))
		t.Logf("new: %v", string(actual))
		t.Fatalf("Byte slices are not equal. Expected: %v, Actual: %v", expected, actual)
	}
}

func listRawDataTag() []byte {
	yamlString := `
- !Find [me, me2]
- some random string
`
	return []byte(yamlString)
}

func MixedRawData() []byte {
	yamlString := `
- some:
	  less:
	  - something
	  - smething2
	  more: !Find [me, me2]
	  mores: !Find somestring
- some random string
`
	return []byte(yamlString)
}

func mapRawDataTag() []byte {
	yamlString := `
bvv: another random string
aaa: !Find me
`
	return []byte(yamlString)
}
