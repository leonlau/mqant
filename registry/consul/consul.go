package consul

import ()
import "github.com/leonlau/mqant/v2/registry"

func NewRegistry(opts ...registry.Option) registry.Registry {
	return registry.NewRegistry(opts...)
}
