package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/tidwall/gjson"
)

var startTime time.Time
var msg string
var number int

type Model struct {
	model  string `json:"model"`
	title  string `json:"title"`
	mem    string `json:"mem"`
	disk   string `json:"disk"`
	price  string `json:"price"`
	status string `json:"status"`
	url    string `json:"url"`
}

func main() {
	startTime = time.Now()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	http.HandleFunc("/", home)

	log.Println("Server running on " + addr)
	go func() { log.Fatal(http.ListenAndServe(addr, nil)) }()

	for {
		go checkInventory()
		go keepAwake()
		time.Sleep(1 * time.Minute)
	}

}

func home(w http.ResponseWriter, r *http.Request) {
	upTime := time.Since(startTime)
	w.Write([]byte("Running " + upTime.String() + "\n"))
	w.Write([]byte("Last Checked " + strconv.Itoa(number) + "\n"))
	w.Write([]byte("Last Message " + msg + "\n"))
}

func keepAwake() {
	appURL := "https://YOUR_LINK.herokuapp.com/"
	http.Get(appURL)
}

func checkInventory() {
	// upTime := time.Since(startTime)
	req()

}

func notify(msg string) {
	botURL := "https://api.telegram.org/BOT_ID:BOT_KEY/sendMessage"
	msgJson := []byte(`{"chat_id":"CHAT_ID", "text":"` + msg + `"}`)
	client := &http.Client{
		Timeout: time.Second * 5,
	}
	req, _ := http.NewRequest("POST", botURL, bytes.NewBuffer(msgJson))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Print(err)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	log.Println(string(body))
	defer resp.Body.Close()
}

func req() {
	url := "https://www.apple.com.cn/shop/refurbished/mac/2020-macbook-air-macbook-pro-16gb"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		log.Println(err)
		return
	}
	cookies := os.Getenv("COOKIES")
	if cookies == "" {
		cookies = "YOUR_COOKIES"
	}
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Pragma", "no-cache")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("sec-ch-ua", "\"Google Chrome\";v=\"89\", \"Chromium\";v=\"89\", \";Not A Brand\";v=\"99\"")
	req.Header.Add("sec-ch-ua-mobile", "?0")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_2_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.82 Safari/537.36")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Add("Sec-Fetch-Site", "none")
	req.Header.Add("Sec-Fetch-Mode", "navigate")
	req.Header.Add("Sec-Fetch-User", "?1")
	req.Header.Add("Sec-Fetch-Dest", "document")
	req.Header.Add("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7,zh-TW;q=0.6")
	req.Header.Add("Cookie", cookies)

	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer res.Body.Close()

	// body, err := ioutil.ReadAll(res.Body)
	// fmt.Println(string(body))
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Println(err)
		return
	}
	// doc.Find("#refurbished-category-grid div div.as-gridpage.as-gridpage-pagination-hidden div.as-gridpage-pane div.as-gridpage-results ul").Each(func(i int, s *goquery.Selection) {
	// For each item found, get the band and title
	// price := s.Find(".as-price-currentprice").Text()
	// title := s.Find(".as-producttile-title a").Text()
	// fmt.Printf("Review %d: %s - %s\n", i, price, title)
	// })

	models := []Model{}

	available := false

	log.Println("Page loaded")
	number = 0
	doc.Find("div[role=main]>script").Each(func(i int, s *goquery.Selection) {
		data := strings.ReplaceAll(s.Text(), "window.REFURB_GRID_BOOTSTRAP =", "")
		products := gjson.Get(data, "tiles")
		products.ForEach(func(key, value gjson.Result) bool {
			number = number + 1
			m := Model{
				model:  value.Get("partNumber").String(),
				title:  value.Get("title").String(),
				mem:    value.Get("filters.dimensions.tsMemorySize").String(),
				disk:   value.Get("filters.dimensions.dimensionCapacity").String(),
				price:  value.Get("price.seoPrice").String(),
				status: value.Get("omnitureModel.customerCommitString").String(),
				url:    "https://www.apple.com.cn" + value.Get("productDetailsUrl").String(),
			}

			if strings.Contains(m.title, "M1") {
				// && m.mem == "16gb"
				// && m.disk == "512gb" {
				available = true

				log.Println(string((m.title + " " + m.price)))
				models = append(models, m)
			}

			if available == true && len(models) > 0 {
				info, err := json.Marshal(models)
				msg = string(info)

				if err != nil {
					log.Println(err)
				}
				notify(msg)
			}

			return true // keep iterating
		})

	})
}
