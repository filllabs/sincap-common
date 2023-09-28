package resources

import "github.com/filllabs/sincap-common/resources/contexts"

// Resource is the the definition of a single service resource with minimum set of predefined information.
type Resource struct {
	PathParamCxtKey contexts.ContextKey
	BodyCtxKey      contexts.ContextKey
}

// NewResource returns a new instance of resource
func NewResource() Resource {
	r := Resource{
		PathParamCxtKey: contexts.NewContextKey(),
		BodyCtxKey:      contexts.NewContextKey(),
	}
	return r
}
