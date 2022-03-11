// Written by Paul Schou (paulschou.com) March 2022
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Repomd struct {
	XMLName  xml.Name `xml:"repomd"`
	Text     string   `xml:",chardata"`
	Xmlns    string   `xml:"xmlns,attr"`
	Rpm      string   `xml:"rpm,attr"`
	Revision string   `xml:"revision"`
	Data     []struct {
		Text     string `xml:",chardata"`
		Type     string `xml:"type,attr"`
		Checksum struct {
			Text string `xml:",chardata"`
			Type string `xml:"type,attr"`
		} `xml:"checksum"`
		Location struct {
			Text string `xml:",chardata"`
			Href string `xml:"href,attr"`
		} `xml:"location"`
		Timestamp    int `xml:"timestamp"`
		Size         int `xml:"size"`
		OpenChecksum struct {
			Text string `xml:",chardata"`
			Type string `xml:"type,attr"`
		} `xml:"open-checksum"`
		OpenSize        string `xml:"open-size"`
		DatabaseVersion string `xml:"database_version"`
	} `xml:"data"`
	fileContents    []byte
	ascFileContents string
	path            string
	mirror          string
}

var client = http.Client{
	Timeout: 5 * time.Second,
}

func readRepomdFile(repomdFile string) *Repomd {
	// Declare file handle for the reading
	var file io.Reader

	if _, err := os.Stat(repomdFile); err == nil {
		log.Println("Reading in file", repomdFile)

		// Open our xmlFile
		rawFile, err := os.Open(repomdFile)
		if err != nil {
			log.Println("Error in HTTP get request", err)
			return nil
		}

		// Make sure the file is closed at the end of the function
		defer rawFile.Close()
		file = rawFile
	} else if strings.HasPrefix(repomdFile, "http") {
		resp, err := client.Get(repomdFile)
		if err != nil {
			log.Println("Error in HTTP get request", err)
			return nil
		}

		defer resp.Body.Close()
		file = resp.Body
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(file)
	contents := buf.Bytes()

	var dat Repomd
	err := xml.Unmarshal(contents, &dat)
	if err != nil {
		log.Println("Error in decoding Repomd", err)
		return nil
	}
	dat.fileContents = contents

	return &dat
}

func readWithChecksum(fileName, checksum, checksumType string) *[]byte {
	// Declare file handle for the reading
	var file io.Reader

	if _, err := os.Stat(fileName); err == nil {
		log.Println("Reading in file", fileName)

		// Open our xmlFile
		rawFile, err := os.Open(fileName)
		if err != nil {
			log.Println("Error in opening file locally", err)
			return nil
		}

		// Make sure the file is closed at the end of the function
		defer rawFile.Close()
		file = rawFile
	} else if strings.HasPrefix(fileName, "http") {
		resp, err := client.Get(fileName)
		if err != nil {
			log.Println("Error in HTTP get request", err)
			return nil
		}

		defer resp.Body.Close()
		file = resp.Body
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(file)
	contents := buf.Bytes()
	var sum string

	switch checksumType {
	case "sha256":
		sum = fmt.Sprintf("%x", sha256.Sum256(contents))
	}

	if sum == checksum {
		return &contents
	}
	return nil
}
