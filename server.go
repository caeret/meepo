package meepo

import (
	"github.com/miekg/dns"
	"log"
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

func (s *Server) Run(addr string) {
	server := &dns.Server{
		Addr:addr,
		Net:"udp",
	}
	server.Handler = s
	panic(server.ListenAndServe())
}

func (s *Server) ServeDNS(w dns.ResponseWriter, req *dns.Msg) {
	resp, _, err := s.client.Exchange(req, s.target)
	if err != nil {
		log.Printf("fail to exchange dns: %v.", err)
		resp = new(dns.Msg)
	}
	resp.SetReply(req)
	err = w.WriteMsg(resp)
	if err != nil {
		log.Printf("fail to send response: %v.", err)
	}
}