package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

type ProductPrice struct {
	ProductName string
	PrductPrice float64
	CheckedIn   time.Time
	Website     string
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func convertStringRupeesFloat(Price string) (float64, error) {
	Price = strings.TrimSpace(Price)
	Price = strings.Replace(Price, "₹", "", -1)
	Price = strings.Replace(Price, ",", "", -1)
	return strconv.ParseFloat(Price, 64)
}

func checkAmazon(url string, userAgent string, webSite string, ch chan ProductPrice) {
	if url == "" {
		return
	}
	time.Sleep(2 * time.Second)
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorln(err)
		return
	}
	req.Header.Set("User-Agent", userAgent)

	res, err := client.Do(req)
	if err != nil {
		log.Errorln(err)
		return
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Errorln("status code error: %d %s", res.StatusCode, res.Status)
		return
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Errorln(err)
		return
	}

	ProductName := strings.TrimSpace(doc.Find("span#productTitle").Text())
	TextPrice := doc.Find("span.a-offscreen").First().Text()
	fmt.Println(TextPrice)
	fmt.Println(ProductName)
	Price, err := convertStringRupeesFloat(TextPrice)
	if err != nil {
		log.Errorln(err)
		return
	}
	tempObj := ProductPrice{
		ProductName: ProductName,
		PrductPrice: Price,
		CheckedIn:   time.Now(),
		Website:     webSite,
	}

	time.Sleep(2 * time.Second)
	ch <- tempObj
}

func checkFlipKart(url string, userAgent string, webSite string, ch chan ProductPrice) {
	if url == "" {
		return
	}
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorln(err)
		return
	}
	req.Header.Set("User-Agent", userAgent)

	res, err := client.Do(req)
	if err != nil {
		log.Errorln(err)
		return
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Errorln("status code error: %d %s", res.StatusCode, res.Status)
		return
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Errorln(err)
		return
	}

	ProductName := strings.TrimSpace(doc.Find("span.B_NuCI").First().Text())
	TextPrice := doc.Find("div._30jeq3._16Jk6d").First().Text()
	fmt.Println(TextPrice)
	fmt.Println(ProductName)
	Price, err := convertStringRupeesFloat(TextPrice)
	if err != nil {
		log.Errorln(err)
		return
	}
	tempObj := ProductPrice{
		ProductName: ProductName,
		PrductPrice: Price,
		CheckedIn:   time.Now(),
		Website:     webSite,
	}

	time.Sleep(2 * time.Second)
	ch <- tempObj
}

func fileRead(fileName string) []string {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
		return []string{}
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	b, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
		return []string{}
	}
	s := string(b[:])
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "\n")

	split_url := strings.Split(s, "\n")
	return split_url
}

func getRandomAgent() string {
	userAgents := []string{
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/603.3.8 (KHTML, like Gecko) Version/10.1.2 Safari/603.3.8",
		"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Safari/537.36",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 10_3_2 like Mac OS X) AppleWebKit/603.2.4 (KHTML, like Gecko) Version/10.0 Mobile/14F89 Safari/602.1",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 10_3_2 like Mac OS X) AppleWebKit/603.2.4 (KHTML, like Gecko) FxiOS/8.1.1b4948 Mobile/14F89 Safari/603.2.4",
		"Mozilla/5.0 (iPad; CPU OS 10_3_2 like Mac OS X) AppleWebKit/603.2.4 (KHTML, like Gecko) Version/10.0 Mobile/14F89 Safari/602.1",
		"Mozilla/5.0 (Linux; Android 4.3; GT-I9300 Build/JSS15J) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.125 Mobile Safari/537.36",
		"Mozilla/5.0 (Android 4.3; Mobile; rv:54.0) Gecko/54.0 Firefox/54.0",
		"Mozilla/5.0 (Linux; Android 4.3; GT-I9300 Build/JSS15J) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.91 Mobile Safari/537.36 OPR/42.9.2246.119956",
		"Opera/9.80 (Android; Opera Mini/28.0.2254/66.318; U; en) Presto/2.12.423 Version/12.16",
	}

	rand.Seed(time.Now().Unix())
	return userAgents[rand.Intn(len(userAgents))]
}

func init() {
	DbInit()
	// The API for setting attributes is a little different than the package level
	// exported logger. See Godoc.
	log.Out = os.Stdout
	file, err := os.OpenFile("logrus.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.Out = file
	} else {
		log.Info("Failed to log to file, using default stderr")
	}
}

func process() {

}

func Pusher_Push(product_val ProductPrice) {
	new_product_name := strings.TrimSpace(product_val.ProductName)
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		log.Errorln(err)
		return
	}
	new_product_name = reg.ReplaceAllString(new_product_name, "")

	if len(new_product_name) > 50 {
		new_product_name = new_product_name[:50]
		new_product_name = strings.Trim(new_product_name, "_")
	}

	completionTime := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: product_val.Website,
		Subsystem: fmt.Sprintf("db_%s_%s_info", product_val.Website, new_product_name),
		Name:      fmt.Sprintf("db_%s_info", product_val.Website),
		Help:      fmt.Sprintf("Price of product %s.", product_val.ProductName),
	})
	prometheus.MustRegister(completionTime)
	completionTime.Set(product_val.PrductPrice)
	if err := push.New("http://192.168.0.101:9091/", fmt.Sprintf("%s_product_price", product_val.Website)).
		Collector(completionTime).
		Grouping(product_val.Website, new_product_name).
		Push(); err != nil {
		log.Error("Could not push completion time to Pushgateway:", err)
	}
}
func main() {
	amazoneUrl := fileRead("amazon_product_url.txt")
	flipkartUrl := fileRead("flipkart_product_url.txt")

	ch := make(chan ProductPrice, len(amazoneUrl)+len(flipkartUrl))
	for _, i := range amazoneUrl {
		i = strings.Replace(i, "\r", "", -1)
		checkAmazon(i, "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Safari/537.36", "amazon", ch)
	}

	for _, i := range flipkartUrl {
		i = strings.Replace(i, "\r", "", -1)
		checkFlipKart(i, "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Safari/537.36", "flipkart", ch)
	}

	for len(ch) > 0 {
		temp := <-ch
		fmt.Println(temp)
		Insert(temp)
		Pusher_Push(temp)
	}
}
