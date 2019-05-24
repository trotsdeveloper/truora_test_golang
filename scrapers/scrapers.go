package scrapers

import (
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/likexian/whois-go"
	"github.com/trotsdeveloper/truora_test/truora_test_golang/dao"
	"golang.org/x/net/html"
)

func EvaluateServerCountry(domain string) (s string, err error) {
	var result string
	result, err = whois.Whois(domain)

	if err != nil {
		return
	}

	re, _ := regexp.Compile(`Country:(.*)`)
	submatch := re.FindStringSubmatch(result) // First submatch

	if len(submatch) == 0 {
		//fmt.Println(submatch)
		err = &dao.CustomError{"No country in WHOIS info."}
		return ``, err
	}
	s = strings.TrimSpace(submatch[1])
	return s, err
}

func EvaluateServerOwner(domain string) (s string, err error) {
	var result string
	result, err = whois.Whois(domain)

	if err != nil {
		return
	}

	re, _ := regexp.Compile(`OrgName:(.*)`)
	submatch := re.FindStringSubmatch(result) // First submatch

	if len(submatch) == 0 {
		//fmt.Println(submatch)
		err = &dao.CustomError{"No server owner in WHOIS info."}
		return ``, err
	}
	s = strings.TrimSpace(submatch[1])
	return s, err
}

func getHTMLinDomain(domain string) (html string, err error) {
	var resp *http.Response
	resp, err = http.Get(domain)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var htmlBytes []byte
	htmlBytes, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	html = string(htmlBytes)
	return
}

func EvaluateLogoInDomain(domain string) (logo string, err error) {
	var htmlS string
	htmlS, err = getHTMLinDomain(domain)
	if err != nil {
		return
	}
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

func EvaluateTitleInDomain(domain string) (s string, err error) {
	var htmlS string
	htmlS, err = getHTMLinDomain(domain)
	if err != nil {
		return
	}
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
