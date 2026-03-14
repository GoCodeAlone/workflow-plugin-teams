package internal

import (
	"context"
	"encoding/json"
	"fmt"

	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/teams"
)

// sendMessageStep implements step.teams_send_message
type sendMessageStep struct {
	name       string
	moduleName string
}

func newSendMessageStep(name string, config map[string]any) (*sendMessageStep, error) {
	return &sendMessageStep{name: name, moduleName: getModuleName(config)}, nil
}

func (s *sendMessageStep) Execute(ctx context.Context, _ map[string]any, _ map[string]map[string]any, current map[string]any, _ map[string]any, config map[string]any) (*sdk.StepResult, error) {
	client, ok := GetClient(s.moduleName)
	if !ok {
		return &sdk.StepResult{Output: map[string]any{"error": "teams client not found: " + s.moduleName}}, nil
	}
	teamID := resolveValue("team_id", current, config)
	channelID := resolveValue("channel_id", current, config)
	content := resolveValue("content", current, config)
	if teamID == "" || channelID == "" || content == "" {
		return &sdk.StepResult{Output: map[string]any{"error": "team_id, channel_id, and content are required"}}, nil
	}

	body := models.NewChatMessage()
	msgBody := models.NewItemBody()
	msgBody.SetContent(&content)
	contentType := models.TEXT_BODYTYPE
	msgBody.SetContentType(&contentType)
	body.SetBody(msgBody)

	msg, err := client.Teams().ByTeamId(teamID).Channels().ByChannelId(channelID).Messages().Post(ctx, body, nil)
	if err != nil {
		return &sdk.StepResult{Output: map[string]any{"error": err.Error()}}, nil
	}

	msgID := ""
	if msg.GetId() != nil {
		msgID = *msg.GetId()
	}
	return &sdk.StepResult{Output: map[string]any{
		"message_id": msgID,
		"team_id":    teamID,
		"channel_id": channelID,
	}}, nil
}

// sendCardStep implements step.teams_send_card (Adaptive Card)
type sendCardStep struct {
	name       string
	moduleName string
}

func newSendCardStep(name string, config map[string]any) (*sendCardStep, error) {
	return &sendCardStep{name: name, moduleName: getModuleName(config)}, nil
}

func (s *sendCardStep) Execute(ctx context.Context, _ map[string]any, _ map[string]map[string]any, current map[string]any, _ map[string]any, config map[string]any) (*sdk.StepResult, error) {
	client, ok := GetClient(s.moduleName)
	if !ok {
		return &sdk.StepResult{Output: map[string]any{"error": "teams client not found: " + s.moduleName}}, nil
	}
	teamID := resolveValue("team_id", current, config)
	channelID := resolveValue("channel_id", current, config)
	cardJSON := resolveValue("card", current, config)
	if teamID == "" || channelID == "" || cardJSON == "" {
		return &sdk.StepResult{Output: map[string]any{"error": "team_id, channel_id, and card are required"}}, nil
	}

	// Validate card JSON
	var cardData map[string]any
	if err := json.Unmarshal([]byte(cardJSON), &cardData); err != nil {
		return &sdk.StepResult{Output: map[string]any{"error": "card must be valid JSON: " + err.Error()}}, nil
	}

	body := models.NewChatMessage()
	attachment := models.NewChatMessageAttachment()
	contentType := "application/vnd.microsoft.card.adaptive"
	attachment.SetContentType(&contentType)
	attachment.SetContent(&cardJSON)
	body.SetAttachments([]models.ChatMessageAttachmentable{attachment})

	// Set empty body required by Teams API when using attachments
	msgBody := models.NewItemBody()
	emptyContent := "<attachment id=\"1\"></attachment>"
	msgBody.SetContent(&emptyContent)
	htmlType := models.HTML_BODYTYPE
	msgBody.SetContentType(&htmlType)
	body.SetBody(msgBody)

	msg, err := client.Teams().ByTeamId(teamID).Channels().ByChannelId(channelID).Messages().Post(ctx, body, nil)
	if err != nil {
		return &sdk.StepResult{Output: map[string]any{"error": err.Error()}}, nil
	}

	msgID := ""
	if msg.GetId() != nil {
		msgID = *msg.GetId()
	}
	return &sdk.StepResult{Output: map[string]any{
		"message_id": msgID,
		"team_id":    teamID,
		"channel_id": channelID,
	}}, nil
}

