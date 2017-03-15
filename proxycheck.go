package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/abh/geoip"
	"github.com/parnurzeal/gorequest"
)

type getWithproxy struct {
	proxy     string
	url       string
	fileout   string
	newstring string
	info      bool
}

func appendStringToFile(path, text string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(text)
	if err != nil {
		return err
	}
	return nil
}

func (g *getWithproxy) getproxy() {
	httpProxy := fmt.Sprintf("http://%s", g.proxy)
	s := strings.Split(g.proxy, ":")
	ip := s[0]
	b, _ := ioutil.ReadFile(g.fileout)
	str := string(b)
	existStr := strings.Contains(str, ip)

	if existStr == false {
		request := gorequest.New().Proxy(httpProxy).Timeout(2 * time.Second)
		timeStart := time.Now()
		_, _, err := request.Get(g.url).End()
		if err != nil {
			fmt.Println("BAD: ", g.proxy)
		} else {
			fmt.Println("GOOD: ", g.proxy)
			if g.info == true {
				country := ipToCountry(ip)
				respone := time.Since(timeStart)
				g.newstring = fmt.Sprintf("%s;%s;%s\n", g.proxy, country, respone)
			} else {
				g.newstring = fmt.Sprintf("%s\n", g.proxy)
			}
			appendStringToFile(g.fileout, g.newstring)
		}
	}
}

func ipToCountry(ip string) string {
	file := "/usr/share/GeoIP/GeoIP.dat"

	gi, err := geoip.Open(file)
	if err != nil {
		fmt.Printf("Could not open GeoIP database\n")
		os.Exit(1)
	}
	country, _ := gi.GetCountry(ip)
	return country
}

func main() {
	var (
		url     = flag.String("url", "https://m.vk.com", "")
		fileIn  = flag.String("in", "proxylist.txt", "full path to proxy file")
		fileOut = flag.String("out", "goodlist.txt", "full path to output file")
		info    = flag.Bool("info", false, "info about proxy: Country, Respone")
		treds   = flag.Int("treds", 50, "number of treds")
	)

	flag.Parse()

	content, _ := ioutil.ReadFile(*fileIn)
	proxys := strings.Split(string(content), "\n")

	workers := *treds

	wg := new(sync.WaitGroup)
	in := make(chan string, 2*workers)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for proxy := range in {
				gp := getWithproxy{
					proxy:   proxy,
					url:     *url,
					fileout: *fileOut,
					info:    *info,
				}
				gp.getproxy()
			}
		}()
	}

	for _, proxy := range proxys {
		if proxy != "" {
			in <- proxy
		}
	}
	close(in)
	wg.Wait()
}