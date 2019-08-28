package main

import (
	"encoding/csv"
	"io"
	"log"
	"net"
	"os"

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
		_, clientAddr, err := server.ReadFromUDP(buf)
		if err != nil {
			log.Printf("Error reading UDP data: %s", err.Error())
			continue
		}
		packet := gopacket.NewPacket(buf, layers.LayerTypeDNS, gopacket.Default)
		dnsPacket := packet.Layer(layers.LayerTypeDNS)
		request, _ := dnsPacket.(*layers.DNS)

		// handle the clients request in a thread
		go handleQuery(server, clientAddr, request)
	}
}

// handleQuery - parse DNS requests and answer them
func handleQuery(server *net.UDPConn, clientAddr net.Addr, request *layers.DNS) {
	for i, question := range request.Questions {
		var answer layers.DNSResourceRecord
		answer.Type = layers.DNSTypeA
		answer.Name = []byte(question.Name)
		ip, _, err := net.ParseCIDR("192.168.0.1/24")
		if err != nil {
			log.Println("Error converting the stored IP")
		}
		answer.IP = ip
		answer.Class = layers.DNSClassIN
		request.QR = true
		request.ANCount = uint16(i)
		request.OpCode = layers.DNSOpCodeNotify
		request.AA = true
		request.Answers = append(request.Answers, answer)
		request.ResponseCode = layers.DNSResponseCodeNoErr

	}
	buf := gopacket.NewSerializeBuffer()
	opt := gopacket.SerializeOptions{}
	err := request.SerializeTo(buf, opt)
	if err != nil {
		log.Printf("Error serializing answers %s", err.Error())
	}
	server.WriteTo(buf.Bytes(), clientAddr)

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
