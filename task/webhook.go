package task

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Webhook struct {
	Username  string  `json:"username"`
	AvatarURL string  `json:"avatar_url"`
	Embeds    []Embed `json:"embeds"`
}

type Embed struct {
	Title       string    `json:"title"`
	Color       int       `json:"color,omitempty"`
	Description string    `json:"description,omitempty"`
	Fields      []Field   `json:"fields"`
	Footer      Footer    `json:"footer,omitempty"`
	Thumbnail   Thumbnail `json:"thumbnail"`
	Timestamp   time.Time `json:"timestamp,omitempty"`
}

type Field struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type Footer struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url"`
}

type Thumbnail struct {
	URL string `json:"url"`
}

const (
	green = 65280
	red   = 16711680
)

func SendSuccessWebhook(site string, sku string, size string, postBackURL string) {
	w := Webhook{
		Username:  "Simple Bot'",
		AvatarURL: "avatar_url_here",
		Embeds: []Embed{
			{
				Title: "Payment Success",
				Color: green,
				Fields: []Field{
					{
						Name:   "Site",
						Value:  site,
						Inline: true,
					},
					{
						Name:   "SKU",
						Value:  sku,
						Inline: true,
					},
					{
						Name:   "Size",
						Value:  size,
						Inline: true,
					},
				},
				Thumbnail: Thumbnail{
					URL: fmt.Sprintf("https://images.footlocker.com/pi/%s/small/%s.jpeg", sku, sku),
				},
				Footer: Footer{
					Text:    "Simple Bot",
					IconURL: "icon_url_here",
				},
				Timestamp: time.Now(),
			},
		},
	}
	client := &http.Client{}
	jsonValue, _ := json.Marshal(w)
	req, _ := http.NewRequest("POST", postBackURL, bytes.NewBuffer(jsonValue))
	req.Header.Add("Content-Type", "application/json")
	client.Do(req)
}


func SendDeclineWebhook(site string, sku string, size string, postBackURL string) {
	w := Webhook{
		Username:  "Simple Bot'",
		AvatarURL: "avatar_url_here",
		Embeds: []Embed{
			{
				Title: "Payment Decline",
				Color: red,
				Fields: []Field{
					{
						Name:   "Site",
						Value:  site,
						Inline: true,
					},
					{
						Name:   "SKU",
						Value:  sku,
						Inline: true,
					},
					{
						Name:   "Size",
						Value:  size,
						Inline: true,
					},
				},
				Thumbnail: Thumbnail{
					URL: fmt.Sprintf("https://images.footlocker.com/pi/%s/small/%s.jpeg", sku, sku),
				},
				Footer: Footer{
					Text:    "Simple Bot",
					IconURL: "icon_url_here",
				},
				Timestamp: time.Now(),
			},
		},
	}
	client := &http.Client{}
	jsonValue, _ := json.Marshal(w)
	req, _ := http.NewRequest("POST", postBackURL, bytes.NewBuffer(jsonValue))
	req.Header.Add("Content-Type", "application/json")
	client.Do(req)
}

