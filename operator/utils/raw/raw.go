// Package raw provides abstractions for using raw data from potentially tagged yaml
// files. One such example is the custom !Find tag found in Authentik blueprints.
// This yaml_v3.Node will subsume all sub-nodes in the yaml file, the user is then
// responsible for manipulating it.
// A Key difference this abstraction makes is that all yaml tags are retained.
package raw

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"

	yaml_v3 "gopkg.in/yaml.v3"
)

// A Raw abstraction so we can optionally deal with the details at a higher level.
// This is effectively an alias for yaml_v3.Node that implements MarshalYAML and
// UnmarshalYAML to retain all tags.
// https://pkg.go.dev/gopkg.in/yaml.v3#Node
//
//	type Node struct {
//		// Kind defines whether the node is a document, a mapping, a sequence,
//		// a scalar value, or an alias to another node. The specific data type of
//		// scalar nodes may be obtained via the ShortTag and LongTag methods.
//		Kind Kind
//
//		// Style allows customizing the appearance of the node in the tree.
//		Style Style
//
//		// Tag holds the YAML tag defining the data type for the value.
//		// When decoding, this field will always be set to the resolved tag,
//		// even when it wasn't explicitly provided in the YAML content.
//		// When encoding, if this field is unset the value type will be
//		// implied from the node properties, and if it is set, it will only
//		// be serialized into the representation if TaggedStyle is used or
//		// the implicit tag diverges from the provided one.
//		Tag string
//
//		// Value holds the unescaped and unquoted representation of the value.
//		Value string
//
//		// Anchor holds the anchor name for this node, which allows aliases to point to it.
//		Anchor string
//
//		// Alias holds the node that this alias points to. Only valid when Kind is AliasNode.
//		Alias *Node
//
//		// Content holds contained nodes for documents, mappings, and sequences.
//		Content []*Node
//
//		// HeadComment holds any comments in the lines preceding the node and
//		// not separated by an empty line.
//		HeadComment string
//
//		// LineComment holds any comments at the end of the line where the node is in.
//		LineComment string
//
//		// FootComment holds any comments following the node and before empty lines.
//		FootComment string
//
//		// Line and Column hold the node position in the decoded YAML text.
//		// These fields are not respected when encoding the node.
//		Line   int
//		Column int
//	}
type Raw yaml_v3.Node

/////////////////////////
// Json Special Functions
/////////////////////////

// MarshalJSON is one of two json interfaces to serialise and deserialise.
// This takes a go-yaml node and serialises it to json.
func (r *Raw) MarshalJSON() ([]byte, error) {
	mp, err := r.MarshalYAML()
	if err != nil {
		return nil, err
	}
	return json.Marshal(mp)
}

// UnmarshalJSON is one of two json interfaces to serialise and deserialise.
// This takes a json byte array and deserialises it to a go-yaml node
func (r *Raw) UnmarshalJSON(data []byte) error {
	decoder := yaml_v3.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(r); err != nil {
		return err
	}
	return nil
}

////////////////////////////
// Go-YAML Special Functions
////////////////////////////

// UnmarshalYAML implements the Unmarshaler interface for go-yaml
// its primary purpose is to preserve tags.
// This function will use a regex to check if a tag is a yaml tag
// or if it is an inbuilt go-yaml !!tag. The former are preserved
// in content the latter are left alone as normal.
// This function recurses down into sub-nodes to preserve tags
func (r *Raw) UnmarshalYAML(value *yaml_v3.Node) error {
	// we have gotten out alive
	err := unmarshalYAMLRecurse(value)
	if err != nil {
		return err
	}
	*r = Raw(*value) // update the actual content of raw by dereference
	return nil
}

// MarshalYAML implements the Marshaler interface for go-yaml
// I really dont like this interface for go-yaml, and is undocumented.
// In short the interface returned can can be one of 3 things:
// - a string
// - a map[string]interface{}
// - a []interface{}
// This depends on the yamly node kind, and is the now another representation
// for this abstraction. We now have:
// - original file: bytes
// - go-yaml node: yaml_v3.Node
// - go-yaml intermediate format: map[string]interface{} | []interface{} | string
// - custom type: Raw
// All to prevent custom yaml tags from authentik getting mangled,
// since there was no other way.
func (r *Raw) MarshalYAML() (interface{}, error) {
	var tmp yaml_v3.Node = yaml_v3.Node(*r)
	// Marshal the children recursively into intermediate representation
	intermediate, err := r.marshalChildren(&tmp)
	if err != nil {
		return nil, err
	}
	return intermediate, nil
}

////////////////////
// Private Functions
////////////////////

