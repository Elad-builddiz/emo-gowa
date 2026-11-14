package usecase

import (
	"encoding/json"
	"testing"

	domainSend "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/send"
)

func TestBuildButtonsMessage(t *testing.T) {
	req := domainSend.ButtonsRequest{Body: "לאשר?", Footer: "EMO", Buttons: []domainSend.ButtonOption{{ID: "allow", Text: "✅ אשר"}, {Text: "❌ דחה"}}}
	msg := buildButtonsMessage(req)
	im := msg.GetInteractiveMessage()
	if im == nil || im.GetBody().GetText() != "לאשר?" || im.GetFooter().GetText() != "EMO" {
		t.Fatalf("body/footer wrong: %v", msg)
	}
	btns := im.GetNativeFlowMessage().GetButtons()
	if len(btns) != 2 || btns[0].GetName() != "quick_reply" {
		t.Fatalf("buttons wrong: %v", btns)
	}
	var p map[string]string
	_ = json.Unmarshal([]byte(btns[1].GetButtonParamsJSON()), &p)
	if p["display_text"] != "❌ דחה" || p["id"] != "❌ דחה" {
		t.Fatalf("params wrong: %v", p)
	}
	req.ViewOnce = true
	if buildButtonsMessage(req).GetViewOnceMessage().GetMessage().GetInteractiveMessage() == nil {
		t.Fatal("view once wrapper missing")
	}
	if n := buttonsBizNode(); n.Tag != "biz" {
		t.Fatal("biz node")
	}
}

func TestValidateButtons(t *testing.T) {
	ok := domainSend.ButtonsRequest{Body: "x", Buttons: []domainSend.ButtonOption{{ID: "1", Text: "a"}}}
	if err := validateButtons(ok); err != nil {
		t.Fatal(err)
	}
	bad := []domainSend.ButtonsRequest{
		{Body: "", Buttons: ok.Buttons},
		{Body: "x"},
		{Body: "x", Buttons: []domainSend.ButtonOption{{Text: "a"}, {Text: "b"}, {Text: "c"}, {Text: "d"}}},
		{Body: "x", Buttons: []domainSend.ButtonOption{{Text: "a"}, {Text: "a"}}},
		{Body: "x", Buttons: []domainSend.ButtonOption{{Text: "12345678901234567890123456"}}},
	}
	for i, r := range bad {
		if validateButtons(r) == nil {
			t.Fatalf("case %d should fail", i)
		}
	}
}
