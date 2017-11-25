package wns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

const AUTHURL string = "https://login.live.com/accesstoken.srf"

func Version() string {
	return "0.0.1"
}

var wnsClient = &http.Client{
	Timeout: time.Second * 10,
}

type AuthResponse struct {
	TokenType   string `json:"token_type"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type WNS struct {
	AppID        string
	ClientSecret string
	AuthToken    string
	Expiration   time.Time
	TokenStatus  bool
}

type TemplateTile struct {
	Tile   string
	Output string
}

type TemplateBadge struct {
	Field  string
	Output string
}

type TemplateToast struct {
	Text      []interface{}
	TextCount int
	Template  string
	Sound     string
	Duration  string
	Output    string
}

type TemplateToastBase struct {
	Template  string
	TextCount int
}

func Setup() {
	log.SetPrefix("[Go-WNS] ")
}

func NewConn(AppID, ClientSecret string) *WNS {
	Setup()

	w := new(WNS)
	w.AppID = AppID
	w.ClientSecret = ClientSecret
	w.TokenStatus = false
	return w
}

func (w *WNS) Auth() bool {
	resp, err := wnsClient.PostForm(AUTHURL, url.Values{"grant_type": {"client_credentials"}, "client_id": {w.AppID}, "client_secret": {w.ClientSecret}, "scope": {"notify.windows.com"}})
	if err != nil {
		log.Println(err)
	} else {
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			log.Fatalln("Error: Authentification was not successfull!")
		} else {
			body := new(AuthResponse)
			err := json.NewDecoder(resp.Body).Decode(body)
			if err != nil {
				log.Fatalln(err)
			} else {
				w.Expiration = time.Now().Add(time.Duration(body.ExpiresIn) * time.Second)
				w.AuthToken = body.AccessToken
				w.TokenStatus = true
				log.Println("Info: Authenticated")
			}
		}
	}
	return w.TokenStatus
}

func (w *WNS) reAuth() {
	dur := time.Since(w.Expiration)
	if dur.Hours() > -23 {
		if w.Auth() {
			log.Println("Info: Re-Authenticated")
		} else {
			log.Fatalln("Error: Re-Authentification failed!")
		}
	}
}

func (w *WNS) send(url, wnsType, body string) bool {
	if w.TokenStatus {
		w.reAuth()

		var xmlStr = []byte(body)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(xmlStr))
		req.Header.Set("Authorization", "Bearer "+w.AuthToken)
		req.Header.Set("Content-Type", "text/xml")
		req.Header.Set("X-WNS-Type", wnsType)

		if err != nil {
			log.Println("Error: Failed while preparing -send- request! ", err)
			return false
		}

		resp, err := wnsClient.Do(req)
		if err != nil {
			log.Println("Error: Failed to run -send- request!", err)
			return false
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			log.Println()
			log.Println("Response Status: ", resp.Status)
			log.Println("Response Headers: ", resp.Header)
			resbody, _ := ioutil.ReadAll(resp.Body)
			fmt.Println("Response Body: ", string(resbody))
			log.Println()
		}

		return resp.StatusCode == 200
	} else {
		log.Fatal("Error: Not Authenticated! Run wns.Auth()!")
		return false
	}
}

func NewTile() *TemplateTile {
	tl := new(TemplateTile)
	return tl
}

func (w *WNS) SendTile(uri string, tile *TemplateTile) bool {
	if len(tile.Output) <= 0 {
		log.Println("Error: You can't send a Tile that wasn't built using msg.Build()!")
		return false
	} else if len(uri) < 25 {
		log.Println("Error: Your URI isn't long enough!")
		return false
	} else {
		return w.send(uri, "wns/tile", tile.Output)
	}
}

func (tl *TemplateTile) SetTile(tile string) bool {
	if len(tile) > 5 {
		tl.Tile = tile
		return true
	} else {
		log.Println("Error: Your Tile XMl is empty!")
		return false
	}
}

func (tl *TemplateTile) Build() bool {
	if len(tl.Tile) <= 0 {
		log.Println("Error: Your Tile XMl is empty!")
		return false
	} else {
		tl.Output = tl.Tile
		return true
	}
}

func NewBadge() *TemplateBadge {
	tb := new(TemplateBadge)
	return tb
}

func (w *WNS) SendBadge(uri string, badge *TemplateBadge) bool {
	if len(badge.Output) <= 0 {
		log.Println("Error: You can't sent a Badge that wasn't built using msg.Build()!")
		return false
	} else if len(uri) < 25 {
		log.Println("Error: Your URI isn't long enough!")
		return false
	} else {
		return w.send(uri, "wns/badge", badge.Output)
	}
}

func (tb *TemplateBadge) SetField(field string) bool {
	if len(field) > 0 {
		tb.Field = field
	} else {
		tb.Field = "none"
	}
	return true
}

func (tb *TemplateBadge) Build() bool {
	if len(tb.Field) <= 0 {
		log.Println("Error: You didn't specify a Badge-Field!")
		return false
	} else {
		tb.Output = fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?><badge value="%s" />`, tb.Field)
		return true
	}
}

