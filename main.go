package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/miekg/dns"
)

//DNSEntry - a struct that defines the way DNS entries work
type DNSRecord struct {
	ID          int
	Fullname    string
	URL         string
	Description string
	Dosage      string
}

//DNSentries - an array that holds all registered DNS entries as DNSEntry structs
var DNSRecords []DNSRecord

const dom = "whoami.miek.nl."

//Out main function. This is where it all starts
func main() {
	loadCSV("dns.csv")
	dns.HandleFunc(".", handleQuery)

	go func() {
		server := &dns.Server{Addr: ":8053", Net: "udp"}
		err := server.ListenAndServe()
		if err != nil {
			log.Fatalf("Could not setup udp listender %s", err.Error())
		}
	}()

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	s := <-sig
	fmt.Printf("Signal (%s) received, stopping\n", s)
}

// handleQuery - parse DNS requests and answer them
func handleQuery(writer dns.ResponseWriter, request *dns.Msg) {
	message := new(dns.Msg)
	message.SetReply(request)
	message.Compress = true
	for _, question := range request.Question {
		recordData := findRecordDataByName(question.Name)
		switch question.Qtype {
		case dns.TypeTXT:
			record := &dns.TXT{
				Hdr: dns.RR_Header{Name: dom, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 0},
				Txt: []string{recordData.Fullname, recordData.Description, recordData.URL},
			}
			message.Answer = append(message.Answer, record)
		case dns.TypeA:
			record := &dns.A{
				Hdr: dns.RR_Header{Name: dom, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 0},
				A:   net.IPv4(127, 0, 0, 1),
			}
			message.Answer = append(message.Answer, record)
		default:
			message.SetRcode(request, dns.RcodeNameError)

		}
	}
	writer.WriteMsg(message)
}

func findRecordDataByName(name string) DNSRecord {
	isNumber := true
	numberRegex := regexp.MustCompile(`\d{3,4}`)
	numberStr := numberRegex.FindString(name)
	number, err := strconv.Atoi(string(numberStr))
	if err != nil || number == 0 {
		isNumber = false
	}
	for _, record := range DNSRecords {
		if isNumber {
			if record.ID == number {
				return record
			}
		} else {
			if strings.Contains(record.Fullname, name) {
				return record
			}
		}
	}
	var emptyRecord DNSRecord
	return emptyRecord
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
		var tmpDNSRecord DNSRecord
		tmpDNSRecord.ID, err = strconv.Atoi(row[0])
		if err != nil {
			tmpDNSRecord.ID = 0
		}
		tmpDNSRecord.Fullname = row[1]
		tmpDNSRecord.URL = row[2]
		tmpDNSRecord.Description = row[3]
		tmpDNSRecord.Dosage = row[4]

		// append the temporary entry to our entries
		DNSRecords = append(DNSRecords, tmpDNSRecord)

		sort.Slice(DNSRecords, func(i, j int) bool {
			return DNSRecords[i].ID < DNSRecords[j].ID
		})

	}
}