// replyMessageStep implements step.teams_reply_message
type replyMessageStep struct {
	name       string
	moduleName string
}

func newReplyMessageStep(name string, config map[string]any) (*replyMessageStep, error) {
	return &replyMessageStep{name: name, moduleName: getModuleName(config)}, nil
}

func (s *replyMessageStep) Execute(ctx context.Context, _ map[string]any, _ map[string]map[string]any, current map[string]any, _ map[string]any, config map[string]any) (*sdk.StepResult, error) {
	client, ok := GetClient(s.moduleName)
	if !ok {
		return &sdk.StepResult{Output: map[string]any{"error": "teams client not found: " + s.moduleName}}, nil
	}
	teamID := resolveValue("team_id", current, config)
	channelID := resolveValue("channel_id", current, config)
	messageID := resolveValue("message_id", current, config)
	content := resolveValue("content", current, config)
	if teamID == "" || channelID == "" || messageID == "" || content == "" {
		return &sdk.StepResult{Output: map[string]any{"error": "team_id, channel_id, message_id, and content are required"}}, nil
	}

	body := models.NewChatMessage()
	msgBody := models.NewItemBody()
	msgBody.SetContent(&content)
	contentType := models.TEXT_BODYTYPE
	msgBody.SetContentType(&contentType)
	body.SetBody(msgBody)

	reply, err := client.Teams().ByTeamId(teamID).Channels().ByChannelId(channelID).Messages().ByChatMessageId(messageID).Replies().Post(ctx, body, nil)
	if err != nil {
		return &sdk.StepResult{Output: map[string]any{"error": err.Error()}}, nil
	}

	replyID := ""
	if reply.GetId() != nil {
		replyID = *reply.GetId()
	}
	return &sdk.StepResult{Output: map[string]any{
		"reply_id":   replyID,
		"message_id": messageID,
		"team_id":    teamID,
		"channel_id": channelID,
	}}, nil
}

// deleteMessageStep implements step.teams_delete_message
type deleteMessageStep struct {
	name       string
	moduleName string
}

func newDeleteMessageStep(name string, config map[string]any) (*deleteMessageStep, error) {
	return &deleteMessageStep{name: name, moduleName: getModuleName(config)}, nil
}

func (s *deleteMessageStep) Execute(ctx context.Context, _ map[string]any, _ map[string]map[string]any, current map[string]any, _ map[string]any, config map[string]any) (*sdk.StepResult, error) {
	client, ok := GetClient(s.moduleName)
	if !ok {
		return &sdk.StepResult{Output: map[string]any{"error": "teams client not found: " + s.moduleName}}, nil
	}
	teamID := resolveValue("team_id", current, config)
	channelID := resolveValue("channel_id", current, config)
	messageID := resolveValue("message_id", current, config)
	if teamID == "" || channelID == "" || messageID == "" {
		return &sdk.StepResult{Output: map[string]any{"error": "team_id, channel_id, and message_id are required"}}, nil
	}

	// Graph API soft-deletes channel messages
	err := client.Teams().ByTeamId(teamID).Channels().ByChannelId(channelID).Messages().ByChatMessageId(messageID).Delete(ctx, nil)
	if err != nil {
		return &sdk.StepResult{Output: map[string]any{"error": err.Error()}}, nil
	}
	return &sdk.StepResult{Output: map[string]any{"deleted": true, "message_id": messageID}}, nil
}

