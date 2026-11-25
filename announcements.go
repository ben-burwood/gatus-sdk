package gatussdk

import (
	"context"
	"net/http"
)

// GetAnnouncements retrieves the system-wide announcements from Gatus.
//
// Announcements are read from the /api/v1/config endpoint.
// Gatus versions without announcement support return an empty slice.
//
// Example:
//
//	announcements, err := client.GetAnnouncements(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
//	for _, announcement := range announcements {
//		fmt.Printf("[%s] %s\n", announcement.Type, announcement.Message)
//	}
func (c *Client) GetAnnouncements(ctx context.Context) ([]Announcement, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/api/v1/config")
	if err != nil {
		return nil, err
	}
	var config struct {
		Announcements []Announcement `json:"announcements"`
	}
	if err := c.decodeResponse(resp, &config); err != nil {
		return nil, err
	}
	if config.Announcements == nil {
		config.Announcements = []Announcement{}
	}
	return config.Announcements, nil
}
