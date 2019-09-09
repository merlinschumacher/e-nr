package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/miekg/dns"
	"golang.org/x/net/idna"
)

//DNSRecord a struct that defines the way DNS entries work
type DNSRecord struct {
	ID          int
	Fullname    string
	URL         string
	Description string
	Dosage      string
}

//DNSRecords an array that holds all registered DNS entries as DNSEntry structs
var DNSRecords []DNSRecord

//global base domain
var baseDomain = ".e-nr.de."

//Our main function. This is where it all starts
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

	go func() {
		http.HandleFunc("/", handleHTTP)
		err := http.ListenAndServe(":9090", nil)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}()

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	s := <-sig
	fmt.Printf("Signal (%s) received, stopping\n", s)
}

// handleQuery - parse DNS requests and answer them
func handleQuery(writer dns.ResponseWriter, request *dns.Msg) {
	var recordData DNSRecord
	var message *dns.Msg

	for _, question := range request.Question {
		var err error
		recordData, err = findRecordDataByName(question.Name)
		if err != nil {
			message = buildResourceRecord(0, request, recordData)
		} else {
			message = buildResourceRecord(question.Qtype, request, recordData)
		}
	}
	writer.WriteMsg(message)
}

func buildResourceRecord(queryType uint16, request *dns.Msg, recordData DNSRecord) *dns.Msg {

	message := new(dns.Msg)
	message.SetReply(request)
	message.Compress = true
	strID := strconv.Itoa(recordData.ID)
	cnames := []string{recordData.Fullname, strID, "e" + strID, "e-" + strID}
	switch queryType {
	case dns.TypeTXT:
		dom := recordData.Fullname + baseDomain
		description, _ := idna.ToASCII(recordData.Fullname)
		record := &dns.TXT{
			Hdr: dns.RR_Header{Name: dom, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 0},
			// Txt: []string{recordData.Fullname, recordData.Description, recordData.URL},
			Txt: []string{description},
		}
		message.Answer = append(message.Answer, record)
		return message
	case dns.TypeA:
		for _, cname := range cnames {
			dom := cname + baseDomain
			record := &dns.A{
				Hdr: dns.RR_Header{Name: dom, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 0},
				A:   net.IPv4(127, 0, 0, 1),
			}
			message.Answer = append(message.Answer, record)
		}
		return message
	case dns.TypeCNAME:
		for _, cname := range cnames {
			dom := cname + baseDomain
			record := &dns.CNAME{
				Hdr:    dns.RR_Header{Name: dom, Rrtype: dns.TypeCNAME, Class: dns.ClassINET, Ttl: 0},
				Target: dom,
			}
			message.Answer = append(message.Answer, record)
		}
		return message
	case dns.TypeURI:
		for _, cname := range cnames {
			dom := cname + baseDomain
			record := &dns.URI{
				Hdr:    dns.RR_Header{Name: dom, Rrtype: dns.TypeURI, Class: dns.ClassINET, Ttl: 0},
				Target: recordData.URL,
			}
			message.Answer = append(message.Answer, record)
		}
		return message
	default:
		log.Printf("defaulting to NXDOMAIN for request\n %s", request.Question[0].String())
		dom := "ns" + baseDomain
		record := &dns.SOA{
			Hdr: dns.RR_Header{Name: dom, Rrtype: dns.TypeSOA, Class: dns.ClassINET, Ttl: 0},
			Ns:  dom,
		}

		message.SetRcode(request, dns.RcodeNameError)
		message.Answer = append(message.Answer, record)
		return message
	}
}

func findRecordDataByName(search string) (DNSRecord, error) {
	var emptyRecord DNSRecord
	emptyRecord.ID = -1
	isNumber := true
	search, _ = idna.ToUnicode(search)

	//parse the given
	numberRegex := regexp.MustCompile(`\d{3,4}`)
	numberStr := numberRegex.FindString(search)

	//check if the searched record is a number
	number, err := strconv.Atoi(string(numberStr))
	if err != nil || number == 0 {
		isNumber = false
		search = strings.ToLower(search)
		search = strings.ReplaceAll(search, ".", "")
	}
	for _, record := range DNSRecords {
		//dns records should be lowercase and be converted to ASCII/Punycode if necessary
		record.Fullname = strings.ToLower(record.Fullname)
		if isNumber {
			if record.ID == number {

				record.Fullname, err = idna.ToASCII(record.Fullname)
				return record, nil
			}
		} else {
			if strings.Contains(search, record.Fullname) {
				record.Fullname, err = idna.ToASCII(record.Fullname)
				return record, nil
			}
		}
	}
	return emptyRecord, errors.New("No record found for: " + search)
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
		tmpDNSRecord.URL = sanitizeURL(row[2], "https://de.wikipedia.org/wiki/")
		tmpDNSRecord.Description = row[3]
		tmpDNSRecord.Dosage = row[4]

		// append the temporary entry to our entries
		DNSRecords = append(DNSRecords, tmpDNSRecord)

		sort.Slice(DNSRecords, func(i, j int) bool {
			return DNSRecords[i].ID < DNSRecords[j].ID
		})

	}
}

func sanitizeURL(link string, prefix string) string {
	shortenedLink := strings.ReplaceAll(link, prefix, "")
	baseURL, err := url.Parse(shortenedLink)
	if err != nil {
		return ""
	}
	baseURL.Path = prefix + url.PathEscape(baseURL.Path)
	return baseURL.Path
}

func handleHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	hostname := strings.Split(request.Host, ".")[0]
	var record DNSRecord
	var err error
	path := strings.ReplaceAll(request.URL.Path, "/", "")
	// path, _ = idna.ToASCII(path)
	if path != "" {
		record, err = findRecordDataByName(path)
		log.Println(record)
	} else {
		record, err = findRecordDataByName(hostname)
	}
	if err != nil || record.ID == -1 {
		http.ServeFile(responseWriter, request, "index.html")
		return
	}
	http.Redirect(responseWriter, request, record.URL, 301)
}
