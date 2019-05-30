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
	fmt.Println(apiInfo)

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

func ScraperSSLabs(currentHour time.Time, domain string) (se dao.ServerEvaluation, err error) {
	byt, err := getHTMLinDomain(fmt.Sprintf("https://api.ssllabs.com/api/v3/analyze?host=%v/", domain))
	if err != nil {
		fmt.Println(fmt.Sprintf("err: %v", err))
		return
	}

	var dat map[string]interface{}
	err = json.Unmarshal(byt, &dat)
	if err != nil {
		fmt.Println(fmt.Sprintf("err: %v", err))
		return
	}

	/* JSON TAP Mappings */
	status, ok := dat["status"]
	if !ok {
		err = errors.New("status TAG not present")
		fmt.Println(fmt.Sprintf("err: %v", err))
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
			return
		}
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
		lowestGrade := "A+"

		for _, v := range endpoints {
			castV := v.(map[string]interface{})
			server := dao.Server{}
			ip := castV["ipAddress"].(string)
			grade := castV["grade"].(string)
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

var APIErrors = newAPIErrorsRegistry()

func newAPIErrorsRegistry() *apiErrorsRegistry {
	e601v := makeAPIError("601", "Error in database.")
	e602v := makeAPIError("602", "Error in SSLabs API.")
	e701v := makeAPIError("701", "Error getting Icon")
	e702v := makeAPIError("702", "Error getting HTML Title")
	e801v := makeAPIError("801", "Error getting country from WHOIS")
	e802v := makeAPIError("802", "Error getting owner from WHOIS")

	return &apiErrorsRegistry{
		e601: e601v,
		e602: e602v,
		e701: e701v,
		e702: e702v,
		e801: e801v,
		e802: e802v,
	}
}

func makeAPIError(code string, description string) func(string) (APIError) {
	return func(err string) (APIError) {
		return APIError{code, description, err}
	}
}

type apiErrorsRegistry struct {
	e601 func(string) (APIError) //
	e602 func(string) (APIError) //
	e701 func(string) (APIError) //
	e702 func(string) (APIError) //
	e801 func(string) (APIError) //
	e802 func(string) (APIError) //
}

type APIError struct {
	code string
	description string
	err string
}

/*
e601v := makeAPIError("601", "Error in database.")
e602v := makeAPIError("602", "Error in SSLabs API.")
e701v := makeAPIError("701", "Error getting Icon")
e702v := makeAPIError("702", "Error getting HTML Title")
e801v := makeAPIError("801", "Error getting country from WHOIS")
e802v := makeAPIError("802", "Error getting owner from WHOIS")
*/
func ScraperTestComplete(domain string, currentHour time.Time, dbc interface{}) (sec dao.ServerEvaluationComplete, appErr []APIError) {

	sec = dao.ServerEvaluationComplete{}
	appErr := make([]APIError, 0)

	var se dao.ServerEvaluation
	var err error
	se, err = dao.MakeEvaluationInDomain(domain, currentHour, ScraperSSLabs, dbc)

	if err != nil {
		appErr = append(appErr, APIErrors.e601(err))	
		return
	}

	sec.Copy(se)
	if !se.IsDown {
		sec.Logo, err = ScraperLogo(domain)
		if err != nil {
			appErr = append(appErr, APIErrors.e701(err))
		}
		sec.Title, err = ScraperTitle(domain)
		if err != nil {
			appErr = append(appErr, APIErrors.e702(err))
		}
	}

	if !se.EvaluationInProgress && !se.IsDown {
		for i := range sec.Servers {
			ip := sec.Servers[i].Address
			sec.Servers[i].Country, err = ScraperCountry(ip)
			if err != nil {
				appErr = append(appErr, APIErrors.e801)
			}
			sec.Servers[i].Owner, err = ScraperOwner(ip)
			if err != nil {
				appErr = append(appErr, APIErrors.e802)
			}
		}
		var serversChangedI int
		serversChangedI, err = se.HaveServersChanged(dbc)
		if err != nil {
			fmt.Println(fmt.Sprintf("err: %v", err))
			return
		}
		sec.ServersChanged = (serversChangedI == dao.SLStatus.Changed)
		sec.PreviousSslGrade, err = se.PreviousSSLgrade(dbc)
		if err != nil {
			fmt.Println(fmt.Sprintf("err: %v", err))
		}
	}

	return
}
