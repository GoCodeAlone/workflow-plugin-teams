package internal

import (
	"context"
	"encoding/base64"
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
		return nil, fmt.Errorf("teams_send_message: client %q not found", s.moduleName)
	}
	teamID := resolveValue("team_id", current, config)
	channelID := resolveValue("channel_id", current, config)
	content := resolveValue("content", current, config)
	if teamID == "" {
		return nil, fmt.Errorf("teams_send_message: team_id required")
	}
	if channelID == "" {
		return nil, fmt.Errorf("teams_send_message: channel_id required")
	}
	if content == "" {
		return nil, fmt.Errorf("teams_send_message: content required")
	}

	body := models.NewChatMessage()
	msgBody := models.NewItemBody()
	msgBody.SetContent(&content)
	contentType := models.TEXT_BODYTYPE
	msgBody.SetContentType(&contentType)
	body.SetBody(msgBody)

	msg, err := client.Teams().ByTeamId(teamID).Channels().ByChannelId(channelID).Messages().Post(ctx, body, nil)
	if err != nil {
		return nil, fmt.Errorf("teams_send_message: %w", err)
	}

	msgID := ""
	if msg.GetId() != nil {
		msgID = *msg.GetId()
	}
	webURL := ""
	if msg.GetWebUrl() != nil {
		webURL = *msg.GetWebUrl()
	}
	return &sdk.StepResult{Output: map[string]any{
		"id":         msgID,
		"message_id": msgID,
		"web_url":    webURL,
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
		return nil, fmt.Errorf("teams_send_card: client %q not found", s.moduleName)
	}
	teamID := resolveValue("team_id", current, config)
	channelID := resolveValue("channel_id", current, config)
	cardJSON := resolveValue("card", current, config)
	if teamID == "" {
		return nil, fmt.Errorf("teams_send_card: team_id required")
	}
	if channelID == "" {
		return nil, fmt.Errorf("teams_send_card: channel_id required")
	}
	if cardJSON == "" {
		return nil, fmt.Errorf("teams_send_card: card required")
	}

	// Validate card JSON
	var cardData map[string]any
	if err := json.Unmarshal([]byte(cardJSON), &cardData); err != nil {
		return nil, fmt.Errorf("teams_send_card: card must be valid JSON: %w", err)
	}

	body := models.NewChatMessage()
	attachment := models.NewChatMessageAttachment()
	contentType := "application/vnd.microsoft.card.adaptive"
	attachment.SetContentType(&contentType)
	attachment.SetContent(&cardJSON)
	attachID := "1"
	attachment.SetId(&attachID)
	body.SetAttachments([]models.ChatMessageAttachmentable{attachment})

	// Teams API requires an HTML body that references the attachment
	msgBody := models.NewItemBody()
	emptyContent := "<attachment id=\"1\"></attachment>"
	msgBody.SetContent(&emptyContent)
	htmlType := models.HTML_BODYTYPE
	msgBody.SetContentType(&htmlType)
	body.SetBody(msgBody)

	msg, err := client.Teams().ByTeamId(teamID).Channels().ByChannelId(channelID).Messages().Post(ctx, body, nil)
	if err != nil {
		return nil, fmt.Errorf("teams_send_card: %w", err)
	}

	msgID := ""
	if msg.GetId() != nil {
		msgID = *msg.GetId()
	}
	webURL := ""
	if msg.GetWebUrl() != nil {
		webURL = *msg.GetWebUrl()
	}
	return &sdk.StepResult{Output: map[string]any{
		"id":         msgID,
		"message_id": msgID,
		"web_url":    webURL,
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
		return nil, fmt.Errorf("teams_reply_message: client %q not found", s.moduleName)
	}
	teamID := resolveValue("team_id", current, config)
	channelID := resolveValue("channel_id", current, config)
	messageID := resolveValue("message_id", current, config)
	content := resolveValue("content", current, config)
	if teamID == "" {
		return nil, fmt.Errorf("teams_reply_message: team_id required")
	}
	if channelID == "" {
		return nil, fmt.Errorf("teams_reply_message: channel_id required")
	}
	if messageID == "" {
		return nil, fmt.Errorf("teams_reply_message: message_id required")
	}
	if content == "" {
		return nil, fmt.Errorf("teams_reply_message: content required")
	}

	body := models.NewChatMessage()
	msgBody := models.NewItemBody()
	msgBody.SetContent(&content)
	contentType := models.TEXT_BODYTYPE
	msgBody.SetContentType(&contentType)
	body.SetBody(msgBody)

	reply, err := client.Teams().ByTeamId(teamID).Channels().ByChannelId(channelID).Messages().ByChatMessageId(messageID).Replies().Post(ctx, body, nil)
	if err != nil {
		return nil, fmt.Errorf("teams_reply_message: %w", err)
	}

	replyID := ""
	if reply.GetId() != nil {
		replyID = *reply.GetId()
	}
	replyToID := messageID
	if reply.GetReplyToId() != nil {
		replyToID = *reply.GetReplyToId()
	}
	return &sdk.StepResult{Output: map[string]any{
		"id":          replyID,
		"reply_id":    replyID,
		"reply_to_id": replyToID,
		"message_id":  messageID,
		"team_id":     teamID,
		"channel_id":  channelID,
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
		return nil, fmt.Errorf("teams_delete_message: client %q not found", s.moduleName)
	}
	teamID := resolveValue("team_id", current, config)
	channelID := resolveValue("channel_id", current, config)
	messageID := resolveValue("message_id", current, config)
	if teamID == "" {
		return nil, fmt.Errorf("teams_delete_message: team_id required")
	}
	if channelID == "" {
		return nil, fmt.Errorf("teams_delete_message: channel_id required")
	}
	if messageID == "" {
		return nil, fmt.Errorf("teams_delete_message: message_id required")
	}

	err := client.Teams().ByTeamId(teamID).Channels().ByChannelId(channelID).Messages().ByChatMessageId(messageID).Delete(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("teams_delete_message: %w", err)
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
		return nil, fmt.Errorf("teams_create_channel: client %q not found", s.moduleName)
	}
	teamID := resolveValue("team_id", current, config)
	displayName := resolveValue("display_name", current, config)
	description := resolveValue("description", current, config)
	if teamID == "" {
		return nil, fmt.Errorf("teams_create_channel: team_id required")
	}
	if displayName == "" {
		return nil, fmt.Errorf("teams_create_channel: display_name required")
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
		return nil, fmt.Errorf("teams_create_channel: %w", err)
	}

	channelID := ""
	if created.GetId() != nil {
		channelID = *created.GetId()
	}
	webURL := ""
	if created.GetWebUrl() != nil {
		webURL = *created.GetWebUrl()
	}
	return &sdk.StepResult{Output: map[string]any{
		"id":           channelID,
		"channel_id":   channelID,
		"display_name": displayName,
		"web_url":      webURL,
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
		return nil, fmt.Errorf("teams_add_member: client %q not found", s.moduleName)
	}
	teamID := resolveValue("team_id", current, config)
	userID := resolveValue("user_id", current, config)
	if teamID == "" {
		return nil, fmt.Errorf("teams_add_member: team_id required")
	}
	if userID == "" {
		return nil, fmt.Errorf("teams_add_member: user_id required")
	}

	member := models.NewAadUserConversationMember()
	roles := resolveStringSlice("roles", current, config)
	if len(roles) == 0 {
		roles = []string{}
	}
	member.SetRoles(roles)
	userOdataID := fmt.Sprintf("https://graph.microsoft.com/v1.0/users/%s", userID)
	member.SetAdditionalData(map[string]any{
		"user@odata.bind": userOdataID,
	})

	added, err := client.Teams().ByTeamId(teamID).Members().Post(ctx, member, nil)
	if err != nil {
		return nil, fmt.Errorf("teams_add_member: %w", err)
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
		return nil, fmt.Errorf("teams_list_channel_messages: client %q not found", s.moduleName)
	}
	teamID := resolveValue("team_id", current, config)
	channelID := resolveValue("channel_id", current, config)
	if teamID == "" {
		return nil, fmt.Errorf("teams_list_channel_messages: team_id required")
	}
	if channelID == "" {
		return nil, fmt.Errorf("teams_list_channel_messages: channel_id required")
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
		return nil, fmt.Errorf("teams_list_channel_messages: %w", err)
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
		return nil, fmt.Errorf("teams_get_message: client %q not found", s.moduleName)
	}
	teamID := resolveValue("team_id", current, config)
	channelID := resolveValue("channel_id", current, config)
	messageID := resolveValue("message_id", current, config)
	if teamID == "" {
		return nil, fmt.Errorf("teams_get_message: team_id required")
	}
	if channelID == "" {
		return nil, fmt.Errorf("teams_get_message: channel_id required")
	}
	if messageID == "" {
		return nil, fmt.Errorf("teams_get_message: message_id required")
	}

	msg, err := client.Teams().ByTeamId(teamID).Channels().ByChannelId(channelID).Messages().ByChatMessageId(messageID).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("teams_get_message: %w", err)
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

// uploadFileStep implements step.teams_upload_file
// Uploads a file to a Teams channel's SharePoint folder via the Drive API.
// Accepts base64-encoded content in the `content` field, or a `drive_id` + `parent_item_id` to
// target a specific location directly.
type uploadFileStep struct {
	name       string
	moduleName string
}

func newUploadFileStep(name string, config map[string]any) (*uploadFileStep, error) {
	return &uploadFileStep{name: name, moduleName: getModuleName(config)}, nil
}

func (s *uploadFileStep) Execute(ctx context.Context, _ map[string]any, _ map[string]map[string]any, current map[string]any, _ map[string]any, config map[string]any) (*sdk.StepResult, error) {
	client, ok := GetClient(s.moduleName)
	if !ok {
		return nil, fmt.Errorf("teams_upload_file: client %q not found", s.moduleName)
	}
	filename := resolveValue("filename", current, config)
	contentB64 := resolveValue("content", current, config)
	if filename == "" {
		return nil, fmt.Errorf("teams_upload_file: filename required")
	}
	if contentB64 == "" {
		return nil, fmt.Errorf("teams_upload_file: content (base64-encoded) required")
	}

	fileBytes, err := base64.StdEncoding.DecodeString(contentB64)
	if err != nil {
		return nil, fmt.Errorf("teams_upload_file: content must be base64-encoded: %w", err)
	}

	// Option A: direct drive_id + parent_item_id
	driveID := resolveValue("drive_id", current, config)
	parentItemID := resolveValue("parent_item_id", current, config)

	// Option B: derive from team_id + channel_id (uses FilesFolder)
	if driveID == "" || parentItemID == "" {
		teamID := resolveValue("team_id", current, config)
		channelID := resolveValue("channel_id", current, config)
		if teamID == "" || channelID == "" {
			return nil, fmt.Errorf("teams_upload_file: provide (drive_id + parent_item_id) or (team_id + channel_id)")
		}
		folder, ferr := client.Teams().ByTeamId(teamID).Channels().ByChannelId(channelID).FilesFolder().Get(ctx, nil)
		if ferr != nil {
			return nil, fmt.Errorf("teams_upload_file: get files folder: %w", ferr)
		}
		if folder.GetParentReference() != nil && folder.GetParentReference().GetDriveId() != nil {
			driveID = *folder.GetParentReference().GetDriveId()
		}
		if folder.GetId() != nil {
			parentItemID = *folder.GetId()
		}
		if driveID == "" || parentItemID == "" {
			return nil, fmt.Errorf("teams_upload_file: could not resolve drive/folder for channel")
		}
	}

	uploadPath := fmt.Sprintf("%s:/%s:/content", parentItemID, filename)
	item, err := client.Drives().ByDriveId(driveID).Items().ByDriveItemId(uploadPath).Content().Put(ctx, fileBytes, nil)
	if err != nil {
		return nil, fmt.Errorf("teams_upload_file: %w", err)
	}

	itemID := filename
	if item != nil && item.GetId() != nil {
		itemID = *item.GetId()
	}
	webURL := ""
	if item != nil && item.GetWebUrl() != nil {
		webURL = *item.GetWebUrl()
	}
	return &sdk.StepResult{Output: map[string]any{
		"item_id":  itemID,
		"filename": filename,
		"web_url":  webURL,
	}}, nil
}
