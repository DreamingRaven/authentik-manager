// Package raw provides abstractions for using raw data from potentially tagged yaml
// files. One such example is the custom !Find tag found in Authentik blueprints.
// This yaml.Node will subsume all sub-nodes in the yaml file, the user is then
// responsible for manipulating it.
// A Key difference this abstraction makes is that all yaml tags are retained.
package raw

import (
	"fmt"
	"regexp"

	"gopkg.in/yaml.v3"
)

// A Raw abstraction so we can optionally deal with the details at a higher level.
// This is effectively an alias for yaml.Node that implements MarshalYAML and
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
type Raw yaml.Node

////////////////////////////
// Go-YAML Special Functions
////////////////////////////

// UnmarshalYAML implements the Unmarshaler interface for go-yaml
// its primary purpose is to preserve tags.
// This function will use a regex to check if a tag is a yaml tag
// or if it is an inbuilt go-yaml !!tag. The former are preserved
// in content the latter are left alone as normal.
// This function recurses down into sub-nodes to preserve tags
func (r *Raw) UnmarshalYAML(value *yaml.Node) error {
	fmt.Printf("UnmarshalYAML IN: %v\n", value)
	// we have gotten out alive
	err := unmarshalYAMLRecurse(value)
	if err != nil {
		return err
	}
	*r = Raw(*value) // update the actual content of raw by dereference
	fmt.Printf("UnmarshalYAML OUT: %v\n", r)
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
// - go-yaml node: yaml.Node
// - go-yaml intermediate format: map[string]interface{} | []interface{} | string
// - custom type: Raw
// All to prevent custom yaml tags from authentik getting mangled,
// since there was no other way.
func (r *Raw) MarshalYAML() (interface{}, error) {
	var tmp yaml.Node = yaml.Node(*r)
	// Marshal the children recursively into intermediate representation
	intermediate, err := r.MarshalChildren(&tmp)
	if err != nil {
		return nil, err
	}
	return intermediate, nil
}

// MarshalChildren is a recurive function that partially marshalls the children
// into a map[string]interface{}, a []interface{}, or a str literal value
// from a given yaml.Node based on its tags.
// This is necessary as go-yaml expects this as return from MarshalYAML
// otherwise if you go straight to string it will treat everything as a yaml
// multiline string.
func (r *Raw) MarshalChildren(value *yaml.Node) (interface{}, error) {
	// the atomic case
	if value.Kind == yaml.ScalarNode {
		return value.Value, nil
	}
	// the map case
	if value.Kind == yaml.MappingNode {
		return r.MarshalMap(value)
	}
	// the list case
	if value.Kind == yaml.SequenceNode {
		return r.MarshalList(value)
	}
	return "", fmt.Errorf("not implemented for node kind: %v", value.Kind)
}

// MarshalMap delegation function to make types easier to handle
func (r *Raw) MarshalMap(value *yaml.Node) (map[string]interface{}, error) {
	tmp := make(map[string]interface{})
	// loop over map i += 2 to since we have key and value as indipendent nodes in mapping nodes
	for i := 1; i < len(value.Content); i += 2 {
		sub, err := r.MarshalChildren(value.Content[i])
		if err != nil {
			return nil, err
		}
		tmp[value.Content[i-1].Value] = sub
	}
	return tmp, nil
}

// MarshalList delegation function to make types easier to handle
func (r *Raw) MarshalList(value *yaml.Node) ([]interface{}, error) {
	var tmp []interface{}
	for i := 0; i < len(value.Content); i++ {
		sub, err := r.MarshalChildren(value.Content[i])
		if err != nil {
			return nil, err
		}
		tmp = append(tmp, sub)
	}
	return tmp, nil
}

// FilterPipe gets the first instance of the | symbol in the message
// and replaces it with nothing returning the new bytes. This is used
// to turn a yaml literal block into a map. This does nothing if no
// pipe is found.
// defunkt since pipe is not created by the marshaller
func (r *Raw) FilterPipe(msg []byte) []byte {
	fmt.Printf("FilterPipe: %v\n", string(msg))
	newMsg := []byte{}
	// regex replace first pipe symbol
	re := regexp.MustCompile(`\|`)
	newMsg = re.ReplaceAll(msg, []byte(""))
	fmt.Printf("FilterPipe: %v\n", string(newMsg))
	return newMsg
}

////////////////////
// Private Functions
////////////////////

// unmarshalYAMLRecurse recurses into sub-nodes and applies the tag logic
func unmarshalYAMLRecurse(value *yaml.Node) error {
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
func tagToContent(value *yaml.Node) error {
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
		value.Kind = yaml.ScalarNode
		value.Tag = "!!str"
		value.Style = yaml.FlowStyle
		fmt.Printf("Node `%+v`\n", value)
	} else {
		fmt.Printf("Tag `%v` ignored `%+v`\n", value.Tag, value)
	}
	//fmt.Printf("Node `%+v`\n", value)
	return nil
}

// gatherChildrenValues returns a slice of all the values of the children
// of the provided node iteratively.
// Maps do not work with preceding yaml tags so this only works for slices / lists
func gatherChildrenValues(node *yaml.Node) []string {
	var values []string
	for i := 0; i < len(node.Content); i++ {
		values = append(values, node.Content[i].Value)
	}
	return values
}
