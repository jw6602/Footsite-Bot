package task

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	ayden "github.com/awxsam/adyen-footsites"
	"github.com/mattia-git/go-capmonster"
	uuid "github.com/nu7hatch/gouuid"
	"github.com/tidwall/gjson"
)

var (
	a         = ayden.NewAdyen("A237060180D24CDEF3E4E27D828BDB6A13E12C6959820770D7F2C1671DD0AEF4729670C20C6C5967C664D18955058B69549FBE8BF3609EF64832D7C033008A818700A9B0458641C5824F5FCBB9FF83D5A83EBDF079E73B81ACA9CA52FDBCAD7CD9D6A337A4511759FA21E34CD166B9BABD512DB7B2293C0FE48B97CAB3DE8F6F1A8E49C08D23A98E986B8A995A8F382220F06338622631435736FA064AEAC5BD223BAF42AF2B66F1FEA34EF3C297F09C10B364B994EA287A5602ACF153D0B4B09A604B987397684D19DBC5E6FE7E4FFE72390D28D6E21CA3391FA3CAADAD80A729FEF4823F6BE9711D4D51BF4DFCB6A3607686B34ACCE18329D415350FD0654D")
	c         = &capmonster.Client{APIKey: "your_cap_monster_key_here"}
	sizeIDMap sync.Map
)

const (
	FtsRecapSiteKey = "6LccSjEUAAAAANCPhaM2c-WiRxCZ5CzsjR_vd8uX"
)

type FtsSession struct {
	UUID       string       `json:"uuid"`
	CSRF       string       `json:"csrf"`
	Site       string       `json:"site"`
	SKU        string       `json:"sku"`
	Size       string       `json:"size"`
	SizeID     string       `json:"size_ID"`
	JSESSIONID string       `json:"JSESSIONID"`
	Client     *http.Client `json:"client"`
	ProxyURL   *url.URL     `json:"proxy_url"`
	DataDome   string       `json:"datadome"`
	CartGuid   string       `json:"cart-guid"`
	Profile    *Profile     `json:"profile"`
}

func (s *FtsSession) InitSession(site string, sku string, size string, proxyStr string, profile Profile) {
	s.Site = site
	s.SKU = sku
	s.Size = size
	u, _ := uuid.NewV4()
	s.UUID = u.String()
	s.ProxyURL, _ = url.Parse(proxyStr)
	s.Client = &http.Client{
		Transport: &http.Transport{Proxy: http.ProxyURL(s.ProxyURL)},
		Timeout:   5 * time.Second,
	}
	s.Profile = &profile
}

func (s *FtsSession) GenerateSession() error {
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://www.%s.com/api/v3/session?timestamp=%v", s.Site, time.Now().UnixNano()/int64(time.Millisecond)), nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.128 Safari/537.36")
	req.Header.Set("X-Fl-Request-Id", s.UUID)
	req.Header.Set("Origin", fmt.Sprintf("https://www.%s.com", s.Site))
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Referer", fmt.Sprintf("https://www.%s.com", s.Site))
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	resp, err := s.Client.Do(req)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200:
		s.CSRF = string(body)[22:58]
		for _, v := range resp.Cookies() {
			if v.Name == "JSESSIONID" {
				s.JSESSIONID = v.Value
				// fmt.Println("[200] JSESSIONID =", v.Value)
				return nil
			}
		}
	case 429:
		// TODO
		return errors.New("429")
	case 403:
		// TODO
		return errors.New("403")
	case 503:
		// TODO
		return errors.New("503")
	}
	return errors.New("UNKNOWN")
}