// unmarshalYAMLRecurse recurses into sub-nodes and applies the tag logic
func unmarshalYAMLRecurse(value *yaml_v3.Node) error {
	// Check this node for tag and deal with it
	err := tagToContent(value)
	if err != nil {
		return err
	}
	// recurse into sub-nodes
	for i := 0; i < len(value.Content); i++ {
		err := unmarshalYAMLRecurse(value.Content[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// TagToContent extracts tags that match the given regex and inserts them as strings
// into the content of the original node otherwise does nothing
func tagToContent(value *yaml_v3.Node) error {
	//fmt.Printf("Node `%+v`\n", value)
	// create regex used throughout
	re := regexp.MustCompile(`^!\w+`)
	if re.MatchString(value.Tag) {
		//fmt.Printf("Tag `%v` converted\n", value.Tag)
		// move tag to value and ensure there is never trailing white-space
		if value.Value != "" {
			// when the value already exists
			value.Value = fmt.Sprintf("%v %v", value.Tag, value.Value)
		} else {
			// when we are free to create a new value
			value.Value = value.Tag
		}
		// move daughter node values into this nodes values
		// TODO: recurse over children nodes and extract values
		values := gatherChildrenValues(value)
		//fmt.Println("gathered values", values)
		if values != nil {
			value.Value = fmt.Sprintf("%v %v", value.Value, values)
		}

		// set other fields
		value.Kind = yaml_v3.ScalarNode
		value.Tag = "!!str"
		value.Style = yaml_v3.FlowStyle
		//fmt.Printf("Node `%+v`\n", value)
	} else {
		//fmt.Printf("Tag `%v` ignored `%+v`\n", value.Tag, value)
	}
	//fmt.Printf("Node `%+v`\n", value)
	return nil
}

// gatherChildrenValues returns a slice of all the values of the children
// of the provided node iteratively.
// Maps do not work with preceding yaml tags so this only works for slices / lists
func gatherChildrenValues(node *yaml_v3.Node) []string {
	var values []string
	for i := 0; i < len(node.Content); i++ {
		values = append(values, node.Content[i].Value)
	}
	return values
}

// MarshalChildren is a recurive function that partially marshalls the children
// into a map[string]interface{}, a []interface{}, or a str literal value
// from a given yaml_v3.Node based on its tags.
// This is necessary as go-yaml expects this as return from MarshalYAML
// otherwise if you go straight to string it will treat everything as a yaml
// multiline string.
func (r *Raw) marshalChildren(value *yaml_v3.Node) (interface{}, error) {
	// the atomic case
	if value.Kind == yaml_v3.ScalarNode {
		return value.Value, nil
	}
	// the map case
	if value.Kind == yaml_v3.MappingNode {
		return r.marshalMap(value)
	}
	// the list case
	if value.Kind == yaml_v3.SequenceNode {
		return r.marshalList(value)
	}
	if value.Kind == yaml_v3.DocumentNode {
		// TODO: do this for each content since there is more than just the first
		return r.marshalChildren(value.Content[0])
	}
	if value.Kind == 0 {
		//return "", fmt.Errorf("This node is in uninitialized state of kind %v", value.Kind)
		return make(map[string]interface{}), nil
	}
	return "", fmt.Errorf("not implemented for node kind: %v", value.Kind)
}

// MarshalMap delegation function to make types easier to handle
func (r *Raw) marshalMap(value *yaml_v3.Node) (map[string]interface{}, error) {
	tmp := make(map[string]interface{})
	// loop over map i += 2 since we have key and value as 1D slice in Content
	for i := 1; i < len(value.Content); i += 2 {
		sub, err := r.marshalChildren(value.Content[i])
		if err != nil {
			return nil, err
		}
		tmp[value.Content[i-1].Value] = sub
	}
	return tmp, nil
}

// MarshalList delegation function to make types easier to handle
func (r *Raw) marshalList(value *yaml_v3.Node) ([]interface{}, error) {
	var tmp []interface{}
	for i := 0; i < len(value.Content); i++ {
		sub, err := r.marshalChildren(value.Content[i])
		if err != nil {
			return nil, err
		}
		tmp = append(tmp, sub)
	}
	return tmp, nil
}

func (in *Raw) DeepCopy() *Raw {
	if in == nil {
		return nil
	}
	out := new(Raw)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Raw) DeepCopyInto(out *Raw) {
	*out = *in
	if in.Alias != nil {
		in, out := &in.Alias, &out.Alias
		*out = new(yaml_v3.Node)
		//(*in).DeepCopyInto(*out)
		deepCopyYamlNode(*in, *out)
	}
	if in.Content != nil {
		in, out := &in.Content, &out.Content
		*out = make([]*yaml_v3.Node, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(yaml_v3.Node)
				//(*in).DeepCopyInto(*out)
				deepCopyYamlNode(*in, *out)
			}
		}
	}
}

// deepCopyYamlNode is a yaml_v3.Node wrapper for DeepCopy which casts the in yaml_v3.Node
// to Raw deepcopies it and converts back to out yaml_v3.Node
// This is to allow reuse of the Raw.DeepCopyInto function
func deepCopyYamlNode(in, out *yaml_v3.Node) {
	rin := Raw(*in)
	rout := new(Raw)
	rin.DeepCopyInto(rout)
	*out = yaml_v3.Node(*rout)
}
