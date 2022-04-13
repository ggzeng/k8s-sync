package process

import (
	"sync"

	"k8sync/gen/proto/k8sync/v1"
)

// Health implements the protobuf interface
type Health struct {
	pb.UnimplementedHealthServiceServer
	mu *sync.RWMutex
}

// NewHealth initializes a new Health struct.
func NewHealth() *Health {
	return &Health{
		mu: &sync.RWMutex{},
	}
}
