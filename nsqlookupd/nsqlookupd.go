package main

import (
	"../util"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
)

const VERSION = "0.1"

var (
	showVersion = flag.Bool("version", false, "print version string")
	bindAddress = flag.String("address", "0.0.0.0", "address to bind to")
	webPort     = flag.String("web-port", "5160", "port to listen on for HTTP connections")
	tcpPort     = flag.String("tcp-port", "5161", "port to listen on for TCP connections")
	debugMode   = flag.Bool("debug", false, "enable debug mode")
	cpuProfile  = flag.String("cpu-profile", "", "write cpu profile to file")
	goMaxProcs  = flag.Int("go-max-procs", 0, "runtime configuration for GOMAXPROCS")
)

var sm *util.SafeMap

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("nsqlookupd v%s\n", VERSION)
		return
	}

	if *goMaxProcs > 0 {
		runtime.GOMAXPROCS(*goMaxProcs)
	}

	endChan := make(chan int)
	signalChan := make(chan os.Signal, 1)
	sm = util.NewSafeMap()

	if *cpuProfile != "" {
		log.Printf("CPU Profiling Enabled")
		f, err := os.Create(*cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	go func() {
		<-signalChan
		endChan <- 1
	}()
	signal.Notify(signalChan, os.Interrupt)

	tcpAddr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(*bindAddress, *tcpPort))
	if err != nil {
		log.Fatal(err)
	}

	webAddr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(*bindAddress, *webPort))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("nsqlookupd v%s", VERSION)

	go TcpServer(tcpAddr)
	HttpServer(webAddr, endChan)
}
