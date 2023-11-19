package raw

import (
	"bytes"
	"fmt"
	"testing"

	"gopkg.in/yaml.v3"
)

// A Raw abstraction so we can optionaly deal with the details at a higher level
// https://pkg.go.dev/gopkg.in/yaml.v3#Node
type Raw yaml.Node

//type Node struct {
//	// Kind defines whether the node is a document, a mapping, a sequence,
//	// a scalar value, or an alias to another node. The specific data type of
//	// scalar nodes may be obtained via the ShortTag and LongTag methods.
//	Kind Kind
//
//	// Style allows customizing the apperance of the node in the tree.
//	Style Style
//
//	// Tag holds the YAML tag defining the data type for the value.
//	// When decoding, this field will always be set to the resolved tag,
//	// even when it wasn't explicitly provided in the YAML content.
//	// When encoding, if this field is unset the value type will be
//	// implied from the node properties, and if it is set, it will only
//	// be serialized into the representation if TaggedStyle is used or
//	// the implicit tag diverges from the provided one.
//	Tag string
//
//	// Value holds the unescaped and unquoted represenation of the value.
//	Value string
//
//	// Anchor holds the anchor name for this node, which allows aliases to point to it.
//	Anchor string
//
//	// Alias holds the node that this alias points to. Only valid when Kind is AliasNode.
//	Alias *Node
//
//	// Content holds contained nodes for documents, mappings, and sequences.
//	Content []*Node
//
//	// HeadComment holds any comments in the lines preceding the node and
//	// not separated by an empty line.
//	HeadComment string
//
//	// LineComment holds any comments at the end of the line where the node is in.
//	LineComment string
//
//	// FootComment holds any comments following the node and before empty lines.
//	FootComment string
//
//	// Line and Column hold the node position in the decoded YAML text.
//	// These fields are not respected when encoding the node.
//	Line   int
//	Column int
//}

func (r *Raw) UnmarshalYAML(value *yaml.Node) error{
	return fmt.Errorf("UnmarshalYAML not implemented")
}

func (r *Raw) MarshalYAML() (interface{}, error) {
	return nil, fmt.Errorf("MarshalYAML not implemented")
}

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
