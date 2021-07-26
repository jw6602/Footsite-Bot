package task

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (s *FtsSession) CallDataDome(captchaURL string) error {
	jsData := map[string]interface{}{
		"ttst":    34.56000000187487,
		"ifov":    false,
		"wdifts":  false,
		"wdifrm":  false,
		"wdif":    false,
		"br_h":    238,
		"br_w":    1414,
		"br_oh":   782,
		"br_ow":   1414,
		"nddc":    1,
		"rs_h":    900,
		"rs_w":    1440,
		"rs_cd":   30,
		"phe":     false,
		"nm":      false,
		"jsf":     false,
		"ua":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.128 Safari/537.36",
		"lg":      "zh-CN",
		"pr":      2,
		"hc":      8,
		"ars_h":   785,
		"ars_w":   1440,
		"tz":      240,
		"str_ss":  true,
		"str_ls":  true,
		"str_idb": true,
		"str_odb": true,
		"plgod":   false,
		"plg":     2,
		"plgne":   true,
		"plgre":   true,
		"plgof":   false,
		"plggt":   false,
		"pltod":   false,
		"lb":      false,
		"eva":     33,
		"lo":      false,
		"ts_mtp":  0,
		"ts_tec":  false,
		"ts_tsa":  false,
		"vnd":     "Google Inc.",
		"bid":     "NA",
		"mmt":     "application/pdf,application/x-google-chrome-pdf",
		"plu":     "Chrome PDF Plugin,Chrome PDF Viewer",
		"hdn":     false,
		"awe":     false,
		"geb":     false,
		"dat":     false,
		"med":     "defined",
		"aco":     "probably",
		"acots":   false,
		"acmp":    "probably",
		"acmpts":  true,
		"acw":     "probably",
		"acwts":   false,
		"acma":    "maybe",
		"acmats":  false,
		"acaa":    "probably",
		"acaats":  true,
		"ac3":     "",
		"ac3ts":   false,
		"acf":     "probably",
		"acfts":   false,
		"acmp4":   "maybe",
		"acmp4ts": false,
		"acmp3":   "probably",
		"acmp3ts": false,
		"acwm":    "maybe",
		"acwmts":  false,
		"ocpt":    false,
		"vco":     "probably",
		"vcots":   false,
		"vch":     "probably",
		"vchts":   true,
		"vcw":     "probably",
		"vcwts":   true,
		"vc3":     "maybe",
		"vc3ts":   false,
		"vcmp":    "",
		"vcmpts":  false,
		"vcq":     "",
		"vcqts":   false,
		"vc1":     "probably",
		"vc1ts":   false,
		"dvm":     8,
		"sqt":     false,
		"so":      "landscape-primary",
		"wbd":     false,
		"wbdm":    true,
		"wdw":     true,
		"cokys":   "bG9hZFRpbWVzY3NpYXBwcnVudGltZQ==L=",
		"ecpc":    false,
		"lgs":     true,
		"lgsod":   false,
		"bcda":    true,
		"idn":     true,
		"capi":    false,
		"svde":    false,
		"vpbq":    true,
		"xr":      true,
		"bgav":    true,
		"rri":     true,
		"idfr":    true,
		"ancs":    true,
		"inlc":    true,
		"cgca":    true,
		"inlf":    true,
		"tecd":    true,
		"sbct":    true,
		"aflt":    true,
		"rgp":     true,
		"bint":    true,
		"spwn":    false,
		"emt":     false,
		"bfr":     false,
		"dbov":    false,
		"glvd":    "Apple",
		"glrd":    "Apple M1",
		"tagpu":   28.374999999869033,
		"prm":     true,
		"tzp":     "America/New_York",
		"cvs":     true,
		"usb":     "defined",
		"mp_cx":   441,
		"mp_cy":   223,
		"mp_tr":   true,
		"mp_mx":   -32,
		"mp_my":   48,
		"mp_sx":   467,
		"mp_sy":   355,
		"dcok":    ".captcha-delivery.com",
		"ewsi":    false,
	}

	events := []map[string]interface{}{
		{
			"source":  map[string]interface{}{"x": 0, "y": 110},
			"message": "scroll",
			"date":    time.Now().UnixNano()/int64(time.Millisecond) - 1000 - rand.Int63n(500),
			"id":      2,
		},
		{
			"source":  map[string]interface{}{"x": 594, "y": 12},
			"message": "mouse move",
			"date":    time.Now().UnixNano()/int64(time.Millisecond) - 800 - rand.Int63n(300),
			"id":      0,
		},
		{
			"source":  map[string]interface{}{"x": 0, "y": 25},
			"message": "scroll",
			"date":    time.Now().UnixNano()/int64(time.Millisecond) - 600 - rand.Int63n(100),
			"id":      2,
		},
		{
			"source":  map[string]interface{}{"x": 645, "y": 237},
			"message": "mouse move",
			"date":    time.Now().UnixNano()/int64(time.Millisecond) - 400 - rand.Int63n(50),
			"id":      0,
		},
	}

	eventCounters := map[string]interface{}{
		"mouse move":  rand.Intn(3),
		"mouse click": rand.Intn(3),
		"scroll":      rand.Intn(3),
		"touch start": 0,
		"touch end":   0,
		"touch move":  0,
		"key press":   0,
		"key down":    0,
		"key up":      0,
	}
	jsDataStr, _ := json.Marshal(jsData)
	eventsStr, _ := json.Marshal(events)
	eventCountersStr, _ := json.Marshal(eventCounters)
	u, _ := url.Parse(captchaURL)
	data := url.Values{
		"jsData":        {string(jsDataStr)},
		"events":        {string(eventsStr)},
		"eventCounters": {string(eventCountersStr)},
		"jsType":        {[]string{"le", "ch"}[rand.Intn(2)]},
		"cid":           {u.Query().Get("cid")},
		"ddk":           {"A55FBF4311ED6F1BF9911EB71931D5"},
		"Referer":       {url.QueryEscape(captchaURL)},
		"request":       {url.QueryEscape(captchaURL[32:])},
		"responsePage":  {"origin"},
		"ddv":           {"4.1.44"},
	}
	req, _ := http.NewRequest("POST", "https://api-js.datadome.co/js/", strings.NewReader(data.Encode()))
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Sec-Ch-Ua", "\"Google Chrome\";v=\"89\", \"Chromium\";v=\"89\", \";Not A Brand\";v=\"99\"")
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.128 Safari/537.36")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Origin", fmt.Sprintf("https://www.%s.com", s.Site))
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Referer", fmt.Sprintf("https://www.%s.com/", s.Site))
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	resp, err := s.Client.Do(req)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		dd := make(map[string]interface{})
		err = json.Unmarshal(body, &dd)
		if err != nil {
			return err
		}
		cookie, ok := dd["cookie"]
		if ok {
			s.DataDome = cookie.(string)[9:115]
		}
	}
	return nil
}