func (s *FtsSession) GetSizeID() error {
	sizeID, ok := sizeIDMap.Load(s.Size)
	if ok {
		s.SizeID = sizeID.(string)
		return nil
	}
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://www.%s.com/api/products/pdp/%s?timestamp=%v", s.Site, s.SKU, time.Now().UnixNano()/int64(time.Millisecond)), nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.128 Safari/537.36")
	req.Header.Set("X-Fl-Request-Id", s.UUID)
	req.Header.Set("X-Flapi-session-id", s.JSESSIONID)
	req.Header.Set("Origin", fmt.Sprintf("https://www.%s.com", s.Site))
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Referer", fmt.Sprintf("https://www.%s.com", s.Site))
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Cookie", "JSESSIONID="+s.JSESSIONID)
	resp, err := s.Client.Do(req)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200:
		result := gjson.Get(string(body), "sellableUnits")
		for _, v := range result.Array() {
			temp := v.Get("attributes")
			for _, v := range temp.Array() {
				if v.Get("value").Str == s.Size {
					s.SizeID = v.Get("id").Str
					sizeIDMap.Store(s.Size, s.SizeID)
					return nil
				}
			}
		}
		if s.SizeID == "" {
			// Unavailable Size
			return errors.New("Unavailable Size")
		}
	case 429:
		// TODO
		return errors.New("429")
	case 403:
		// TODO
		return errors.New("403")
	case 503:
		// TODO
		return errors.New("503")
	}
	return errors.New("UNKNOWN")
}

