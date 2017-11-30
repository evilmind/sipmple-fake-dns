package main

import (
	"fmt"
	"log"
	"github.com/miekg/dns"
	"strings"
	"regexp"
	"os/signal"
	"os"
)

func isValidIP(host string) bool {
	if strings.Contains(host, ".") {
		Re := regexp.MustCompile(`(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)`)
		return Re.MatchString(host)
	} else if strings.Contains(host, ":") {
		Re := regexp.MustCompile(`(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))`)
		return Re.MatchString(host)
	} else {
		return false
	}
}

func getIPFromEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		if isValidIP(value) {
			return value
		} else {
			log.Fatalf("Invalid IP")
		}
	}
	return fallback
	
}

var (
	ipv4 = getIPFromEnv("DEFAULT_IPV4", "127.0.0.1")
	ipv6 = getIPFromEnv("DEFAULT_IPV6", "::1")
)

func parseQuery(m *dns.Msg) {
	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeA:
			log.Printf("Query A record for %s\n", q.Name)
			rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ipv4))
				if err == nil {
					m.Answer = append(m.Answer, rr)
				}
		case dns.TypeAAAA:
			log.Printf("Query AAAA record for %s\n", q.Name)
			rr, err := dns.NewRR(fmt.Sprintf("%s AAAA %s", q.Name, ipv6))
			if err == nil {
				m.Answer = append(m.Answer, rr)
				}
			}
		}
	}

func handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case dns.OpcodeQuery:
		parseQuery(m)
	}
	w.WriteMsg(m)
}

func main() {
	dns.HandleFunc(".", handleDnsRequest)
	go serve("udp", ":53")
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, os.Kill)
	for {
		select {
		case s := <-sig:
			log.Fatalf("fatal: signal %s received\n", s)
		}
	}
}

func serve(net, addr string) {
	server := &dns.Server{Addr: addr, Net: net, TsigSecret: nil}
	log.Printf("Starting at %s\n", addr)
	err := server.ListenAndServe()
	
	if err != nil {
		log.Fatalf("Failed to setup the %s server: %v\n", net, err)
	}
}
