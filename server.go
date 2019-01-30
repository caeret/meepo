package meepo

import (
	"log"

	"github.com/miekg/dns"
)

func NewServer(target string) *Server {
	return &Server{
		target: target,
		client: new(dns.Client),
	}
}

type Server struct {
	target string
	client *dns.Client
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
	resp, t, err := s.client.Exchange(req, s.target)
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
