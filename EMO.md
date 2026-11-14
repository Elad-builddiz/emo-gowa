# emo-gowa

A copy of [go-whatsapp-web-multidevice](https://github.com/aldinokemal/go-whatsapp-web-multidevice) (GOWA) v9.7.0 (upstream commit `5e17e8c`, MIT licence, see `LICENCE.txt`), used by EMO's local WhatsApp connection.

One addition: `POST /send/buttons` sends a text with up to 3 quick-reply buttons as one message (an interactive native-flow message). A tap comes back on the webhook as a button reply (`selection.selected_id` = the button's id, `replied_to_id` = the buttons message).

```json
{ "phone": "9725xxxxxxxx@s.whatsapp.net", "body": "Approve?", "footer": "optional",
  "buttons": [ { "id": "allow", "text": "✅ Approve" }, { "id": "deny", "text": "❌ Reject" } ] }
```

Files: `src/domains/send/buttons.go`, `src/usecase/send_buttons.go` (+ test), the route in `src/ui/rest/send.go`.

Releases are built by `.github/workflows/release.yml` on a `v*` tag, with the same zip names as upstream (`whatsapp_<version>_<os>_<arch>.zip`).
