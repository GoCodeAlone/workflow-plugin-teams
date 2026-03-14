package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

// teamsTrigger listens for Microsoft Teams events via Graph change notifications.
//
// # Important: Public HTTPS Endpoint Required
//
// The Teams trigger uses Graph API change notifications, which require Microsoft's
// servers to POST webhook events to a publicly accessible HTTPS endpoint. You must:
//   - Expose the callback_url over HTTPS with a valid TLS certificate
//   - Use a service like ngrok, a load balancer, or deploy to a public host
//   - The endpoint must respond to Graph validation requests (handled automatically)
//
// Config:
//   - team_id: Teams team ID to subscribe to
//   - channel_id: Channel ID within the team
//   - callback_url: Public HTTPS base URL Graph will POST notifications to (e.g. https://myapp.example.com)
//   - listen_addr: Local address to listen on (default :8080)
//   - module: Name of the teams.provider module to use
//
// See: https://learn.microsoft.com/en-us/graph/webhooks
type teamsTrigger struct {
	name        string
	moduleName  string
	teamID      string
	channelID   string
	callbackURL string
	listenAddr  string
	callback    sdk.TriggerCallback
	server      *http.Server
	subID       string
	mu          sync.Mutex
	cancel      context.CancelFunc
}

func newTeamsTrigger(name string, config map[string]any, callback sdk.TriggerCallback) (*teamsTrigger, error) {
	moduleName := getModuleName(config)
	teamID, _ := config["team_id"].(string)
	channelID, _ := config["channel_id"].(string)
	callbackURL, _ := config["callback_url"].(string)
	listenAddr, _ := config["listen_addr"].(string)
	if listenAddr == "" {
		listenAddr = ":8080"
	}
	return &teamsTrigger{
		name:        name,
		moduleName:  moduleName,
		teamID:      teamID,
		channelID:   channelID,
		callbackURL: callbackURL,
		listenAddr:  listenAddr,
		callback:    callback,
	}, nil
}

// Start registers a Graph change notification subscription and starts the HTTP listener.
func (t *teamsTrigger) Start(ctx context.Context) error {
	ctx, t.cancel = context.WithCancel(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/teams/notifications", t.handleNotification)

	listener, err := net.Listen("tcp", t.listenAddr)
	if err != nil {
		return fmt.Errorf("teams trigger %q: listen %s: %w", t.name, t.listenAddr, err)
	}

	t.server = &http.Server{Handler: mux}
	go func() {
		if err := t.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Printf("teams trigger %q: http server: %v", t.name, err)
		}
	}()

	// Create subscription if client and callback URL are configured
	if t.callbackURL != "" && t.teamID != "" && t.channelID != "" {
		if err := t.createSubscription(ctx); err != nil {
			log.Printf("teams trigger %q: create subscription: %v (trigger will still receive notifications if subscription exists)", t.name, err)
		}
	}

	// Renew subscription periodically (subscriptions expire after ~60 minutes for Teams)
	go t.renewLoop(ctx)

	return nil
}

// Stop cancels the subscription and shuts down the HTTP server.
func (t *teamsTrigger) Stop(ctx context.Context) error {
	if t.cancel != nil {
		t.cancel()
	}
	t.mu.Lock()
	subID := t.subID
	t.mu.Unlock()

	if subID != "" {
		client, ok := GetClient(t.moduleName)
		if ok {
			_ = client.Subscriptions().BySubscriptionId(subID).Delete(ctx, nil)
		}
	}

	if t.server != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return t.server.Shutdown(shutdownCtx)
	}
	return nil
}

func (t *teamsTrigger) createSubscription(ctx context.Context) error {
	client, ok := GetClient(t.moduleName)
	if !ok {
		return fmt.Errorf("teams client not found: %s", t.moduleName)
	}

	resource := fmt.Sprintf("/teams/%s/channels/%s/messages", t.teamID, t.channelID)
	changeType := "created,updated"
	expiry := time.Now().UTC().Add(55 * time.Minute)
	notificationURL := t.callbackURL + "/teams/notifications"
	latestTLSVersion := "v1_2"

	sub := models.NewSubscription()
	sub.SetResource(&resource)
	sub.SetChangeType(&changeType)
	sub.SetNotificationUrl(&notificationURL)
	sub.SetExpirationDateTime(&expiry)
	sub.SetLatestSupportedTlsVersion(&latestTLSVersion)

	created, err := client.Subscriptions().Post(ctx, sub, nil)
	if err != nil {
		return fmt.Errorf("create subscription: %w", err)
	}

	if created.GetId() != nil {
		t.mu.Lock()
		t.subID = *created.GetId()
		t.mu.Unlock()
	}
	return nil
}

func (t *teamsTrigger) renewLoop(ctx context.Context) {
	ticker := time.NewTicker(50 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			t.mu.Lock()
			subID := t.subID
			t.mu.Unlock()
			if subID == "" {
				continue
			}
			client, ok := GetClient(t.moduleName)
			if !ok {
				continue
			}
			newExpiry := time.Now().UTC().Add(55 * time.Minute)
			patch := models.NewSubscription()
			patch.SetExpirationDateTime(&newExpiry)
			if _, err := client.Subscriptions().BySubscriptionId(subID).Patch(ctx, patch, nil); err != nil {
				log.Printf("teams trigger %q: renew subscription: %v", t.name, err)
			}
		}
	}
}

// handleNotification processes incoming Graph change notifications.
func (t *teamsTrigger) handleNotification(w http.ResponseWriter, r *http.Request) {
	// Graph sends a validation token on first subscription — must echo it back
	if token := r.URL.Query().Get("validationToken"); token != "" {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(token))
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}

	var payload changeNotificationCollection
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "parse error", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusAccepted)

	for _, notification := range payload.Value {
		eventData := map[string]any{
			"type":            notification.ChangeType,
			"resource":        notification.Resource,
			"subscription_id": notification.SubscriptionID,
			"client_state":    notification.ClientState,
		}
		if notification.ResourceData != nil {
			if id, ok := notification.ResourceData["id"].(string); ok {
				eventData["message_id"] = id
			}
		}
		if err := t.callback(notification.ChangeType, eventData); err != nil {
			log.Printf("teams trigger %q: callback error: %v", t.name, err)
		}
	}
}

// changeNotificationCollection is the Graph API notification payload shape.
type changeNotificationCollection struct {
	Value []changeNotification `json:"value"`
}

type changeNotification struct {
	ChangeType     string         `json:"changeType"`
	Resource       string         `json:"resource"`
	SubscriptionID string         `json:"subscriptionId"`
	ClientState    string         `json:"clientState"`
	ResourceData   map[string]any `json:"resourceData"`
}