func (s *FtsSession) AddToCart() error {
	data := map[string]interface{}{
		"productId":       s.SizeID,
		"productQuantity": 1,
	}
	payloadBytes, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", fmt.Sprintf("https://www.%s.com/api/users/carts/current/entries?timestamp=%v", s.Site, time.Now().UnixNano()/int64(time.Millisecond)), bytes.NewReader(payloadBytes))
	req.Header.Set("X-Csrf-Token", s.CSRF)
	req.Header.Set("X-Fl-Productid", s.SizeID)
	req.Header.Set("X-Flapi-session-id", s.JSESSIONID)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.128 Safari/537.36")
	req.Header.Set("X-Fl-Request-Id", s.UUID)
	req.Header.Set("Origin", fmt.Sprintf("https://www.%s.com", s.Site))
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Referer", fmt.Sprintf("https://www.%s.com/product/~/%s.html", s.Site, s.SKU))
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Cookie", "JSESSIONID="+s.JSESSIONID+"; "+"datadome="+s.DataDome)
	resp, err := s.Client.Do(req)
	if err != nil {
		return err
	}
	for _, v := range resp.Cookies() {
		switch v.Name {
		case "datadome":
			s.DataDome = v.Value
		case "cart-guid":
			s.CartGuid = v.Value
		default:
			// Do Nothing
		}
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()
	switch resp.StatusCode {
	case 200:
		// SUCCESS
		// fmt.Println("[200] datadome =", s.DataDome)
		// fmt.Println("[200] cart-guild =", s.CartGuid)
		return nil
	case 429:
		// TODO
		// fmt.Println(string(body))
		return errors.New("429")
	case 403:
		// TO BE TESTED
		// Add Capmonster Support
		data := map[string]string{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			return err
		}

		captchaURL := data["url"]
		fmt.Println(captchaURL)
		// Decode captchaURL
		u, _ := url.Parse(captchaURL)
		q := u.Query()
		if q.Get("cid") == "" || q.Get("initialCid") == "" {
			return errors.New("FATAL")
		}
		token, err := c.SendRecaptchaV2(captchaURL, FtsRecapSiteKey, time.Second*25)
		if err != nil {
			return err
		}
		req, _ := http.NewRequest("GET", "https://geo.captcha-delivery.com/captcha/check", nil)

		query := req.URL.Query()
		query.Add("cid", q.Get("cid"))
		query.Add("icid", q.Get("initialCid"))

		// curl 'https://api-js.datadome.co/js/' \
		// -H 'Connection: keep-alive' \
		// -H 'sec-ch-ua: "Google Chrome";v="89", "Chromium";v="89", ";Not A Brand";v="99"' \
		// -H 'sec-ch-ua-mobile: ?0' \
		// -H 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.128 Safari/537.36' \
		// -H 'Content-type: application/x-www-form-urlencoded' \
		// -H 'Accept: */*' \
		// -H 'Origin: https://geo.captcha-delivery.com' \
		// -H 'Sec-Fetch-Site: cross-site' \
		// -H 'Sec-Fetch-Mode: cors' \
		// -H 'Sec-Fetch-Dest: empty' \
		// -H 'Referer: https://geo.captcha-delivery.com/' \
		// -H 'Accept-Language: zh-CN,zh;q=0.9' \
		// --data-raw 'jsData=%7B%22ttst%22%3A34.56000000187487%2C%22ifov%22%3Afalse%2C%22wdifts%22%3Afalse%2C%22wdifrm%22%3Afalse%2C%22wdif%22%3Afalse%2C%22br_h%22%3A238%2C%22br_w%22%3A1414%2C%22br_oh%22%3A782%2C%22br_ow%22%3A1414%2C%22nddc%22%3A1%2C%22rs_h%22%3A900%2C%22rs_w%22%3A1440%2C%22rs_cd%22%3A30%2C%22phe%22%3Afalse%2C%22nm%22%3Afalse%2C%22jsf%22%3Afalse%2C%22ua%22%3A%22Mozilla%2F5.0%20(Macintosh%3B%20Intel%20Mac%20OS%20X%2011_1_0)%20AppleWebKit%2F537.36%20(KHTML%2C%20like%20Gecko)%20Chrome%2F89.0.4389.128%20Safari%2F537.36%22%2C%22lg%22%3A%22zh-CN%22%2C%22pr%22%3A2%2C%22hc%22%3A8%2C%22ars_h%22%3A785%2C%22ars_w%22%3A1440%2C%22tz%22%3A240%2C%22str_ss%22%3Atrue%2C%22str_ls%22%3Atrue%2C%22str_idb%22%3Atrue%2C%22str_odb%22%3Atrue%2C%22plgod%22%3Afalse%2C%22plg%22%3A2%2C%22plgne%22%3Atrue%2C%22plgre%22%3Atrue%2C%22plgof%22%3Afalse%2C%22plggt%22%3Afalse%2C%22pltod%22%3Afalse%2C%22lb%22%3Afalse%2C%22eva%22%3A33%2C%22lo%22%3Afalse%2C%22ts_mtp%22%3A0%2C%22ts_tec%22%3Afalse%2C%22ts_tsa%22%3Afalse%2C%22vnd%22%3A%22Google%20Inc.%22%2C%22bid%22%3A%22NA%22%2C%22mmt%22%3A%22application%2Fpdf%2Capplication%2Fx-google-chrome-pdf%22%2C%22plu%22%3A%22Chrome%20PDF%20Plugin%2CChrome%20PDF%20Viewer%22%2C%22hdn%22%3Afalse%2C%22awe%22%3Afalse%2C%22geb%22%3Afalse%2C%22dat%22%3Afalse%2C%22med%22%3A%22defined%22%2C%22aco%22%3A%22probably%22%2C%22acots%22%3Afalse%2C%22acmp%22%3A%22probably%22%2C%22acmpts%22%3Atrue%2C%22acw%22%3A%22probably%22%2C%22acwts%22%3Afalse%2C%22acma%22%3A%22maybe%22%2C%22acmats%22%3Afalse%2C%22acaa%22%3A%22probably%22%2C%22acaats%22%3Atrue%2C%22ac3%22%3A%22%22%2C%22ac3ts%22%3Afalse%2C%22acf%22%3A%22probably%22%2C%22acfts%22%3Afalse%2C%22acmp4%22%3A%22maybe%22%2C%22acmp4ts%22%3Afalse%2C%22acmp3%22%3A%22probably%22%2C%22acmp3ts%22%3Afalse%2C%22acwm%22%3A%22maybe%22%2C%22acwmts%22%3Afalse%2C%22ocpt%22%3Afalse%2C%22vco%22%3A%22probably%22%2C%22vcots%22%3Afalse%2C%22vch%22%3A%22probably%22%2C%22vchts%22%3Atrue%2C%22vcw%22%3A%22probably%22%2C%22vcwts%22%3Atrue%2C%22vc3%22%3A%22maybe%22%2C%22vc3ts%22%3Afalse%2C%22vcmp%22%3A%22%22%2C%22vcmpts%22%3Afalse%2C%22vcq%22%3A%22%22%2C%22vcqts%22%3Afalse%2C%22vc1%22%3A%22probably%22%2C%22vc1ts%22%3Afalse%2C%22dvm%22%3A8%2C%22sqt%22%3Afalse%2C%22so%22%3A%22landscape-primary%22%2C%22wbd%22%3Afalse%2C%22wbdm%22%3Atrue%2C%22wdw%22%3Atrue%2C%22cokys%22%3A%22bG9hZFRpbWVzY3NpYXBwcnVudGltZQ%3D%3DL%3D%22%2C%22ecpc%22%3Afalse%2C%22lgs%22%3Atrue%2C%22lgsod%22%3Afalse%2C%22bcda%22%3Atrue%2C%22idn%22%3Atrue%2C%22capi%22%3Afalse%2C%22svde%22%3Afalse%2C%22vpbq%22%3Atrue%2C%22xr%22%3Atrue%2C%22bgav%22%3Atrue%2C%22rri%22%3Atrue%2C%22idfr%22%3Atrue%2C%22ancs%22%3Atrue%2C%22inlc%22%3Atrue%2C%22cgca%22%3Atrue%2C%22inlf%22%3Atrue%2C%22tecd%22%3Atrue%2C%22sbct%22%3Atrue%2C%22aflt%22%3Atrue%2C%22rgp%22%3Atrue%2C%22bint%22%3Atrue%2C%22spwn%22%3Afalse%2C%22emt%22%3Afalse%2C%22bfr%22%3Afalse%2C%22dbov%22%3Afalse%2C%22glvd%22%3A%22Apple%22%2C%22glrd%22%3A%22Apple%20M1%22%2C%22tagpu%22%3A28.374999999869033%2C%22prm%22%3Atrue%2C%22tzp%22%3A%22America%2FNew_York%22%2C%22cvs%22%3Atrue%2C%22usb%22%3A%22defined%22%2C%22mp_cx%22%3A441%2C%22mp_cy%22%3A223%2C%22mp_tr%22%3Atrue%2C%22mp_mx%22%3A-32%2C%22mp_my%22%3A48%2C%22mp_sx%22%3A467%2C%22mp_sy%22%3A355%2C%22dcok%22%3A%22.captcha-delivery.com%22%2C%22ewsi%22%3Afalse%7D&events=%5B%7B%22source%22%3A%7B%22x%22%3A0%2C%22y%22%3A110%7D%2C%22message%22%3A%22scroll%22%2C%22date%22%3A1618952086117%2C%22id%22%3A2%7D%2C%7B%22source%22%3A%7B%22x%22%3A594%2C%22y%22%3A12%7D%2C%22message%22%3A%22mouse%20move%22%2C%22date%22%3A1618952086238%2C%22id%22%3A0%7D%2C%7B%22source%22%3A%7B%22x%22%3A0%2C%22y%22%3A25%7D%2C%22message%22%3A%22scroll%22%2C%22date%22%3A1618952086413%2C%22id%22%3A2%7D%2C%7B%22source%22%3A%7B%22x%22%3A645%2C%22y%22%3A237%7D%2C%22message%22%3A%22mouse%20move%22%2C%22date%22%3A1618952096203%2C%22id%22%3A0%7D%5D&eventCounters=%7B%22mouse%20move%22%3A2%2C%22mouse%20click%22%3A0%2C%22scroll%22%3A2%2C%22touch%20start%22%3A0%2C%22touch%20end%22%3A0%2C%22touch%20move%22%3A0%2C%22key%20press%22%3A0%2C%22key%20down%22%3A0%2C%22key%20up%22%3A0%7D&jsType=le&cid=FSr.0BDa9rIfJyR2k-Nfn.P3hI~VBbRCfiS9oXSUYAO7KW6dm7C5wL88M_BtwI~hP1dfQ0Y04.tireX1~GuXP4G~z~AtsfXc4p0c_PDw4I&ddk=A55FBF4311ED6F1BF9911EB71931D5&Referer=https%253A%252F%252Fgeo.captcha-delivery.com%252Fcaptcha%252F%253FinitialCid%253DAHrlqAAAAAMATp_lr5NTZ40ASVxejQ%253D%253D%2526cid%253DRz5iEB77LUtoVBOufDhkP2jiMSOV%7EUG2QeaztMBJD1pfw05LJ6CXhQxd%7ERWfQtUVjszCmrKC6q5YmscoCi8bNJ2ihK8CuXTIknMB_f.brV%2526referer%253Dhttp%25253A%25252F%25252Fwww.footlocker.com%25252Fapi%25252Fusers%25252Fcarts%25252Fcurrent%25252Fentries%2526hash%253DA55FBF4311ED6F1BF9911EB71931D5%2526t%253Dbv%2526s%253D17434&request=%252Fcaptcha%252F%253FinitialCid%253DAHrlqAAAAAMATp_lr5NTZ40ASVxejQ%253D%253D%2526cid%253DRz5iEB77LUtoVBOufDhkP2jiMSOV%7EUG2QeaztMBJD1pfw05LJ6CXhQxd%7ERWfQtUVjszCmrKC6q5YmscoCi8bNJ2ihK8CuXTIknMB_f.brV%2526referer%253Dhttp%25253A%25252F%25252Fwww.footlocker.com%25252Fapi%25252Fusers%25252Fcarts%25252Fcurrent%25252Fentries%2526hash%253DA55FBF4311ED6F1BF9911EB71931D5%2526t%253Dbv%2526s%253D17434&responsePage=hard-block&ddv=4.1.44' \
		// --compressed

		query.Add("g-recaptcha-response", token)
		query.Add("hash", q.Get("hash"))
		query.Add("ua", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.128 Safari/537.36")
		query.Add("referer", fmt.Sprintf("https://www.%s.com/api/users/carts/current/entries", s.Site))
		query.Add("parent_url", captchaURL)
		query.Add("x-forwarded-for", "")
		// query.Add("captchaChallenge", "")
		// query.Add("ccid", "datadome_cookie_here")
		query.Add("s", q.Get("s"))
		req.URL.RawQuery = query.Encode()
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
			// The dd cookie is in response body :)
			dd := make(map[string]interface{})
			err = json.Unmarshal(body, &dd)
			if err != nil {
				return err
			}
			cookie, ok := dd["cookie"]
			if ok {
				s.DataDome = cookie.(string)[9:115]
			}
			fmt.Println(s.DataDome)
			return errors.New("DONE")
		}

		return errors.New("403")
	case 503:
		// TODO
		return errors.New("503")
	}
	return errors.New("UNKNOWN")
}

func (s *FtsSession) SubmitEmail() error {
	req, _ := http.NewRequest("PUT", fmt.Sprintf("https://www.%s.com/api/users/carts/current/email/%s?timestamp=%v", s.Site, s.Profile.Email, time.Now().UnixNano()/int64(time.Millisecond)), nil)
	req.Header.Set("X-Csrf-Token", s.CSRF)
	req.Header.Set("X-Fl-Productid", s.SizeID)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.128 Safari/537.36")
	req.Header.Set("X-Fl-Request-Id", s.UUID)
	req.Header.Set("X-Flapi-session-id", s.JSESSIONID)
	req.Header.Set("Origin", fmt.Sprintf("https://www.%s.com", s.Site))
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Referer", fmt.Sprintf("https://www.%s.com/product/~/%s.html", s.Site, s.SKU))
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Cookie", "JSESSIONID="+s.JSESSIONID+"; "+"datadome="+s.DataDome+"; "+"cart-guid="+s.CartGuid)
	resp, err := s.Client.Do(req)
	if err != nil {
		return err
	}
	ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200:
		// SUCCESS
		for _, v := range resp.Cookies() {
			switch v.Name {
			case "datadome":
				s.DataDome = v.Value
			default:
				// Do Nothing
			}
		}
		return nil
	case 429:
		// TODO
		return errors.New("429")
	case 403:
		// TODO
		return errors.New("403")
	case 503:
		// TODO
		return errors.New("503")
	}
	return errors.New("UNKNOWN")
}

