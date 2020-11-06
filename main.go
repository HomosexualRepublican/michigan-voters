package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func getVoterStatus(firstName string, lastName string, birthYear string, zipCode string) (string, int, error) {
	validResponse := ""
	birthMonth := -1
	for i := 1; i < 13; i++ {
		time.Sleep(50 * time.Millisecond)
		resp, err := http.PostForm("https://mvic.sos.state.mi.us/Voter/SearchByName",
			url.Values{"FirstName": {firstName}, "LastName": {lastName},
				"NameBirthMonth": {fmt.Sprintf("%d", i)}, "NameBirthYear": {birthYear}, "ZipCode": {zipCode},
				"Din": {""}, "DinBirthMonth": {"0"}, "DinBirthYear": {""}, "DpaID": {"0"},
				"Months": {""}, "VoterNotFound": {"false"}, "TransitionVoter": {"false"}})
		if err != nil {
			return "", -1, err
		}
		defer resp.Body.Close()
		respBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", -1, err
		}
		if bytes.Contains(respBytes, []byte("No voter record matched your search criteria")) {
			continue
		}
		validResponse = string(respBytes)
		birthMonth = i
	}
	return validResponse, birthMonth, nil
}

func main() {
	csvFile, err := os.Open("detroit_index.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer csvFile.Close()
	r := csv.NewReader(csvFile)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		resp, month, err := getVoterStatus(record[0], record[1], record[2], record[3])
		if err != nil {
			log.Fatal(err)
		}
		if strings.Contains(resp, "Ballot received") {
			log.Printf("%s, %s, %s, %s, %d, RECEIVED", record[0], record[1], record[2], record[3], month)
		} else {
			log.Printf("%s, %s, %s, %s, %d, NORETURN", record[0], record[1], record[2], record[3], month)
		}
	}
}
