package k8sutil

import (
	"log"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Builder is a generic builder for any type T
type Builder[T any] struct {
	obj *T
}

func NewBuilder[T any](obj *T) *Builder[T] {
	return &Builder[T]{obj: obj}
}

func (b *Builder[T]) Build() *T {
	return b.obj
}

// OwnableBuilder extends Builder for types that implement metav1.Object
type OwnableBuilder[T any, PT interface {
	*T
	metav1.Object
}] struct {
	*Builder[T]
}

func NewOwnableBuilder[T any, PT interface {
	*T
	metav1.Object
}](name, namespace string) *OwnableBuilder[T, PT] {
	// Create a new instance of T
	obj := new(T)
	// Convert to PT and set ObjectMeta
	pt := PT(obj)
	pt.SetName(name)
	pt.SetNamespace(namespace)
	return &OwnableBuilder[T, PT]{Builder: NewBuilder[T](obj)}
}

func (b *OwnableBuilder[T, PT]) WithOwner(owner metav1.Object, scheme *runtime.Scheme) *OwnableBuilder[T, PT] {
	if err := ctrl.SetControllerReference(owner, PT(b.obj), scheme); err != nil {
		log.Fatal(err, "Failed to set controller reference")
	}
	return b
}

// MergeMaps merges two maps, with the second map taking precedence.
func MergeMaps(base, override map[string]string) map[string]string {
	if base == nil && override == nil {
		return nil
	}
	result := make(map[string]string)
	for k, v := range base {
		result[k] = v
	}
	for k, v := range override {
		result[k] = v
	}
	return result
}

// FilterMapWithPrefix returns a new map with certain prefix from old map
func FilterMapWithPrefix(k8sMap map[string]string, prefix string) map[string]string {
	r := make(map[string]string)
	for k, v := range k8sMap {
		if strings.HasPrefix(k, prefix) {
			r[k] = v
		}
	}
	return r
}
