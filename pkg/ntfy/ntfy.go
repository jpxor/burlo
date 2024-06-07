package ntfy

import (
	"fmt"
	"net/http"
	"strings"
)

type Notify struct {
	address string
}

func New(host, topic string) Notify {
	return Notify{
		address: fmt.Sprintf("http://%s/%s", host, topic),
	}
}

func (ntfy Notify) Publish(title, message string, tags []string) error {
	req, _ := http.NewRequest("POST", ntfy.address, strings.NewReader(message))
	if len(title) > 0 {
		req.Header.Set("Title", title)
	}
	if len(tags) > 0 {
		req.Header.Set("Tags", strings.Join(tags, ","))
	}
	_, err := http.DefaultClient.Do(req)
	return err
}
