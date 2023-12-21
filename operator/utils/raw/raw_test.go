package raw

import (
	"bytes"
	"encoding/json"
	"testing"

	yaml_v3 "gopkg.in/yaml.v3"
	yaml_k8s "sigs.k8s.io/yaml"
)

///////////////////////
// TESTING k8s yaml lib
///////////////////////

func TestK8sYAMLToJSON(t *testing.T) {
	byteData := mapRawDataTag()
	tmp := &RawMapStruct{}
	decoder := yaml_v3.NewDecoder(bytes.NewReader(byteData))
	if err := decoder.Decode(tmp); err != nil {
		t.Fatalf("Failed to decode YAML: %v", err)
	}
	// Check k8s Marshal working as expected
	yb, err := yaml_k8s.Marshal(tmp)
	if err != nil {
		t.Fatalf("Failed to marshal YAML: %v", err)
	}
	// Check k8s yaml to json working as expected
	b, err := yaml_k8s.YAMLToJSON(yb)
	if err != nil {
		t.Fatalf("Failed to convert YAML to JSON: %v", err)
	}
	checkByteSlicesEqual(t, jsonRawDataTag(), b)
}

func TestK8sJSONToYAML(t *testing.T) {
	jsonData := jsonRawDataTag()
	tmp := &RawMapStruct{}
	err := yaml_k8s.Unmarshal(jsonData, tmp)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}
	yb, err := yaml_k8s.Marshal(tmp)
	if err != nil {
		t.Fatalf("Failed to marshal YAML: %v", err)
	}
	b, err := yaml_k8s.YAMLToJSON(yb)
	if err != nil {
		t.Fatalf("Failed to convert YAML to JSON: %v", err)
	}
	checkByteSlicesEqual(t, jsonRawDataTag(), b)
}

///////////////////////
// TESTING JSON LIBRARY
///////////////////////

func TestYAMLToJSON(t *testing.T) {
	byteData := mapRawDataTag()
	tmp := &RawMapStruct{}
	t.Log(string(byteData))
	decoder := yaml_v3.NewDecoder(bytes.NewReader(byteData))
	if err := decoder.Decode(tmp); err != nil {
		t.Fatalf("Failed to decode YAML: %v", err)
	}
	byteDataNew, err := json.Marshal(tmp)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}
	checkByteSlicesEqual(t, jsonRawDataTag(), byteDataNew)

}

func TestJSONToYAML(t *testing.T) {
	jsonData := jsonRawDataTag()
	tmp := &Raw{}
	decoder := yaml_v3.NewDecoder(bytes.NewReader(jsonData))
	if err := decoder.Decode(tmp); err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}
	byteDataNew, err := yaml_v3.Marshal(tmp)
	if err != nil {
		t.Fatalf("Failed to marshal YAML: %v", err)
	}
	checkByteSlicesEqual(t, mapRawDataTag(), byteDataNew)
}

/////////////////////////////////
// TESTING LIST OF RAWS WITH TAGS
/////////////////////////////////

type RawListStruct struct {
	Root *Raw `yaml:"root" json:"root"`
}

type RawMapStruct struct {
	Root *Raw `yaml:"root" json:"root"`
}

// TestRawSimple checks if a list of Raws can be marshaled and unmarshaled
func TestRawSimpleList(t *testing.T) {
	byteData := listRawDataTag()
	tmp := &RawListStruct{}
	t.Log(string(byteData))

	decoder := yaml_v3.NewDecoder(bytes.NewReader(byteData))
	if err := decoder.Decode(tmp); err != nil {
		t.Fatalf("Failed to decode YAML: %v", err)
	}

	byteDataNew, err := yaml_v3.Marshal(tmp)
	if err != nil {
		t.Fatalf("Failed to marshal YAML: %v", err)
	}

	//err = os.WriteFile("test-list.yaml", byteDataNew, 0644)
	//if err != nil {
	//	t.Fatalf("Failed to write YAML: %v", err)
	//}

	checkByteSlicesEqual(t, byteData, byteDataNew)
}

// TestRawSimple checks if a list of Raws can be marshaled and unmarshaled
func TestRawSimpleMap(t *testing.T) {
	byteData := mapRawDataTag()
	tmp := &RawMapStruct{}
	t.Log(string(byteData))

	decoder := yaml_v3.NewDecoder(bytes.NewReader(byteData))
	if err := decoder.Decode(tmp); err != nil {
		t.Fatalf("Failed to decode YAML: %v", err)
	}

	byteDataNew, err := yaml_v3.Marshal(tmp)
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
	yamlString := `root:
    - '!Find [me, me2]'
    - some random string
`
	return []byte(yamlString)
}

func MixedRawData() []byte {
	yamlString := `root:
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
	yamlString := `root:
    aaa: '!Find me'
    bvv: another random string
`
	return []byte(yamlString)
}

func jsonRawDataTag() []byte {
	jsonString := `{"root":{"aaa":"!Find me","bvv":"another random string"}}`
	return []byte(jsonString)
}
