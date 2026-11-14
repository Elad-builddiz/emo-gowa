package send

// ButtonsRequest sends a text with quick-reply buttons under it, as one message
// (an interactive native-flow message). A tap comes back on the webhook as a
// button reply whose selected_id is the button's ID.
type ButtonsRequest struct {
	BaseRequest
	Body    string         `json:"body" form:"body"`
	Header  string         `json:"header,omitempty" form:"header"`
	Footer  string         `json:"footer,omitempty" form:"footer"`
	Buttons []ButtonOption `json:"buttons" form:"buttons"`
	// Wrap the message as "view once" (an older way some clients need to show buttons). Default false.
	ViewOnce bool `json:"view_once,omitempty" form:"view_once"`
}

type ButtonOption struct {
	ID   string `json:"id" form:"id"`
	Text string `json:"text" form:"text"`
}
