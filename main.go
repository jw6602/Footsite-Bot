package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/TwinProduction/go-color"
	"github.com/jw6602/footsite-bot/task"
)

var (
	wg sync.WaitGroup
)

func main() {
	proxyStr, _ := ioutil.ReadFile("proxy.txt")
	proxies, _ := task.LoadProxies(string(proxyStr))
	profileStr, _ := ioutil.ReadFile("profile.json")
	profiles, _ := task.LoadProfile(profileStr)
	fmt.Print(color.Green)
	fmt.Printf("%v proxies have been loaded.\n", len(proxies))
	fmt.Printf("%v profiles have been loaded.\n", len(profiles))
	fmt.Print(color.Reset)
	fmt.Println("Enter the sites you want to run: ")
	fmt.Println("[1] Footlocker [2] EastBay [3] Champs [4] Footaction [5] Kidsfootlocker")
	var siteNum int
	fmt.Scanf("%d\n", &siteNum)
	var site string
	switch siteNum {
	case 1:
		site = "footlocker"
	case 2:
		site = "eastbay"
	case 3:
		site = "champssports"
	case 4:
		site = "footaction"
	case 5:
		site = "kidsfootlocker"
	}
	var sku string
	fmt.Println("Enter your sku here: ")
	fmt.Scanln(&sku)
	var sizes string
	fmt.Println("Enter your size here: (eg: 07.5,08.0,08.5,...)")
	fmt.Scanln(&sizes)
	sizeRange := strings.Split(sizes, ",")
	fmt.Println()
	fmt.Print(color.Green)
	fmt.Println("All Tasks will be started after three seconds. :)")
	fmt.Print(color.Reset)
	fmt.Println()
	time.Sleep(3 * time.Second)
	logger := log.New(os.Stdout, "", 0)
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			s := task.FtsSession{}
			rand.Seed(time.Now().UnixNano())
			proxy := proxies[rand.Intn(len(proxies))]
			rand.Seed(time.Now().UnixNano())
			size := sizeRange[rand.Intn(len(sizeRange))]
			rand.Seed(time.Now().UnixNano())
			profile := profiles[rand.Intn(len(profiles))]
			s.InitSession(site, sku, size, proxy.ConString(), profile)
			for {
				logger.Printf(" [%s] [%s, sku=%s] Genereating Session\n", s.UUID, s.Site, s.SKU)
				if s.GenerateSession() == nil {
					break
				}
				time.Sleep(2 * time.Second)
				rand.Seed(time.Now().UnixNano())
				// Rotate Proxies
				s.ProxyURL, _ = url.Parse(proxies[rand.Intn(len(proxies))].ConString())
				logger.Printf(" [%s] [%s, sku=%s] Rotating Proxies\n", s.UUID, s.Site, s.SKU)
			}
			for {
				if s.GetSizeID() == nil {
					break
				}
				time.Sleep(1 * time.Second)
			}

			for {
				logger.Printf(" [%s] [%s, sku=%s, size=%s] Adding To Cart\n", s.UUID, s.Site, s.SKU, s.Size)
				err := s.AddToCart()
				if err == nil {
					break
				}
				if err.Error() == "DONE" {
					logger.Printf(" [%s] [%s, sku=%s, size=%s] Captcha Token Received\n", s.UUID, s.Site, s.SKU, s.Size)
				}
				if err.Error() == "FATAL" {
					wg.Done()
					return
				}
				time.Sleep(3 * time.Second)
			}
			logger.Printf("%s [%s] [%s, sku=%s, size=%s] ATC Success %s\n", color.Green, s.UUID, s.Site, s.SKU, s.Size, color.Reset)
			for {
				logger.Printf(" [%s] [%s, sku=%s, size=%s] Submitting Shipping Step.1\n", s.UUID, s.Site, s.SKU, s.Size)
				if s.SubmitEmail() == nil {
					break
				}
				time.Sleep(2 * time.Second)
			}

			for {
				logger.Printf(" [%s] [%s, sku=%s, size=%s] Submitting Shipping Step.2\n", s.UUID, s.Site, s.SKU, s.Size)
				if s.SubmitShipping() == nil {
					break
				}
				time.Sleep(1 * time.Second)
			}

			for {
				logger.Printf(" [%s] [%s, sku=%s, size=%s] Submitting Billing Step.1\n", s.UUID, s.Site, s.SKU, s.Size)
				if s.SubmitBilling() == nil {
					break
				}
				time.Sleep(1 * time.Second)
			}

			for {
				logger.Printf(" [%s] [%s, sku=%s, size=%s] Submitting Billing Step.2\n", s.UUID, s.Site, s.SKU, s.Size)
				if s.PickPerson() == nil {
					break
				}
				time.Sleep(1 * time.Second)
			}

			for i := 0; i < 10; i++ {
				logger.Printf(" [%s] [%s, sku=%s, size=%s] Submitting Order\n", s.UUID, s.Site, s.SKU, s.Size)
				if s.SubmitOrder() == nil {
					logger.Printf("%s [%s] [%s, sku=%s, size=%s] Check Email%s\n", color.Green, s.UUID, s.Site, s.SKU, s.Size, color.Reset)
					go task.SendSuccessWebhook(strings.Title(s.Site), s.SKU, s.Size, "your_webhook_url_here")
					wg.Done()
					return
				} else {
					logger.Printf("%s [%s] [%s, sku=%s, size=%s] Payment Decline%s\n", color.Red, s.UUID, s.Site, s.SKU, s.Size, color.Reset)
				}
			}
			go task.SendDeclineWebhook(strings.Title(s.Site), s.SKU, s.Size, "your_webhook_url_here")
			wg.Done()
		}()
	}
	wg.Wait()
}
