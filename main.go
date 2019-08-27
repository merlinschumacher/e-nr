package main

import (
	"encoding/csv"
	"io"
	"log"
	"net"
	"os"

	"golang.org/x/net/dns/dnsmessage"
	"github.com/google/gopacket"
	layers "github.com/google/gopacket/layers"
)

//DNSEntry - a struct that defines the way DNS entries work
type DNSEntry struct {
	Shortname   string
	Fullname    string
	URL         string
	Description string
	Dosage      string
}

//DNSentries an array that holds all registered DNS entries as DNSEntry structs
var DNSentries []DNSEntry

var server *net.UDPConn

//Out main function. This is where it all starts
func main() {
	loadCSV("dns.csv")
	var err error
	server, err = net.ListenUDP("udp", &net.UDPAddr{Port: 8053})

	if err != nil {
		log.Fatalf("Error resolving UDP address: %s", err.Error())
		os.Exit(1)
	}

	defer server.Close()

	for {
		buf := make([]byte, 512)
		_, addr, err := server.ReadFromUDP(buf)
		if err != nil {
			log.Printf("Error reading UDP request: %s", err.Error())
			continue
		}
		var message dnsmessage.Message
		err = message.Unpack(buf)
		if err != nil {
			log.Printf("Invalid DNS request: %s", err.Error())
		}
		// Skip this loop if there are no queries
		if len(message.Questions) == 0 {
			continue
		}
		// handle the clients request in a thread
		go handleQuery(message)
	}
}

// handleQuery - parse DNS requests and answer them
func handleQuery(message dnsmessage.Message) {
	for _, request := range message.Questions {
		recordname := request.Name
		recordtype := request.Type
		log.Println(recordname)
		log.Println(recordtype)
		var dnsAnswer layers.DNSResourceRecord
		dnsAnswer.Type = layers.DNSTypeA

		go server.WriteToUDP(b []byte, addr *net.UDPAddr)
	}

}

//loadCSV - this functions loads all dns entries from a file identified by the given filename
func loadCSV(filename string) {
	// open the file and check for errors
	csvFile, err := os.Open(filename)
	if err != nil {
		log.Fatal("Could not load given CSV.")
	}
	// create a reader to walk the files contents line by line
	reader := csv.NewReader(csvFile)
	for {
		// a line is loaded and stored in row. If the load fails check if the file has ended (io.EOF) or anything else has failed.
		row, err := reader.Read()
		if err == io.EOF {
			log.Println("Loaded all entries.")
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		//create a temporary DNS entry and fill it with the loaded data
		var tmpDNSEntry DNSEntry
		tmpDNSEntry.Shortname = row[0]
		tmpDNSEntry.Fullname = row[1]
		tmpDNSEntry.URL = row[2]
		tmpDNSEntry.Description = row[3]
		tmpDNSEntry.Dosage = row[4]

		// append the temporary entry to our entries
		DNSentries = append(DNSentries, tmpDNSEntry)
	}
}