func (s *FtsSession) SubmitShipping() error {
	isoState, _ := mapkey(usc, s.Profile.State)
	data := map[string]interface{}{
		"shippingAddress": map[string]interface{}{
			"setAsDefaultBilling":  false,
			"setAsDefaultShipping": false,
			"firstName":            s.Profile.FirstName,
			"lastName":             s.Profile.LastName,
			"email":                s.Profile.Email,
			"phone":                s.Profile.Phone,
			"country": map[string]string{
				"isocode": "US",
				"name":    "United States",
			},
			"id":                nil,
			"setAsBilling":      true,
			"saveInAddressBook": false,
			"region": map[string]string{
				"countryIso":   "US",
				"isocode":      "US-" + isoState,
				"isocodeShort": isoState,
				"name":         s.Profile.State,
			},
			"type":            "default",
			"LoqateSearch":    "",
			"line1":           s.Profile.Line1,
			"line2":           s.Profile.Line2,
			"postalCode":      s.Profile.PostalCode,
			"town":            strings.ToUpper(s.Profile.City),
			"regionFPO":       nil,
			"shippingAddress": true,
			"recordType":      "H",
		},
	}
	payloadBytes, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", fmt.Sprintf("https://www.%s.com/api/users/carts/current/addresses/shipping?timestamp=%v", s.Site, time.Now().UnixNano()/int64(time.Millisecond)), bytes.NewReader(payloadBytes))
	req.Header.Set("X-Csrf-Token", s.CSRF)
	req.Header.Set("X-Fl-Productid", s.SizeID)
	req.Header.Set("X-Flapi-session-id", s.JSESSIONID)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.128 Safari/537.36")
	req.Header.Set("X-Fl-Request-Id", s.UUID)
	req.Header.Set("Origin", fmt.Sprintf("https://www.%s.com", s.Site))
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Referer", fmt.Sprintf("https://www.%s.com/product/~/%s.html", s.Site, s.SKU))
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Cookie", "JSESSIONID="+s.JSESSIONID+"; "+"datadome="+s.DataDome+"; "+"cart-guid="+s.CartGuid)
	resp, err := s.Client.Do(req)
	if err != nil {
		return err
	}
	ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200, 201:
		// SUCCESS
		for _, v := range resp.Cookies() {
			switch v.Name {
			case "datadome":
				s.DataDome = v.Value
			default:
				// Do Nothing
			}
		}
		return nil
	case 429:
		// TODO
		return errors.New("429")
	case 403:
		// TODO
		return errors.New("403")
	case 503:
		// TODO
		return errors.New("503")
	}
	return errors.New("UNKNOWN")
}

