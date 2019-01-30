package meepo

import (
	"log"
	"net"
	"os"
	"time"

	"github.com/miekg/dns"
)

func NewServer(internal, trusted string, routes *Routes) *Server {
	return &Server{
		internal: internal,
		trusted:  trusted,
		routes:   routes,
		client:   new(dns.Client),
		logger: log.New(os.Stderr, "[meepo]", log.LstdFlags),
	}
}

type Server struct {
	internal string
	trusted  string
	routes   *Routes
	client   *dns.Client
	logger Logger
}

func (s *Server) SetLogger(logger Logger) {
	if logger == nil {
		s.logger = nopLogger{}
		return
	}
	s.logger = logger
}

func (s *Server) Run(addr string) error {
	server := &dns.Server{
		Addr: addr,
		Net:  "udp",
	}
	server.Handler = s

	s.logger.Printf(`listen on addr: "%s".`, addr)

	return server.ListenAndServe()
}

func (s *Server) ServeDNS(w dns.ResponseWriter, req *dns.Msg) {
	for _, ques := range req.Question {
		s.logger.Printf(`-> class:"%s" type:"%s", name:"%s".`, dns.Class(ques.Qclass).String(), dns.Type(ques.Qtype).String(), ques.Name)
	}
	resp, t, err := s.transfer(req)
	if err != nil {
		s.logger.Printf("<- fail to transfer dns: %v.", err)
		resp = new(dns.Msg)
	}

	s.logger.Printf(`<- in %v.`, t)

	for _, ans := range resp.Answer {
		switch v := ans.(type) {
		case *dns.A:
			s.logger.Printf(`<- name:"%s" a:"%s".`, v.Header().Name, v.A.String())
		default:
			s.logger.Printf("<- %v.", v)
		}
	}

	resp.SetReply(req)
	err = w.WriteMsg(resp)
	if err != nil {
		s.logger.Printf("<- fail to write reply: %v.", err)
	}
}

func (s *Server) transfer(req *dns.Msg) (resp *dns.Msg, rtt time.Duration, err error) {
	var (
		internal = make(chan ExchangeResult, 2)
		trusted  = make(chan ExchangeResult, 1)
	)
	go func() {
		s.logger.Printf("<-> used time %v @%s.", s.timing(func() {
			s.exchange(req, s.internal, internal)
		}), s.internal)
	}()
	go func() {
		s.logger.Printf("<-> used time %v @%s.", s.timing(func() {
			s.exchange(req, s.trusted, trusted)
		}), s.trusted)
	}()

	select {
	case result := <-internal:
		s.logger.Printf("<- read msg from \"%s\".", s.internal)
		if result.err != nil {
			err = result.err
			return
		}
		ip := s.findIP(result.msg)
		s.logger.Printf("<- find ip: %s.", ip)
		if ip == nil || !s.routes.Test(ip) {
			s.logger.Printf("<- ip is nil or ip is not in routes, read msg from \"%s\".", s.trusted)
			result := <-trusted
			return result.msg, result.rtt, result.err
		}
		return result.msg, result.rtt, result.err
	}
}

func (s *Server) exchange(req *dns.Msg, addr string, ch chan<- ExchangeResult)  {
	var result ExchangeResult
	s.logger.Printf("-> send msg %s.", addr)
	result.msg, result.rtt, result.err = s.client.Exchange(req, addr)
	s.logger.Printf("<- receive msg %s.", addr)
	ch <- result
}

func (s *Server) findIP(msg *dns.Msg) net.IP {
	for _, ans := range msg.Answer {
		if v, ok := ans.(*dns.A); ok {
			return v.A
		}
	}
	return nil
}

func (s *Server) timing(fn func()) time.Duration {
	t := time.Now()
	fn()
	return time.Now().Sub(t)
}

type ExchangeResult struct {
	msg      *dns.Msg
	rtt      time.Duration
	err      error
}
