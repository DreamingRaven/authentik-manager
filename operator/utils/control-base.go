package utils

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
	"k8s.io/apimachinery/pkg/runtime"
)

// ControlBase struct centralises common controller functions into an embeded base struct
// to make the functions available with as little repettition as possible
// https://stackoverflow.com/a/31505875
type ControlBase struct {
	client.Client
	Scheme *runtime.Scheme
}

type Control interface {}
