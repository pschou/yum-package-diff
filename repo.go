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
	"bufio"
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"path"
)

type Matchable interface {
	matches(in Matchable) bool
	size() string
	print(out io.Writer, repoPath string)
}

func open(fileName string) (file io.Reader, closure func()) {
	log.Println("Reading in file", fileName)

	// Open our xmlFile
	rawFile, err := os.Open(fileName)
	check(err)

	// Declare file handle for the reading
	file = rawFile

	// Detect magic number
	b1 := make([]byte, 4)
	n1, err := file.Read(b1)
	_, err = rawFile.Seek(0, 0)

	if n1 == 4 && string(b1) == "\x1f\x8b\x08\x00" {
		//fmt.Println("Using gz decoder")
		check(err)

		gz, err := gzip.NewReader(rawFile)
		check(err)

		// Make sure gzip handle is closed at the end of the function
		closure = func() {
			gz.Close()
			rawFile.Close()
		}

		// Substitute the gz reader in the place of file to handle the compressed file
		file = gz
	} else {
		bufReader := bufio.NewReaderSize(file, 100000)
		file = bufReader
		// Make sure the file is closed
		closure = func() {
			rawFile.Close()
		}
	}
	return
}

type Package struct {
	XMLName xml.Name `xml:"package"`
	//Text     string   `xml:",chardata"`
	//Type string `xml:"type,attr"`
	//Name     string `xml:"name"`
	Checksum struct {
		Text string `xml:",chardata"`
		Type string `xml:"type,attr"`
		//Pkgid string `xml:"pkgid,attr"`
	} `xml:"checksum"`
	Size struct {
		//Text      string `xml:",chardata"`
		Package string `xml:"package,attr"`
		//Installed string `xml:"installed,attr"`
		//Archive   string `xml:"archive,attr"`
	} `xml:"size"`
	Location struct {
		//Text string `xml:",chardata"`
		Href string `xml:"href,attr"`
	} `xml:"location"`
	//Time struct {
	//	File  float64 `xml:"file,attr"`
	//	Build float64 `xml:"build,attr"`
	//} `xml:"time"`
}

func (p1 Package) matches(in Matchable) bool {
	if p2, ok := in.(Package); ok {
		return p1.Checksum.Text == p2.Checksum.Text &&
			p1.Checksum.Type == p2.Checksum.Type &&
			p1.Size.Package == p2.Size.Package &&
			p1.Location.Href == p2.Location.Href
	}
	return false
}
func (p Package) size() string { return p.Size.Package }
func (p Package) print(out io.Writer, repoPath string) {
	fmt.Fprintf(out, "{%s}%s %s %s\n", p.Checksum.Type, p.Checksum.Text, p.Size.Package, path.Join(repoPath, p.Location.Href))
}

type PackageMetadata struct {
	XMLName xml.Name `xml:"metadata"`
	//Text        string    `xml:",chardata"`
	//Xmlns       string    `xml:"xmlns,attr"`
	//Rpm         string    `xml:"rpm,attr"`
	Packages    int       `xml:"packages,attr"`
	PackageList []Package `xml:"package"`
}

func readFile(fileName string) []Matchable {
	file, closure := open(fileName)
	if file == nil {
		return nil
	}
	defer closure()
	decoder := xml.NewDecoder(file)
	var dat PackageMetadata
	err := decoder.Decode(&dat)
	check(err)

	if len(dat.PackageList) == 0 {
		log.Fatal("No packages found")
	}
	if len(dat.PackageList) != dat.Packages {
		log.Fatal("XML Packages count does not match the number of Packages")
	}
	m := make([]Matchable, len(dat.PackageList))
	for i, v := range dat.PackageList {
		m[i] = v
	}
	return m
}

type DeltaPackage struct {
	XMLName xml.Name `xml:"newpackage"`
	Text    string   `xml:",chardata"`
	Name    string   `xml:"name,attr"`
	Epoch   string   `xml:"epoch,attr"`
	Version string   `xml:"version,attr"`
	Release string   `xml:"release,attr"`
	Arch    string   `xml:"arch,attr"`
	Delta   struct {
		Text       string `xml:",chardata"`
		Oldepoch   string `xml:"oldepoch,attr"`
		Oldversion string `xml:"oldversion,attr"`
		Oldrelease string `xml:"oldrelease,attr"`
		Filename   string `xml:"filename"`
		Sequence   string `xml:"sequence"`
		Size       string `xml:"size"`
		Checksum   struct {
			Text string `xml:",chardata"`
			Type string `xml:"type,attr"`
		} `xml:"checksum"`
	} `xml:"delta"`
}

func (p1 DeltaPackage) matches(in Matchable) bool {
	if p2, ok := in.(DeltaPackage); ok {
		return p1.Name == p2.Name &&
			p1.Version == p2.Version &&
			p1.Release == p2.Release &&
			p1.Delta.Oldversion == p2.Delta.Oldversion &&
			p1.Delta.Oldrelease == p2.Delta.Oldrelease &&
			p1.Delta.Size == p2.Delta.Size &&
			p1.Delta.Checksum.Text == p2.Delta.Checksum.Text
	}
	return false
}
func (p DeltaPackage) size() string { return p.Delta.Size }
func (p DeltaPackage) print(out io.Writer, repoPath string) {
	fmt.Fprintf(out, "{%s}%s %s %s\n", p.Delta.Checksum.Type, p.Delta.Checksum.Text, p.Delta.Size, path.Join(repoPath, p.Delta.Filename))
}

type DeltaPackageMetadata struct {
	XMLName     xml.Name       `xml:"prestodelta"`
	PackageList []DeltaPackage `xml:"newpackage"`
}

func readDeltaFile(fileName string) []Matchable {
	file, closure := open(fileName)
	if file == nil {
		return nil
	}
	defer closure()
	decoder := xml.NewDecoder(file)
	var dat DeltaPackageMetadata
	err := decoder.Decode(&dat)
	check(err)

	if len(dat.PackageList) == 0 {
		log.Fatal("No packages found")
	}
	m := make([]Matchable, len(dat.PackageList))
	for i, v := range dat.PackageList {
		m[i] = v
	}
	return m
}
