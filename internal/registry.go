package internal

import (
	"sync"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
)

var (
	clientMu       sync.RWMutex
	clientRegistry = make(map[string]*msgraphsdk.GraphServiceClient)
)

// RegisterClient adds a Graph client to the global registry under the given name.
func RegisterClient(name string, c *msgraphsdk.GraphServiceClient) {
	clientMu.Lock()
	defer clientMu.Unlock()
	clientRegistry[name] = c
}

// GetClient looks up a Graph client by name.
func GetClient(name string) (*msgraphsdk.GraphServiceClient, bool) {
	clientMu.RLock()
	defer clientMu.RUnlock()
	c, ok := clientRegistry[name]
	return c, ok
}

// UnregisterClient removes a client from the registry.
func UnregisterClient(name string) {
	clientMu.Lock()
	defer clientMu.Unlock()
	delete(clientRegistry, name)
}
