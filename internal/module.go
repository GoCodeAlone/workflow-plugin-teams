package internal

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
)

// teamsModule creates a Graph client and registers it.
type teamsModule struct {
	name     string
	tenantID string
	clientID string
	secret   string
}

func newTeamsModule(name string, config map[string]any) (*teamsModule, error) {
	tenantID, _ := config["tenant_id"].(string)
	clientID, _ := config["client_id"].(string)
	secret, _ := config["client_secret"].(string)
	return &teamsModule{
		name:     name,
		tenantID: tenantID,
		clientID: clientID,
		secret:   secret,
	}, nil
}

// Init creates the Graph client and registers it in the global registry.
func (m *teamsModule) Init() error {
	if m.tenantID == "" || m.clientID == "" || m.secret == "" {
		return fmt.Errorf("teams.provider %q: tenant_id, client_id, and client_secret are required", m.name)
	}
	cred, err := azidentity.NewClientSecretCredential(m.tenantID, m.clientID, m.secret, nil)
	if err != nil {
		return fmt.Errorf("teams.provider %q: auth: %w", m.name, err)
	}
	client, err := msgraphsdk.NewGraphServiceClientWithCredentials(cred, []string{"https://graph.microsoft.com/.default"})
	if err != nil {
		return fmt.Errorf("teams.provider %q: graph client: %w", m.name, err)
	}
	RegisterClient(m.name, client)
	return nil
}

// Start is a no-op for this module.
func (m *teamsModule) Start(_ context.Context) error { return nil }

// Stop unregisters the Graph client.
func (m *teamsModule) Stop(_ context.Context) error {
	UnregisterClient(m.name)
	return nil
}
