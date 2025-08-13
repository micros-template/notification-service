package helper

import (
	"encoding/json"
	"io"
	"mime/quotedprintable"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func RetrieveDataFromEmail(email, regex, types string, t *testing.T) string {
	var (
		mailhogResp struct {
			Total int `json:"total"`
			Items []struct {
				ID      string `json:"ID"`
				Content struct {
					Headers map[string][]string `json:"Headers"`
					Body    string              `json:"Body"`
				} `json:"Content"`
			} `json:"items"`
		}
		emailFound bool
	)
	mailhogURL := "http://localhost:8025/api/v2/messages"

	var link string
	re := regexp.MustCompile(regex)

	for range 10 {
		resp, err := http.Get(mailhogURL)
		assert.NoError(t, err)
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}

		mailhogResp.Total = 0
		err = json.NewDecoder(resp.Body).Decode(&mailhogResp)
		if err := resp.Body.Close(); err != nil {
			t.Errorf("error closing response body: %v", err)
		}
		assert.NoError(t, err)
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}

		emailFound = false
		link = ""

		for _, item := range mailhogResp.Items {
			toList := item.Content.Headers["To"]
			for _, to := range toList {
				if strings.EqualFold(strings.TrimSpace(to), email) {
					// decode body
					qpReader := quotedprintable.NewReader(strings.NewReader(item.Content.Body))
					decodedBody, err := io.ReadAll(qpReader)
					if err != nil {
						continue
					}
					bodyStr := string(decodedBody)
					bodyStr = strings.ReplaceAll(bodyStr, "&amp;", "&")
					if types == "otp" {
						matches := re.FindStringSubmatch(bodyStr)
						if len(matches) > 1 {
							link = matches[1]
							emailFound = true
							break
						} else if len(matches) == 1 {
							link = matches[0]
							emailFound = true
							break
						}
					} else {
						found := re.FindString(bodyStr)
						if found != "" {
							link = found
							emailFound = true
							break
						}
					}
				}
			}
			if emailFound {
				break
			}
		}
		if emailFound && link != "" {
			break
		}
		time.Sleep(2 * time.Second)
	}
	return link
}
