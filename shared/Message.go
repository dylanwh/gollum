package shared

import (
	"fmt"
	"time"
)

type Message struct {
	Text      string
	Stream    string
	Timestamp time.Time
}

func (msg Message) Format() string {
	return fmt.Sprintf("%s | %s | %s",
		msg.GetDateString(),
		msg.Stream,
		msg.Text)
}

func (msg Message) GetDateString() string {
	return fmt.Sprintf("%s", msg.Timestamp.Format("2006-01-02 15:04:05 MST"))
}
