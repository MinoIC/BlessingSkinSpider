package main

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Item struct {
	Tid      uint   `json:"tid"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Hash     string `json:"hash"`
	Size     uint   `json:"size"`
	Uploader uint   `json:"uploader"`
	Public   bool   `json:"public"`
	UploadAt string `json:"upload_at"`
	Likes    uint   `json:"likes"`
}

type Decoder struct {
	Data struct {
		Items []Item `json:"items"`
	} `json:"data"`
}

const (
	/*
		request per second smaller than 1000
		a high rps may make you fail
	*/
	rps = 10
	/*
		target host name such as https://skin.example.com
		dont put / as suffix !!!!
	*/
	target = "https://littleskin.cn"
	/* spider type
	"skin": all the skins
	"steve": steve skins
	"alex": alex skins
	"cape": all the capes
	*/
	filter = "skin"
	/* uploader uid set to 0 as default to get all uploader s' skin*/
	uploader = 0
	/* how many pages you want to get [1-pages] */
	pages = 35
	/* save path */
	path = `E:\download`
)

var (
	count   = 0
	success = 0
	failed  = 0
)

func main() {
	/* to get list pages */
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"),
	)
	c.OnRequest(
		func(request *colly.Request) {
			fmt.Println("visiting:", request.URL)
		},
	)
	c.OnResponse(
		func(response *colly.Response) {
			//fmt.Println("visited: ", response.Request.URL)
			//fmt.Println("response: ", string(response.Body))
			if response.StatusCode == 200 {
				var dec Decoder
				if err := json.Unmarshal(response.Body, &dec); err != nil {
					fmt.Println("err:", err)
				} else {
					for _, item := range dec.Data.Items {
						time.Sleep((1000 / rps) * time.Millisecond)
						go func(item *Item) {
							imgSrc := target + "/textures/" + item.Hash
							img, _ := http.Get(imgSrc)
							f, err := os.Create(path + `\` + item.Name + strconv.Itoa(int(item.Tid)) + `.png`)
							if err != nil {
								fmt.Println(err)
							} else {
								_, _ = io.Copy(f, img.Body)
							}
							_ = f.Close()
							fmt.Println("name: "+item.Name+"count: ", count, "status: ", img.StatusCode)
							if img.StatusCode == 200 {
								success = success + 1
							} else {
								failed = failed + 1
							}
							count = count + 1
						}(&item)
					}
				}
			}
		},
	)
	start := time.Now()
	for i := 1; i <= pages; i++ {
		err := c.Visit(target + "/skinlib/data?filter=" + filter + "&uploader=" + strconv.Itoa(uploader) + "&sort=time&keyword=&page=" + strconv.Itoa(i))
		if err != nil {
			fmt.Println(err)
		}
	}
	dur := time.Now().Sub(start).String()
	fmt.Println("finished! SUCCESS: ", success, "FAILED: ", failed, "TOTAL: ", count, "(", dur, ")")
}
