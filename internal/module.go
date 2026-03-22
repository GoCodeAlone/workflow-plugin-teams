package internal

import (
	"context"
	"fmt"
	"io"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	messaging "github.com/GoCodeAlone/workflow-plugin-messaging-core"
	kiota "github.com/microsoft/kiota-abstractions-go"
	kiotaabstractions "github.com/microsoft/kiota-abstractions-go/authentication"
	absser "github.com/microsoft/kiota-abstractions-go/serialization"
	kiotahttp "github.com/microsoft/kiota-http-go"
	kiotajson "github.com/microsoft/kiota-serialization-json-go"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

// Compile-time check: teamsModule implements messaging.Provider.
var _ messaging.Provider = (*teamsModule)(nil)

// teamsModule creates a Graph client and registers it.
type teamsModule struct {
	name     string
	tenantID string
	clientID string
	secret   string
	teamID   string // default team for messaging.Provider methods
	baseURL  string // optional: overrides Graph API base URL (for testing)
	mockMode bool   // when true, skips OAuth and uses mock HTTP client
	client   *msgraphsdk.GraphServiceClient
}

func newTeamsModule(name string, config map[string]any) (*teamsModule, error) {
	tenantID, _ := config["tenant_id"].(string)
	clientID, _ := config["client_id"].(string)
	secret, _ := config["client_secret"].(string)
	teamID, _ := config["team_id"].(string)
	baseURL, _ := config["baseURL"].(string)
	mockMode := false
	if v, ok := config["mock_mode"].(bool); ok {
		mockMode = v
	}
	return &teamsModule{
		name:     name,
		tenantID: tenantID,
		clientID: clientID,
		secret:   secret,
		teamID:   teamID,
		baseURL:  baseURL,
		mockMode: mockMode,
	}, nil
}

// Init creates the Graph client and registers it in the global registry.
func (m *teamsModule) Init() error {
	if m.mockMode {
		// In mock mode, create a Graph client pointing to the mock HTTP server.
		// AnonymousAuthenticationProvider skips token acquisition.
		// Register JSON serializers so the SDK can serialize/deserialize requests.
		kiota.RegisterDefaultSerializer(func() absser.SerializationWriterFactory {
			return kiotajson.NewJsonSerializationWriterFactory()
		})
		kiota.RegisterDefaultDeserializer(func() absser.ParseNodeFactory {
			return kiotajson.NewJsonParseNodeFactory()
		})
		authProvider := &kiotaabstractions.AnonymousAuthenticationProvider{}
		adapter, err := kiotahttp.NewNetHttpRequestAdapterWithParseNodeFactory(authProvider, nil)
		if err != nil {
			return fmt.Errorf("teams.provider %q: mock adapter: %w", m.name, err)
		}
		baseURL := m.baseURL
		if baseURL == "" {
			baseURL = "http://localhost:19061"
		}
		adapter.SetBaseUrl(baseURL)
		client := msgraphsdk.NewGraphServiceClient(adapter)
		m.client = client
		RegisterClient(m.name, client)
		return nil
	}
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
	m.client = client
	RegisterClient(m.name, client)
	return nil
}

// Start is a no-op for this module.
func (m *teamsModule) Start(_ context.Context) error { return nil }

// Stop unregisters the Graph client.
func (m *teamsModule) Stop(_ context.Context) error {
	m.client = nil
	UnregisterClient(m.name)
	return nil
}

// ---- messaging.Provider implementation ----

// Name returns the platform identifier.
func (m *teamsModule) Name() string { return "teams" }

// SendMessage posts a text message to the given channelID (within the default team).
func (m *teamsModule) SendMessage(ctx context.Context, channelID, content string, _ *messaging.MessageOpts) (string, error) {
	if m.client == nil {
		return "", fmt.Errorf("teams.provider %q: not initialized", m.name)
	}
	if m.teamID == "" {
		return "", fmt.Errorf("teams.provider %q: team_id required for messaging.Provider", m.name)
	}
	body := models.NewChatMessage()
	msgBody := models.NewItemBody()
	msgBody.SetContent(&content)
	contentType := models.TEXT_BODYTYPE
	msgBody.SetContentType(&contentType)
	body.SetBody(msgBody)

	msg, err := m.client.Teams().ByTeamId(m.teamID).Channels().ByChannelId(channelID).Messages().Post(ctx, body, nil)
	if err != nil {
		return "", fmt.Errorf("teams: SendMessage: %w", err)
	}
	if msg.GetId() == nil {
		return "", nil
	}
	return *msg.GetId(), nil
}

// EditMessage is not supported by the Teams Graph API for channel messages.
func (m *teamsModule) EditMessage(_ context.Context, _, _, _ string) error {
	return fmt.Errorf("teams: EditMessage not supported via Graph API v1.0 for channel messages")
}

// DeleteMessage soft-deletes a channel message.
func (m *teamsModule) DeleteMessage(ctx context.Context, channelID, messageID string) error {
	if m.client == nil {
		return fmt.Errorf("teams.provider %q: not initialized", m.name)
	}
	if m.teamID == "" {
		return fmt.Errorf("teams.provider %q: team_id required for messaging.Provider", m.name)
	}
	return m.client.Teams().ByTeamId(m.teamID).Channels().ByChannelId(channelID).Messages().ByChatMessageId(messageID).Delete(ctx, nil)
}

// SendReply posts a threaded reply to a channel message.
func (m *teamsModule) SendReply(ctx context.Context, channelID, parentID, content string, _ *messaging.MessageOpts) (string, error) {
	if m.client == nil {
		return "", fmt.Errorf("teams.provider %q: not initialized", m.name)
	}
	if m.teamID == "" {
		return "", fmt.Errorf("teams.provider %q: team_id required for messaging.Provider", m.name)
	}
	body := models.NewChatMessage()
	msgBody := models.NewItemBody()
	msgBody.SetContent(&content)
	contentType := models.TEXT_BODYTYPE
	msgBody.SetContentType(&contentType)
	body.SetBody(msgBody)

	reply, err := m.client.Teams().ByTeamId(m.teamID).Channels().ByChannelId(channelID).Messages().ByChatMessageId(parentID).Replies().Post(ctx, body, nil)
	if err != nil {
		return "", fmt.Errorf("teams: SendReply: %w", err)
	}
	if reply.GetId() == nil {
		return "", nil
	}
	return *reply.GetId(), nil
}

// React is not supported via Graph API v1.0 for channel messages.
func (m *teamsModule) React(_ context.Context, _, _, _ string) error {
	return fmt.Errorf("teams: React not supported via Graph API v1.0")
}

// UploadFile uploads a file to a Teams channel's SharePoint folder.
// channelID is the Teams channel ID; the default team_id is used.
func (m *teamsModule) UploadFile(ctx context.Context, channelID string, file io.Reader, filename string) (string, error) {
	if m.client == nil {
		return "", fmt.Errorf("teams.provider %q: not initialized", m.name)
	}
	if m.teamID == "" {
		return "", fmt.Errorf("teams.provider %q: team_id required for messaging.Provider", m.name)
	}

	// Get the channel's files folder (SharePoint DriveItem)
	folder, err := m.client.Teams().ByTeamId(m.teamID).Channels().ByChannelId(channelID).FilesFolder().Get(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("teams: UploadFile: get files folder: %w", err)
	}

	driveID := ""
	if folder.GetParentReference() != nil && folder.GetParentReference().GetDriveId() != nil {
		driveID = *folder.GetParentReference().GetDriveId()
	}
	parentID := ""
	if folder.GetId() != nil {
		parentID = *folder.GetId()
	}
	if driveID == "" || parentID == "" {
		return "", fmt.Errorf("teams: UploadFile: could not determine drive/folder for channel %s", channelID)
	}

	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("teams: UploadFile: read: %w", err)
	}

	uploadPath := fmt.Sprintf("%s:/%s:/content", parentID, filename)
	item, err := m.client.Drives().ByDriveId(driveID).Items().ByDriveItemId(uploadPath).Content().Put(ctx, content, nil)
	if err != nil {
		return "", fmt.Errorf("teams: UploadFile: upload: %w", err)
	}

	if item != nil && item.GetId() != nil {
		return *item.GetId(), nil
	}
	return filename, nil
}
