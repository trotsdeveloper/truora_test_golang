package scrapers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	"github.com/trotsdeveloper/truora_test/truora_test_golang/dao"
	"golang.org/x/net/html"
)


func ScraperCountry(ip string) (country string, err error) {
	url := "https://geoipify.whoisxmlapi.com/api/v1?apiKey=at_7UiqEmpdxBJmZ9rQxIlkzACwNDiXA&ipAddress="+ip+"&outputFormat=json"
	var apiInfo map[string]interface{}

	r, err := http.Get(url)
	if err != nil {
		return
	}
	defer r.Body.Close()
	if err = json.NewDecoder(r.Body).Decode(&apiInfo); err != nil {
		return
	}

	location, ok := apiInfo["location"]
	if !ok {
		err = errors.New("Error loading data in GEOIPIFY")
		return
	}
	var v map[string]interface{}
	v = location.(map[string]interface{})
	country = v["country"].(string)
	return
}

func ScraperOwner(ip string) (owner string, err error) {
	url := "https://www.whoisxmlapi.com/whoisserver/WhoisService?apiKey=at_5UhpXqA9prtTSlHrPE2UJiUyASacC&domainName="+ip+"&outputFormat=json"
	var apiInfo map[string]interface{}

	r, err := http.Get(url)
	if err != nil {
		return
	}
	defer r.Body.Close()
	if err = json.NewDecoder(r.Body).Decode(&apiInfo); err != nil {
		return
	}
	whoisRecord, ok := apiInfo["WhoisRecord"]
	if !ok {
		err = errors.New("Error loading data in WHOISXMLAPI")
		return
	}
	var v map[string]interface{}
	v = whoisRecord.(map[string]interface{})

	registryData := v["registryData"].(map[string]interface{})
	registrant := registryData["registrant"].(map[string]interface{})
	owner = registrant["organization"].(string)

	return
}

func getHTMLinDomain(domain string) (htmlB []byte, err error) {
	var resp *http.Response
	resp, err = http.Get(domain)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	htmlB, err = ioutil.ReadAll(resp.Body)
	return
}

func ScraperLogo(domain string) (logo string, err error) {
	var htmlB []byte
	htmlB, err = getHTMLinDomain(fmt.Sprintf("http://%v/", domain))
	if err != nil {
		return
	}
	htmlS := string(htmlB)

	var doc *html.Node
	doc, err = html.Parse(strings.NewReader(htmlS))

	var f func(*html.Node) string
	f = func(n *html.Node) string {
		icon := "NO ICON"
		if n.Type == html.ElementNode && n.Data == "link" {
			iconInAttr := false
			for _, a := range n.Attr {
				if a.Key == "rel" && strings.Contains(a.Val,"icon") {
					iconInAttr = true
					break
				}
			}
			if iconInAttr {
				for _, a := range n.Attr {
					if a.Key == "href" {
						return a.Val
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			icon = f(c)
			if icon != "NO ICON" {
				break
			}
		}
		return icon
	}
	return f(doc), nil
}

func ScraperTitle(domain string) (s string, err error) {
	var htmlB []byte
	htmlB, err = getHTMLinDomain(fmt.Sprintf("http://%v/", domain))
	if err != nil {
		return
	}
	htmlS := string(htmlB)

	var doc *html.Node
	doc, err = html.Parse(strings.NewReader(htmlS))
	if err != nil {
		return
	}

	var f func(*html.Node) string
	f = func(n *html.Node) string {
		title := ""
		if n.Type == html.ElementNode && n.Data == "title" {
			return n.FirstChild.Data
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			title = f(c)
			if title != "" {
				break
			}
		}
		return title
	}
	fmt.Println(fmt.Sprintf("doc: %v", doc))
	return f(doc), nil
}

func ScraperSSLabs(currentHour time.Time, domain string) (se dao.ServerEvaluation, err error) {
	byt, err := getHTMLinDomain(fmt.Sprintf("https://api.ssllabs.com/api/v3/analyze?host=%v/", domain))
	if err != nil {
		return
	}

	var dat map[string]interface{}
	err = json.Unmarshal(byt, &dat)
	if err != nil {
		return
	}

	/* JSON TAP Mappings */
	status, ok := dat["status"]
	if !ok {
		err = errors.New("status TAG not present")
		return
	}

	// Assignation in server evaluation
	se.Domain = domain
	se.EvaluationHour = currentHour.Format(time.RFC3339)

	if status == "DNS" || status == "IN_PROGRESS" {
		se.EvaluationInProgress = true
	} else if status == "ERROR" {
		se.IsDown = true
	}

	servers := make([]dao.Server, 0)

	if !se.EvaluationInProgress && !se.IsDown {
		endpoints, ok := dat["endpoints"].([]interface{})
		if !ok {
			err = errors.New("endpoints TAG not present")
			se.Servers = servers
			return
		}
		califications := make(map[string]int)
		// A+, A-, A-F, T (no trust) and M
		califications["NaN"] = -1
		califications["M"] = 0
		califications["T"] = 1
		califications["F"] = 2
		califications["E"] = 3
		califications["D"] = 4
		califications["C"] = 5
		califications["B"] = 6
		califications["A"] = 7
		califications["A-"] = 8
		califications["A+"] = 9

		lowest := califications["A+"]
		lowestGrade := "A+"

		for _, v := range endpoints {
			castV := v.(map[string]interface{})
			server := dao.Server{}
			var ipI, gradeI interface{}
			var ok bool
			ipI, ok = castV["ipAddress"]
			if ok {
				server.Address = ipI.(string)
				gradeI, ok = castV["grade"]
				if ok {
					grade := gradeI.(string)
					server.SslGrade = grade
					servers = append(servers, server)
					if califications[grade] < lowest {
						lowest = califications[grade]
						lowestGrade = grade
					}
				} else {
					servers = append(servers, server)
					lowest = califications["NaN"]
					lowestGrade = "NaN"
				}
			}
		}
		se.SslGrade = lowestGrade
	}
	fmt.Println(fmt.Sprintf("What:%v",servers))
	se.Servers = servers


	return
}
