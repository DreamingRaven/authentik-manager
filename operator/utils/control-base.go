// Package utils implements various utilities  for general use in our controllers.
package utils

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
	"k8s.io/apimachinery/pkg/runtime"
)

// ControlBase struct centralises common controller functions into an embedded base struct
// to make the functions available with as little repetition as possible.
// https://stackoverflow.com/a/31505875
type ControlBase struct {
	client.Client
	Scheme *runtime.Scheme
}

// Control composes additional functionality we would like available to our controllers.
// This functionality is key to ensuring we KISS, and implements common routines
// like searching namespaces for resources or lists, along with common transformations.
// This does not include functions that do not require client or scheme context
// since those are better as standalone implementations rather than bundled routines.
type Control interface {}

// ListInNamespace lists resources of given group, version, kind in the given namespace.
func (c *ControlBase) ListInNamespace() {}

func (c *ControlBase) GetReleaseValues() {}

// HELM routines
