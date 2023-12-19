package raw

import (
	"bytes"
	"testing"

	yaml_v3 "gopkg.in/yaml.v3"
)

/////////////////////////
// TESTING JSON WITH TAGS
/////////////////////////

//// TestRawSimpleMapJSON checks if a list of Raws can be marshaled and unmarshaled
//func TestRawSimpleMapJSON(t *testing.T) {
//	byteData := mapRawDataTag()
//	tmp := &RawMapStruct{}
//	t.Log(string(byteData))
//	decoder := yaml_v3.NewDecoder(bytes.NewReader(byteData))
//	if err := decoder.Decode(tmp); err != nil {
//		t.Fatalf("Failed to decode YAML: %v", err)
//	}
//	fmt.Printf("%+v\n", tmp)
//	t.Fatalf("Not yet implemented")
//}

func TestYAMLToJSON(t *testing.T) {
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

func TestJSONToYAML(t *testing.T) {
	jsonData := []byte(`{"root":{"aaa":"!Find me","bvv":"another random string"}}`)
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
	Root *Raw `yaml:"root"`
}

type RawMapStruct struct {
	Root *Raw `yaml:"root"`
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