func (s *FtsSession) SubmitBilling() error {
	isoState, _ := mapkey(usc, s.Profile.State)
	data := map[string]interface{}{
		"setAsDefaultBilling":  false,
		"setAsDefaultShipping": false,
		"firstName":            s.Profile.FirstName,
		"lastName":             s.Profile.LastName,
		"email":                s.Profile.Email,
		"phone":                s.Profile.Phone,
		"country": map[string]string{
			"isocode": "US",
			"name":    "United States",
		},
		"id":                nil,
		"setAsBilling":      true,
		"saveInAddressBook": false,
		"region": map[string]string{
			"countryIso":   "US",
			"isocode":      "US-" + isoState,
			"isocodeShort": isoState,
			"name":         s.Profile.State,
		},
		"type":            "default",
		"LoqateSearch":    "",
		"line1":           s.Profile.Line1,
		"line2":           s.Profile.Line2,
		"postalCode":      s.Profile.PostalCode,
		"town":            strings.ToUpper(s.Profile.City),
		"regionFPO":       nil,
		"shippingAddress": true,
		"recordType":      "H",
	}

	payloadBytes, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", fmt.Sprintf("https://www.%s.com/api/users/carts/current/set-billing?timestamp=%v", s.Site, time.Now().UnixNano()/int64(time.Millisecond)), bytes.NewReader(payloadBytes))
	req.Header.Set("X-Csrf-Token", s.CSRF)
	req.Header.Set("X-Fl-Productid", s.SizeID)
	req.Header.Set("X-Flapi-session-id", s.JSESSIONID)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.128 Safari/537.36")
	req.Header.Set("X-Fl-Request-Id", s.UUID)
	req.Header.Set("Origin", fmt.Sprintf("https://www.%s.com", s.Site))
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Referer", fmt.Sprintf("https://www.%s.com/product/~/%s.html", s.Site, s.SKU))
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Cookie", "JSESSIONID="+s.JSESSIONID+"; "+"datadome="+s.DataDome+"; "+"cart-guid="+s.CartGuid)
	resp, err := s.Client.Do(req)
	if err != nil {
		return err
	}
	ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200, 201:
		// SUCCESS
		for _, v := range resp.Cookies() {
			switch v.Name {
			case "datadome":
				s.DataDome = v.Value
			default:
				// Do Nothing
			}
		}
		return nil
	case 429:
		// TODO
		return errors.New("429")
	case 403:
		// TODO
		return errors.New("403")
	case 503:
		// TODO
		return errors.New("503")
	}
	return errors.New("UNKNOWN")
}

