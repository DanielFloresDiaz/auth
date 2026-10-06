package templatemailer

import "context"

type recordingMailClient struct {
	body string
}

func (m *recordingMailClient) Mail(_ context.Context, _, _, body string, _ map[string][]string, _ string) error {
	m.body = body
	return nil
}
