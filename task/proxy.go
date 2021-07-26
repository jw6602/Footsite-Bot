package task

import (
	"errors"
	"strings"
)

type Proxy struct {
	IP   string
	Port string
	User string
	Pass string
}

func (p *Proxy) rawString() string {

	raw := p.IP + ":" + p.Port

	if p.User != "" && p.Pass != "" {
		raw = raw + ":" + p.User + ":" + p.Pass
	}

	return raw
}

func (p *Proxy) ConString() string {

	raw := p.IP + ":" + p.Port

	if p.User != "" && p.Pass != "" {
		raw = "http://" + p.User + ":" + p.Pass + "@" + raw
	}

	return raw
}

func stringToProxy(line string) (Proxy, error) {

	parts := strings.Split(line, ":")

	if len(parts) == 2 {
		return Proxy{parts[0], parts[1], "", ""}, nil

	} else if len(parts) == 4 { 
		return Proxy{parts[0], parts[1], parts[2], parts[3]}, nil

	} else { 
		return Proxy{"", "", "", ""}, errors.New("Error parsing proxy")
	}
}

func LoadProxies(proxyStr string) ([]Proxy, error) {
	proxyArr := strings.Split(proxyStr, "\n")
	var proxies []Proxy
	for _, v := range proxyArr {

		proxy, err := stringToProxy(v)

		if err == nil {
			proxies = append(proxies, proxy)
		}
	}

	return proxies, nil
}
