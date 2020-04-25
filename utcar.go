package main

import (
	"bytes"
	"expvar"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"time"
)

type SIA struct {
	time     time.Time
	sequence string
	receiver string
	line     string
	account  string
	command  string
	zone     string
}

type Heartbeat struct {
	time time.Time
}

var (
	ftaddr string
	ftuser string
	ftpwd  string
	fport  int
	fdebug int

	pchan chan SIA

	requests = expvar.NewInt("requests")
)

// init function.  Used to read input parameters to the program.
func init() {
	flag.StringVar(&ftaddr, "taddr", "", "Target addr (e.g. http://openhab.local:8080)")
	flag.StringVar(&ftuser, "tuser", "", "Target username")
	flag.StringVar(&ftpwd, "tpwd", "", "Target password")
	flag.IntVar(&fport, "port", 12300, "Listen port number (default: 12300)")
	flag.IntVar(&fdebug, "debug", 0, "Debug server port number (default: no debug server)")
	flag.Parse()
}

// handleConnection handles connections from the alarm system.
// In short, it accepts a connection and sends a new, encrypted key.  Then it
// receives an encrypted message from the alarm system, after which it completes
// with an ACK message.
func handleConnection(c net.Conn, q chan SIA) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Message processing panic (%v)\n", r)
			debug.PrintStack()
		}
	}()
	key := GenerateKey()
	scrambled_key := Scramble(key)
	// Send key to alarm system
	n, err := c.Write(scrambled_key)
	if err != nil {
		log.Panic(err)
	}

	buf := make([]byte, 1024) // receive buffer
	n, err = c.Read(buf)
	if err != nil {
		if err != io.EOF {
			log.Panic("Read error: ", err)
		}
	}
	encryptedData := buf[:n]

	data := Decrypt3DESECB(encryptedData, key)
	// Remove leading/trailing new line, line feeds, NUL chars
	data = bytes.Trim(data, "\n\r\x00")
	log.Println("Message: ", string(data[:]))

	ack := []byte("ACK\r")
	ack = append(ack, []byte{0, 0, 0, 0}...)
	encryptedAck := Encrypt3DESECB(ack, key)
	n, err = c.Write(encryptedAck)
	if err != nil {
		log.Panic(err)
	}

	if IsHeartbeat(data) {
		log.Println("Heartbeat.")
		return // don't know what to do with this yet.
	}
	parsed, err := ParseSIA(data)
	if err != nil {
		log.Panicf("Not a recognized message: %s", string(data[:]))
	}
	sia := SIA{time.Now(), parsed[0], parsed[1], parsed[2], parsed[3], parsed[4], parsed[5]}
	log.Println(sia)

	requests.Add(1) // accessible through expvar

	if q == nil {
		return
	} else {
		q <- sia
	}
}

func receiveSignal() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		<-sig
		os.Exit(0)
	}()
}

func main() {
	// setup response to CTRL-C
	receiveSignal()
	// Listen on TCP port 12300 on all interfaces
	l, err := net.Listen("tcp", ":"+strconv.Itoa(fport))
	if err != nil {
		log.Fatal(err) // exit.. something serious must be wrong.
	}
	log.Printf("Listing on port %d...", fport)
	defer l.Close()

	// setup debug server
	if fdebug != 0 {
		go func() {
			err = http.ListenAndServe(":"+strconv.Itoa(fdebug), nil)
		}()
		if err != nil {
			log.Printf("Failed to start debug server (%v)\n", err)
		} else {
			log.Printf("Debug server running on port %d\n", fdebug)
		}
	}

	// setup pusher channel (if addr is provided)
	if ftaddr != "" {
		log.Printf("Pushing to %s\n", ftaddr)
		pchan = make(chan SIA)
		go func() {
			for {
				sia := <-pchan
				// TODO: handle panics from this function (if any?)
				err := HttpPost(ftaddr, ftuser, ftpwd, sia)
				if err != nil {
					log.Printf("Push error: %v", err)
				}
			}
		}()
	}

	for { // eternally...
		// Wait for a connection
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		// Handle the connection in a new routine
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.
		go func(c net.Conn) {
			defer c.Close()

			handleConnection(c, pchan)
		}(conn)
	}
}
