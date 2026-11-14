package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	domainSend "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/send"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	pkgError "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/error"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	waBinary "go.mau.fi/whatsmeow/binary"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

const maxButtons = 3

func validateButtons(request domainSend.ButtonsRequest) error {
	if strings.TrimSpace(request.Body) == "" {
		return pkgError.ValidationError("body: cannot be blank")
	}
	if len(request.Buttons) == 0 || len(request.Buttons) > maxButtons {
		return pkgError.ValidationError(fmt.Sprintf("buttons: between 1 and %d", maxButtons))
	}
	seen := map[string]bool{}
	for i, b := range request.Buttons {
		if strings.TrimSpace(b.Text) == "" {
			return pkgError.ValidationError(fmt.Sprintf("buttons[%d].text: cannot be blank", i))
		}
		if len([]rune(b.Text)) > 25 {
			return pkgError.ValidationError(fmt.Sprintf("buttons[%d].text: at most 25 characters", i))
		}
		id := b.ID
		if id == "" {
			id = b.Text
		}
		if seen[id] {
			return pkgError.ValidationError(fmt.Sprintf("buttons[%d].id: duplicate", i))
		}
		seen[id] = true
	}
	return nil
}

// buildButtonsMessage makes the interactive native-flow message: the body text with quick-reply buttons under it.
func buildButtonsMessage(request domainSend.ButtonsRequest) *waE2E.Message {
	buttons := make([]*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton, 0, len(request.Buttons))
	for _, b := range request.Buttons {
		id := b.ID
		if id == "" {
			id = b.Text
		}
		params, _ := json.Marshal(map[string]string{"display_text": b.Text, "id": id})
		buttons = append(buttons, &waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
			Name:             proto.String("quick_reply"),
			ButtonParamsJSON: proto.String(string(params)),
		})
	}
	im := &waE2E.InteractiveMessage{
		Body: &waE2E.InteractiveMessage_Body{Text: proto.String(request.Body)},
		InteractiveMessage: &waE2E.InteractiveMessage_NativeFlowMessage_{NativeFlowMessage: &waE2E.InteractiveMessage_NativeFlowMessage{
			Buttons:        buttons,
			MessageVersion: proto.Int32(1),
		}},
	}
	if request.Header != "" {
		im.Header = &waE2E.InteractiveMessage_Header{Title: proto.String(request.Header), HasMediaAttachment: proto.Bool(false)}
	}
	if request.Footer != "" {
		im.Footer = &waE2E.InteractiveMessage_Footer{Text: proto.String(request.Footer)}
	}
	inner := &waE2E.Message{
		MessageContextInfo: &waE2E.MessageContextInfo{
			DeviceListMetadata:        &waE2E.DeviceListMetadata{},
			DeviceListMetadataVersion: proto.Int32(2),
		},
		InteractiveMessage: im,
	}
	if request.ViewOnce {
		return &waE2E.Message{ViewOnceMessage: &waE2E.FutureProofMessage{Message: inner}}
	}
	return inner
}

// buttonsBizNode is the stanza node WhatsApp expects next to a native-flow message (whatsmeow adds it only for the old buttons/list types).
func buttonsBizNode() waBinary.Node {
	return waBinary.Node{
		Tag: "biz",
		Content: []waBinary.Node{{
			Tag:   "interactive",
			Attrs: waBinary.Attrs{"type": "native_flow", "v": "1"},
			Content: []waBinary.Node{{
				Tag:   "native_flow",
				Attrs: waBinary.Attrs{"v": "9", "name": "mixed"},
			}},
		}},
	}
}

func buttonsContent(request domainSend.ButtonsRequest) string {
	var sb strings.Builder
	sb.WriteString(request.Body)
	for _, b := range request.Buttons {
		sb.WriteString("\n[" + b.Text + "]")
	}
	return sb.String()
}

func (service serviceSend) SendButtons(ctx context.Context, request domainSend.ButtonsRequest) (response domainSend.GenericResponse, err error) {
	if err = validateButtons(request); err != nil {
		return response, err
	}
	client := whatsapp.ClientFromContext(ctx)
	if client == nil {
		return response, pkgError.ErrWaCLI
	}
	recipient, err := utils.ValidateJidWithLogin(client, request.BaseRequest.Phone)
	if err != nil {
		return response, err
	}
	msg := buildButtonsMessage(request)
	nodes := []waBinary.Node{buttonsBizNode()}
	ts, err := client.SendMessage(ctx, recipient, msg, whatsmeow.SendRequestExtra{AdditionalNodes: &nodes})
	if err != nil {
		return response, normalizeSendError(err)
	}
	senderJID := whatsapp.OwnSenderJID(client)
	content := buttonsContent(request)
	go func() {
		storeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 15*time.Second)
		defer cancel()
		if err := service.chatStorageRepo.StoreSentMessageWithContext(storeCtx, ts.ID, senderJID, recipient.String(), content, ts.Timestamp, msg); err != nil {
			logrus.Warnf("Failed to store sent buttons message %s to %s: %v", ts.ID, recipient.String(), err)
		}
	}()
	response.MessageID = ts.ID
	response.Status = fmt.Sprintf("Send buttons success %s (server timestamp: %s)", request.BaseRequest.Phone, ts.Timestamp.String())
	return response, nil
}
