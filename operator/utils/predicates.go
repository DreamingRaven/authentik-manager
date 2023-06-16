package utils

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
)

// NamespacePredicate is a custom predicate that filters events based on the namespace.
type NamespacePredicate struct {
	Namespace string
}

// Create implements the predicate's Create function.
func (np NamespacePredicate) Create(evt event.CreateEvent) bool {
	//https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Object
	return np.filterEvent(evt.Object.GetNamespace())
}

// Delete implements the predicate's Delete function.
func (np NamespacePredicate) Delete(evt event.DeleteEvent) bool {
	//https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Object
	return np.filterEvent(evt.Object.GetNamespace())
}

// Generic implements the predicate's Generic function.
func (np NamespacePredicate) Generic(evt event.GenericEvent) bool {
	//https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Object
	return np.filterEvent(evt.Object.GetNamespace())
}

// Update implements the predicate's Update function.
func (np NamespacePredicate) Update(evt event.UpdateEvent) bool {
	//https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Object
	return np.filterEvent(evt.ObjectOld.GetNamespace()) || np.filterEvent(evt.ObjectNew.GetNamespace())
}

// filterEvent checks if the event's namespace matches the desired namespace.
func (np NamespacePredicate) filterEvent(eventNamespace string) bool {
	return eventNamespace == np.Namespace
}
