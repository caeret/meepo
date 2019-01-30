package meepo

import (
	"log"
	"net"
	"time"

	"github.com/miekg/dns"
)

func NewServer(target, trusted string, routes *Routes) *Server {
	return &Server{
		target:  target,
		trusted: trusted,
		routes:  routes,
		client:  new(dns.Client),
	}
}

type Server struct {
	target  string
	trusted string
	routes  *Routes
	client  *dns.Client
}

func (s *Server) Run(addr string) error {
	server := &dns.Server{
		Addr: addr,
		Net:  "udp",
	}
	server.Handler = s

	log.Printf(`listen on addr: "%s".`, addr)

	return server.ListenAndServe()
}

func (s *Server) ServeDNS(w dns.ResponseWriter, req *dns.Msg) {
	id := randStr(16)

	for _, ques := range req.Question {
		log.Printf(`[%s][in]class:"%s" type:"%s", name:"%s".`, id, dns.Class(ques.Qclass).String(), dns.Type(ques.Qtype).String(), ques.Name)
	}
	resp, t, err := s.exchange(req)
	if err != nil {
		log.Printf(`[%s]fail to exchange dns: "%v"".`, id, err)
		resp = new(dns.Msg)
	}

	log.Printf(`[%s][out] in %v.`, id, t)

	for _, ans := range resp.Answer {
		switch v := ans.(type) {
		case *dns.A:
			log.Printf(`[%s][out]name:"%s" a:"%s".`, id, v.Header().Name, v.A.String())
		default:
			log.Printf(`[%s][out]%v`, id, v)
		}
	}

	resp.SetReply(req)
	err = w.WriteMsg(resp)
	if err != nil {
		log.Printf(`[%s][out]fail to send response: "%v".`, id, err)
	}
}

func (s *Server) exchange(req *dns.Msg) (resp *dns.Msg, rtt time.Duration, err error) {
	var (
		targetResult  = make(chan ExchangeResult, 1)
		trustedResult = make(chan ExchangeResult, 1)
	)
	go func() {
		var result ExchangeResult
		result.Msg, result.RTT, result.err = s.client.Exchange(req, s.target)
		targetResult <- result
	}()
	go func() {
		var result ExchangeResult
		result.Msg, result.RTT, result.err = s.client.Exchange(req, s.trusted)
		trustedResult <- result
	}()

	select {
	case result := <-targetResult:
		log.Printf("from target.")
		if result.err != nil {
			err = result.err
			return
		}
		ip := s.findIP(result.Msg)
		if ip != nil && !s.routes.Test(ip) {
			log.Printf("from trusted.")
			result := <-trustedResult
			return result.Msg, result.RTT, result.err
		}
		return result.Msg, result.RTT, result.err
	}
}

func (s *Server) findIP(msg *dns.Msg) net.IP {
	for _, ans := range msg.Answer {
		if v, ok := ans.(*dns.A); ok {
			return v.A
		}
	}
	return nil
}

type ExchangeResult struct {
	Msg *dns.Msg
	RTT time.Duration
	err error
}
