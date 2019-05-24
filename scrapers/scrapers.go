package scrapers

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/likexian/whois-go"
	"github.com/trotsdeveloper/truora_test/truora_test_golang/dao"
	"golang.org/x/net/html"
)

// ServerEvaluation ...

func ScraperCountry(ip string) (s string, err error) {
	var result string
	result, err = whois.Whois(ip)

	if err != nil {
		return
	}

	re, _ := regexp.Compile(`Country:(.*)`)
	submatch := re.FindStringSubmatch(result) // First submatch

	if len(submatch) == 0 {
		//fmt.Println(submatch)
		err = errors.New("No country in WHOIS info.")
		return ``, err
	}
	s = strings.TrimSpace(submatch[1])
	return s, err
}

func ScraperOwner(ip string) (s string, err error) {
	var result string
	result, err = whois.Whois(ip)

	if err != nil {
		return
	}

	re, _ := regexp.Compile(`OrgName:(.*)`)
	submatch := re.FindStringSubmatch(result) // First submatch

	if len(submatch) == 0 {
		//fmt.Println(submatch)
		err = errors.New("No server owner in WHOIS info.")
		return ``, err
	}
	s = strings.TrimSpace(submatch[1])
	return s, err
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
	htmlB, err = getHTMLinDomain(domain)
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
				if a.Key == "type" && a.Val == "image/x-icon" {
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
	htmlB, err = getHTMLinDomain(domain)
	if err != nil {
		return
	}
	htmlS := string(htmlB)

	var doc *html.Node
	doc, err = html.Parse(strings.NewReader(htmlS))

	var f func(*html.Node) string
	f = func(n *html.Node) string {
		var title string
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
	return f(doc), nil
}

func ScraperSSLabs(domain string) (se dao.ServerEvaluation, err error) {
	byt, err := getHTMLinDomain(domain)
	if err != nil {
		return
	}

	var dat map[string]interface{}
	err = json.Unmarshal(byt, &dat)
	if err != nil {
		return
	}

	/* JSON TAP Mappings */
	status, ok := dat["status"].(string)
	if !ok {
		return se, errors.New("status TAG not present")
	}
	endpoints, ok := dat["endpoints"].([]map[string]interface{})
	if !ok {
		return se, errors.New("endpoints TAG not present")
	}

	// Assignation in server evaluation
	se = dao.ServerEvaluation{}

	se.Domain = domain
	t := time.Now()
	se.EvaluationHour = t.Format(time.RFC3339)

	if status == "DNS" || status == "IN_PROGRESS" {
		se.EvaluationInProgress = true
	} else if status == "ERROR" {
		se.IsDown = true
	}

	servers := make([]dao.Server, 0)
	if !se.EvaluationInProgress && !se.IsDown {
		califications := make(map[string]int)
		// A+, A-, A-F, T (no trust) and M
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
		lowestGrade := ""

		for _, v := range endpoints {
			server := dao.Server{}
			ip := v["ipAddress"].(string)
			grade := v["grade"].(string)
			server.Address = ip
			server.SslGrade = grade
			servers = append(servers, server)
			if califications[grade] < lowest {
				lowest = califications[grade]
				lowestGrade = grade
			}
		}
		se.SslGrade = lowestGrade
	}

	se.Servers = servers
	return
}
