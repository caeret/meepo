package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/caeret/meepo"
	"github.com/go-ini/ini"
)

var (
	config string
)

func init() {
	flag.StringVar(&config, "config", "meepo.ini", "ini config file.")
	flag.Parse()
}

func main() {
	logger := log.New(os.Stderr, "[meepo]", log.LstdFlags)
	cfg, err := ini.Load(config)
	if err != nil {
		printf("fail to read config file: %s.", err)
		os.Exit(1)
	}
	routeFile := cfg.Section("").Key("route-file").Value()
	logger.Printf("try to read routes from file: %s.", routeFile)
	file, err := os.Open(routeFile)
	if err != nil {
		printf("fail to read routes: %s.", err)
		os.Exit(1)
	}
	defer file.Close()
	routes := meepo.NewRoutes()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		route := scanner.Text()
		err = routes.Add(route)
		if err != nil {
			printf("fail to add routes: %s.", route)
			os.Exit(1)
		}
	}

	internal := cfg.Section("server").Key("internal").Value()
	trusted := cfg.Section("server").Key("trusted").Value()
	if internal == "" || trusted == "" {
		printf("internal and trusted server cannot be empty.")
		os.Exit(1)
	}

	addr := cfg.Section("").Key("addr").Value()
	if addr == "" {
		printf("dns listen addr required.")
		os.Exit(1)
	}

	server := meepo.NewServer(internal, trusted, routes)
	server.SetLogger(logger)
	err = server.Run(addr)
	if err != nil {
		logger.Panic(err)
	}
}

func printf(format string, args ...interface{}) {
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	fmt.Fprintf(os.Stderr, format, args...)
}
