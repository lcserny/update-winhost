package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const (
	RESOLVCONF_PATH = "/etc/resolv.conf"
	HOSTS_PATH = "/etc/hosts"
	HOSTNAME   = "winhost"
)

var ipPattern = regexp.MustCompile(`(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})`)

func main() {
	resolvConfFile := openFile(RESOLVCONF_PATH, os.O_RDONLY)
	defer resolvConfFile.Close()

	var winIP string
	resolvConfScanner := bufio.NewScanner(resolvConfFile)
	for resolvConfScanner.Scan() {
		line := resolvConfScanner.Text()
		if strings.Contains(line, "nameserver") {
			winIP = ipPattern.FindString(line)
			break
		}
	}
	if err := resolvConfScanner.Err(); err != nil {
		panic(err)
	}

	hostsFile := openFile(HOSTS_PATH, os.O_RDWR)
	defer hostsFile.Close()

	var hostsLines []string
	var hostExists bool
	hostsScanner := bufio.NewScanner(hostsFile)
	for hostsScanner.Scan() {
		line := hostsScanner.Text()
		if strings.Contains(line, HOSTNAME) {
			currentWinIP := ipPattern.FindString(line)
			if currentWinIP != winIP {
				line = fmt.Sprintf("%s\t%s", winIP, HOSTNAME)
			}
			hostExists = true
		}
		hostsLines = append(hostsLines, line)
	}
	if err := hostsScanner.Err(); err != nil {
		panic(err)
	}

	if !hostExists {
		hostsLines = append(hostsLines, fmt.Sprintf("%s\t%s", winIP, HOSTNAME))
	}

	if err := hostsFile.Truncate(0); err != nil {
		panic(err)
	}
	hostsFile.Seek(0, 0)

	hostsWriter := bufio.NewWriter(hostsFile)
	for _, line := range hostsLines {
		hostsWriter.WriteString(line + "\n")
	}
	hostsWriter.Flush()
}

func openFile(path string, perms int) *os.File {
	f, err := os.OpenFile(path, perms, 0644)
	if err != nil {
		panic(err)
	}
	return f
}