func (s *FtsSession) PickPerson() error {
	data := map[string]interface{}{
		"email":     s.Profile.Email,
		"firstName": s.Profile.FirstName,
		"lastName":  s.Profile.LastName,
	}
	payloadBytes, _ := json.Marshal(data)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("https://www.%s.com/api/users/carts/current/pickperson?timestamp=%v", s.Site, time.Now().UnixNano()/int64(time.Millisecond)), bytes.NewReader(payloadBytes))
	req.Header.Set("X-Csrf-Token", s.CSRF)
	req.Header.Set("X-Fl-Productid", s.SizeID)
	req.Header.Set("X-Flapi-session-id", s.JSESSIONID)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.128 Safari/537.36")
	req.Header.Set("X-Fl-Request-Id", s.UUID)
	req.Header.Set("Origin", fmt.Sprintf("https://www.%s.com", s.Site))
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Referer", fmt.Sprintf("https://www.%ss.com/checkout", s.Site)) // Important
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Cookie", "JSESSIONID="+s.JSESSIONID+"; "+"datadome="+s.DataDome+"; "+"cart-guid="+s.CartGuid)
	resp, err := s.Client.Do(req)
	if err != nil {
		return err
	}
	ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200, 201:
		// SUCCESS
		for _, v := range resp.Cookies() {
			switch v.Name {
			case "datadome":
				s.DataDome = v.Value
			default:
				// Do Nothing
			}
		}
		return nil
	case 429:
		// TODO
		return errors.New("429")
	case 403:
		// TODO
		return errors.New("403")
	case 503:
		// TODO
		return errors.New("503")
	}
	return errors.New("UNKNOWN")
}