// createChannelStep implements step.teams_create_channel
type createChannelStep struct {
	name       string
	moduleName string
}

func newCreateChannelStep(name string, config map[string]any) (*createChannelStep, error) {
	return &createChannelStep{name: name, moduleName: getModuleName(config)}, nil
}

func (s *createChannelStep) Execute(ctx context.Context, _ map[string]any, _ map[string]map[string]any, current map[string]any, _ map[string]any, config map[string]any) (*sdk.StepResult, error) {
	client, ok := GetClient(s.moduleName)
	if !ok {
		return &sdk.StepResult{Output: map[string]any{"error": "teams client not found: " + s.moduleName}}, nil
	}
	teamID := resolveValue("team_id", current, config)
	displayName := resolveValue("display_name", current, config)
	description := resolveValue("description", current, config)
	if teamID == "" || displayName == "" {
		return &sdk.StepResult{Output: map[string]any{"error": "team_id and display_name are required"}}, nil
	}

	channel := models.NewChannel()
	channel.SetDisplayName(&displayName)
	if description != "" {
		channel.SetDescription(&description)
	}
	membershipType := models.STANDARD_CHANNELMEMBERSHIPTYPE
	if resolveValue("membership_type", current, config) == "private" {
		membershipType = models.PRIVATE_CHANNELMEMBERSHIPTYPE
	}
	channel.SetMembershipType(&membershipType)

	created, err := client.Teams().ByTeamId(teamID).Channels().Post(ctx, channel, nil)
	if err != nil {
		return &sdk.StepResult{Output: map[string]any{"error": err.Error()}}, nil
	}

	channelID := ""
	if created.GetId() != nil {
		channelID = *created.GetId()
	}
	return &sdk.StepResult{Output: map[string]any{
		"channel_id":   channelID,
		"display_name": displayName,
		"team_id":      teamID,
	}}, nil
}

// addMemberStep implements step.teams_add_member
type addMemberStep struct {
	name       string
	moduleName string
}

func newAddMemberStep(name string, config map[string]any) (*addMemberStep, error) {
	return &addMemberStep{name: name, moduleName: getModuleName(config)}, nil
}

func (s *addMemberStep) Execute(ctx context.Context, _ map[string]any, _ map[string]map[string]any, current map[string]any, _ map[string]any, config map[string]any) (*sdk.StepResult, error) {
	client, ok := GetClient(s.moduleName)
	if !ok {
		return &sdk.StepResult{Output: map[string]any{"error": "teams client not found: " + s.moduleName}}, nil
	}
	teamID := resolveValue("team_id", current, config)
	userID := resolveValue("user_id", current, config)
	if teamID == "" || userID == "" {
		return &sdk.StepResult{Output: map[string]any{"error": "team_id and user_id are required"}}, nil
	}

	member := models.NewAadUserConversationMember()
	roles := resolveStringSlice("roles", current, config)
	if len(roles) == 0 {
		roles = []string{}
	}
	member.SetRoles(roles)
	userOdataID := fmt.Sprintf("https://graph.microsoft.com/v1.0/users/%s", userID)
	additionalData := map[string]any{
		"user@odata.bind": userOdataID,
	}
	member.SetAdditionalData(additionalData)

	added, err := client.Teams().ByTeamId(teamID).Members().Post(ctx, member, nil)
	if err != nil {
		return &sdk.StepResult{Output: map[string]any{"error": err.Error()}}, nil
	}

	memberID := ""
	if added.GetId() != nil {
		memberID = *added.GetId()
	}
	return &sdk.StepResult{Output: map[string]any{
		"member_id": memberID,
		"user_id":   userID,
		"team_id":   teamID,
	}}, nil
}

// listChannelMessagesStep implements step.teams_list_channel_messages
type listChannelMessagesStep struct {
	name       string
	moduleName string
}

func newListChannelMessagesStep(name string, config map[string]any) (*listChannelMessagesStep, error) {
	return &listChannelMessagesStep{name: name, moduleName: getModuleName(config)}, nil
}

