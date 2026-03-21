package metrics

import "github.com/prometheus/client_golang/prometheus"

// RegisterOrPanic registers the given collectors with the registerer.
// If a collector is already registered, it is ignored.
// If any other error occurs during registration, it panics.
func RegisterOrPanic(reg prometheus.Registerer, collectors ...prometheus.Collector) {
	if reg == nil {
		return
	}
	for _, c := range collectors {
		if err := reg.Register(c); err != nil {
			if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
				panic(err)
			}
		}
	}
}