func (s *FtsSession) SubmitOrder() error {
	cc, en, ey, cvv, _ := a.EncryptCreditcardDetails(s.Profile.CC.CCNumber, s.Profile.CC.ExpMonth, s.Profile.CC.ExpYear, s.Profile.CC.Cvv)
	data := map[string]interface{}{
		"preferredLanguage":     "en",
		"termsAndCondition":     false,
		"deviceId":              "0400tyDoXSFjKeoNf94lis1ztioT9A1DShgAnrp/XmcfWoVVgr+Rt2dAZPhMS97Z4yfjSLOS3mruQCzk1eXuO7gGCUfgUZuLE2xCJiDbCfVZTGBk19tyNs7g8zV85QpvmtF/PiH81LzIHY89C7pjSl/JxUN13n2vmAeykQgdlVeDidx1G2mpGiKJ4Ao5VNMvaDXf7E1Pf46IvXtYdEyMOakFzprLKk3u1s0Iq1jEc21Hw6sowi9Jf88gkkjXzk77ILZ/eUsQ7RNrLro1kTKIs1496YkpIh3A707lm2e25SQbo1OuF8qR6VxrbC1wRHKPI15Qt45gqkMMYYfmY1XpDGBtyepPcth3j49FbUw/Y7k8g+pI+pjSFsqkUH6/N04I+t4CSViaqSM75eAyoDLQV8tjoO5wWTLXZT8tb6o9WQLwaJrb8RzWdzZK4bduoeAODFWWn+P0HNw3kOx04hiAcL/dBjV9nBJW1Y3JLACqjtcRnAvzJ1F0W5Ivre9RJsn4u3PdHf7WUtiodkTscXNLCbvmpFoNB0kg9fZ5lAWxTN4DFe0pQs/EGzmZL5NHY4PWj2UUP1K1FVKPNfm2EMlMtbRnNwQxSxzPuRtycb0H5IjHkCcKSs4KVYKB4i89PjJjT+QZ/Uoiy6yqSntHDo0vYYDC35oMiutO9ehTdLDs+1JC1RHBTwTVw5JVoFEa2lLiexN4OaKgIv3sbV/sTU9/joi9e1h0TIw5qQXOmssqTe7WzQirWMRzbUfDqyjCL0l/zyCSSNfOTvsgtn95SxDtE2suujWRMoizXj3piSkiHcDvTuWbZ7blJBujU64XypHpXGtsLXBEco8jXlC3jmCqQwxhh+ZjVekMYG3JYXyFHBbsmyyzb0vSwWRA+eSMeWgCk0M8kgQVobtOy5nu7pRmssxc1n3Rzh1ozHM1w5rNRhjMFvlSwejOX5dhKGYrsu13rc4RCbryu9G8AaTIauRtCQwBH44X08jLtlj16MHb+jKUckBfu00yz9/YQl2DyjiNQ5FMWmFLFbkufQOZu721S+SsYPbPQSWnaPfVhKRSjB4oZ28MlyI9q4+qN2hV2SY+Vz58fqYAndlMe7bZ0GxLd+c4lDeGG4xIicZNyoU7LduN6ZrcE4nZ3meVnnb6dxDioksejgtHHqqE28hv5t2asbd8NQZ+i8aFjp2buuK69Oy8ERmOBnl5tj9a8Yrr+l3H+CrCxGFawmFFcIAKFEZFQI9MzuB56CKvT5VTxedGLMijsMcQRH9L1eBfqylVWCPnhiZw",
		"cartId":                s.CartGuid,
		"encryptedCardNumber":   cc,
		"encryptedExpiryMonth":  en,
		"encryptedExpiryYear":   ey,
		"encryptedSecurityCode": cvv,
		"paymentMethod":         "CREDITCARD",
		"returnUrl":             "https://www.footlocker.com/adyen/checkout",
		"browserInfo": map[string]interface{}{
			"screenWidth":    1440,
			"screenHeight":   900,
			"colorDepth":     30,
			"userAgent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.128 Safari/537.36",
			"timeZoneOffset": 240,
			"language":       "en-US",
			"javaEnabled":    false,
		},
	}
	payloadBytes, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", fmt.Sprintf("https://www.%s.com/api/v2/users/orders?timestamp=%v", s.Site, time.Now().UnixNano()/int64(time.Millisecond)), bytes.NewReader(payloadBytes))
	req.Header.Set("X-Csrf-Token", s.CSRF)
	req.Header.Set("X-Fl-Productid", s.SizeID)
	req.Header.Set("X-Flapi-session-id", s.JSESSIONID)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.128 Safari/537.36")
	req.Header.Set("X-Fl-Request-Id", s.UUID)
	req.Header.Set("Origin", fmt.Sprintf("https://www.%s.com", s.Site))
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Referer", fmt.Sprintf("https://www.%ss.com/checkout", s.Site)) // Important
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Cookie", "JSESSIONID="+s.JSESSIONID+"; "+"datadome="+s.DataDome+"; "+"cart-guid="+s.CartGuid)
	resp, err := s.Client.Do(req)
	if err != nil {
		return err
	}
	ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200, 201:
		// SUCCESS
		for _, v := range resp.Cookies() {
			switch v.Name {
			case "datadome":
				s.DataDome = v.Value
			default:
				// Do Nothing
			}
		}
		return nil
	case 429:
		// TODO
		return errors.New("429")
	case 403:
		// TODO
		return errors.New("403")
	case 503:
		// TODO
		return errors.New("503")
	}
	return errors.New("Payment Decline")
}