func NewToast() *TemplateToast {
	tt := new(TemplateToast)
	tt.TextCount = -1
	return tt
}

func (w *WNS) SendToast(uri string, toast *TemplateToast) bool {
	if len(toast.Output) <= 0 {
		log.Println("Error: You can't sent a Toast that wasn't built using msg.Build()!")
		return false
	} else if len(uri) < 25 {
		log.Println("Error: Your URI isn't long enough!")
		return false
	} else {
		return w.send(uri, "wns/toast", toast.Output)
	}
}

func (tt *TemplateToast) Build() bool {
	if tt.TextCount < 0 {
		log.Println("Error: You didn't specify a ToastTemplate!")
		return false
	} else if tt.TextCount != len(tt.Text) {
		log.Println("Error: The text count you supplied mismateches the one required for your ToastTemplate!")
		return false
	} else {
		populatedTemplate := fmt.Sprintf(tt.Template, tt.Text...)
		tt.Output = fmt.Sprint(`<?xml version="1.0" encoding="utf-8"?><toast`, tt.Duration, `>`, populatedTemplate, tt.Sound, `</toast>`)
		return true
	}
}

func (tt *TemplateToast) SetText(text ...string) bool {
	if tt.TextCount < 0 {
		log.Println("Error: You didn't specify a ToastTemplate!")
		return false
	} else if tt.TextCount != len(text) {
		log.Println("Error: The text count you wish to supplie mismateches the one required for your ToastTemplate!")
		return false
	} else {
		tt.Text = make([]interface{}, len(text))
		for i, t := range text {
			tt.Text[i] = t
		}
		return true
	}
}

func (tt *TemplateToast) SetTemplate(template string) bool {
	templates := map[string]TemplateToastBase{}
	templates["ToastText01"] = TemplateToastBase{TextCount: 1, Template: `<visual><binding template="ToastText01"><text id="1">%s</text></binding></visual>`}
	templates["ToastText02"] = TemplateToastBase{TextCount: 2, Template: `<visual><binding template="ToastText02"><text id="1">%s</text><text id="2">%s</text></binding></visual>`}
	templates["ToastText03"] = TemplateToastBase{TextCount: 2, Template: `<visual><binding template="ToastText03"><text id="1">%s</text><text id="2">%s</text></binding></visual>`}
	templates["ToastText04"] = TemplateToastBase{TextCount: 3, Template: `<visual><binding template="ToastText04"><text id="1">%s</text><text id="2">%s</text><text id="3">%s</text></binding></visual>`}
	templates["ToastImageAndText01"] = TemplateToastBase{TextCount: 3, Template: `<visual><binding template="ToastImageAndText01"><image id="1" src="%s" alt="%s"/><text id="1">%s</text></binding></visual>`}
	templates["ToastImageAndText02"] = TemplateToastBase{TextCount: 4, Template: `<visual><binding template="ToastImageAndText02"><image id="1" src="%s" alt="%s"/><text id="1">%s</text><text id="2">%s</text></binding></visual>`}
	templates["ToastImageAndText03"] = TemplateToastBase{TextCount: 4, Template: `<visual><binding template="ToastImageAndText03"><image id="1" src="%s" alt="%s"/><text id="1">%s</text><text id="2">%s</text></binding></visual>`}
	templates["ToastImageAndText04"] = TemplateToastBase{TextCount: 5, Template: `<visual><binding template="ToastImageAndText04"><image id="1" src="%s" alt="%s"/><text id="1">%s</text><text id="2">%s</text><text id="3">%s</text></binding></visual>`}
	if _, ok := templates[template]; ok {
		tt.Template = templates[template].Template
		tt.TextCount = templates[template].TextCount
		return true
	} else {
		log.Println("Error: Couldn't find requested template!")
		return false
	}
}