func (s *listChannelMessagesStep) Execute(ctx context.Context, _ map[string]any, _ map[string]map[string]any, current map[string]any, _ map[string]any, config map[string]any) (*sdk.StepResult, error) {
	client, ok := GetClient(s.moduleName)
	if !ok {
		return &sdk.StepResult{Output: map[string]any{"error": "teams client not found: " + s.moduleName}}, nil
	}
	teamID := resolveValue("team_id", current, config)
	channelID := resolveValue("channel_id", current, config)
	if teamID == "" || channelID == "" {
		return &sdk.StepResult{Output: map[string]any{"error": "team_id and channel_id are required"}}, nil
	}

	top := resolveInt("top", current, config)
	reqParams := &teams.ItemChannelsItemMessagesRequestBuilderGetRequestConfiguration{}
	if top > 0 {
		queryParams := &teams.ItemChannelsItemMessagesRequestBuilderGetQueryParameters{
			Top: func() *int32 { v := int32(top); return &v }(),
		}
		reqParams.QueryParameters = queryParams
	}

	result, err := client.Teams().ByTeamId(teamID).Channels().ByChannelId(channelID).Messages().Get(ctx, reqParams)
	if err != nil {
		return &sdk.StepResult{Output: map[string]any{"error": err.Error()}}, nil
	}

	messages := make([]any, 0)
	for _, msg := range result.GetValue() {
		msgID := ""
		if msg.GetId() != nil {
			msgID = *msg.GetId()
		}
		content := ""
		if msg.GetBody() != nil && msg.GetBody().GetContent() != nil {
			content = *msg.GetBody().GetContent()
		}
		messages = append(messages, map[string]any{
			"message_id": msgID,
			"content":    content,
		})
	}

	return &sdk.StepResult{Output: map[string]any{
		"messages": messages,
		"count":    len(messages),
	}}, nil
}

// getMessageStep implements step.teams_get_message
type getMessageStep struct {
	name       string
	moduleName string
}

func newGetMessageStep(name string, config map[string]any) (*getMessageStep, error) {
	return &getMessageStep{name: name, moduleName: getModuleName(config)}, nil
}

func (s *getMessageStep) Execute(ctx context.Context, _ map[string]any, _ map[string]map[string]any, current map[string]any, _ map[string]any, config map[string]any) (*sdk.StepResult, error) {
	client, ok := GetClient(s.moduleName)
	if !ok {
		return &sdk.StepResult{Output: map[string]any{"error": "teams client not found: " + s.moduleName}}, nil
	}
	teamID := resolveValue("team_id", current, config)
	channelID := resolveValue("channel_id", current, config)
	messageID := resolveValue("message_id", current, config)
	if teamID == "" || channelID == "" || messageID == "" {
		return &sdk.StepResult{Output: map[string]any{"error": "team_id, channel_id, and message_id are required"}}, nil
	}

	msg, err := client.Teams().ByTeamId(teamID).Channels().ByChannelId(channelID).Messages().ByChatMessageId(messageID).Get(ctx, nil)
	if err != nil {
		return &sdk.StepResult{Output: map[string]any{"error": err.Error()}}, nil
	}

	msgID := ""
	if msg.GetId() != nil {
		msgID = *msg.GetId()
	}
	content := ""
	if msg.GetBody() != nil && msg.GetBody().GetContent() != nil {
		content = *msg.GetBody().GetContent()
	}
	authorID := ""
	if msg.GetFrom() != nil && msg.GetFrom().GetUser() != nil && msg.GetFrom().GetUser().GetId() != nil {
		authorID = *msg.GetFrom().GetUser().GetId()
	}

	return &sdk.StepResult{Output: map[string]any{
		"message_id": msgID,
		"content":    content,
		"author_id":  authorID,
		"team_id":    teamID,
		"channel_id": channelID,
	}}, nil
}