func (tt *TemplateToast) SetSound(sound string) bool {
	singleSounds := map[string]string{
		"Silent":               `<audio silent="true" />`,
		"NotificationDefault":  `<audio src="ms-winsoundevent:Notification.Default" loop="false" />`,
		"NotificationIM":       `<audio src="ms-winsoundevent:Notification.IM" loop="false" />`,
		"NotificationMail":     `<audio src="ms-winsoundevent:Notification.Mail" loop="false" />`,
		"NotificationReminder": `<audio src="ms-winsoundevent:Notification.Reminder" loop="false" />`,
		"NotificationSms":      `<audio src="ms-winsoundevent:Notification.SMS" loop="false" />`,
	}
	loopSounds := map[string]string{
		"NotificationLoopingAlarm":   `<audio src="ms-winsoundevent:Notification.Looping.Alarm" loop="true" />`,
		"NotificationLoopingAlarm2":  `<audio src="ms-winsoundevent:Notification.Looping.Alarm2" loop="true" />`,
		"NotificationLoopingAlarm3":  `<audio src="ms-winsoundevent:Notification.Looping.Alarm3" loop="true" />`,
		"NotificationLoopingAlarm4":  `<audio src="ms-winsoundevent:Notification.Looping.Alarm4" loop="true" />`,
		"NotificationLoopingAlarm5":  `<audio src="ms-winsoundevent:Notification.Looping.Alarm5" loop="true" />`,
		"NotificationLoopingAlarm6":  `<audio src="ms-winsoundevent:Notification.Looping.Alarm6" loop="true" />`,
		"NotificationLoopingAlarm7":  `<audio src="ms-winsoundevent:Notification.Looping.Alarm7" loop="true" />`,
		"NotificationLoopingAlarm8":  `<audio src="ms-winsoundevent:Notification.Looping.Alarm8" loop="true" />`,
		"NotificationLoopingAlarm9":  `<audio src="ms-winsoundevent:Notification.Looping.Alarm9" loop="true" />`,
		"NotificationLoopingAlarm10": `<audio src="ms-winsoundevent:Notification.Looping.Alarm10" loop="true" />`,
		"NotificationLoopingCall":    `<audio src="ms-winsoundevent:Notification.Looping.Call" loop="true" />`,
		"NotificationLoopingCall2":   `<audio src="ms-winsoundevent:Notification.Looping.Call2" loop="true" />`,
		"NotificationLoopingCall3":   `<audio src="ms-winsoundevent:Notification.Looping.Call3" loop="true" />`,
		"NotificationLoopingCall4":   `<audio src="ms-winsoundevent:Notification.Looping.Call4" loop="true" />`,
		"NotificationLoopingCall5":   `<audio src="ms-winsoundevent:Notification.Looping.Call5" loop="true" />`,
		"NotificationLoopingCall6":   `<audio src="ms-winsoundevent:Notification.Looping.Call6" loop="true" />`,
		"NotificationLoopingCall7":   `<audio src="ms-winsoundevent:Notification.Looping.Call7" loop="true" />`,
		"NotificationLoopingCall8":   `<audio src="ms-winsoundevent:Notification.Looping.Call8" loop="true" />`,
		"NotificationLoopingCall9":   `<audio src="ms-winsoundevent:Notification.Looping.Call9" loop="true" />`,
		"NotificationLoopingCall10":  `<audio src="ms-winsoundevent:Notification.Looping.Call10" loop="true" />`,
	}
	if _, ok := singleSounds[sound]; ok {
		tt.Sound = singleSounds[sound]
		tt.Duration = ""
		return true
	} else if _, ok := loopSounds[sound]; ok {
		tt.Sound = loopSounds[sound]
		tt.Duration = ` duration="long"`
		return true
	} else {
		log.Println("Error: Couldn't find requested sound!")
		return false
	}
}
