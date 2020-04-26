package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// ReadConfiguration reads a property file
func ReadConfiguration(filename string) (map[string]string, error) {
	config := map[string]string{}

	if len(filename) == 0 {
		return config, nil
	}

	file, err := os.Open(filename)
	if err != nil {
		SendException(err)
		time.Sleep(30 * time.Second)
		fmt.Printf("%s", err.Error())
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		currentLine := scanner.Text()
		if equalIndex := strings.Index(currentLine, "="); equalIndex >= 0 {
			if key := strings.TrimSpace(currentLine[:equalIndex]); len(key) > 0 {
				value := ""
				if len(currentLine) > equalIndex {
					value = strings.TrimSpace(currentLine[equalIndex+1:])
				}
				config[key] = value
			}
		}
	}

	if err := scanner.Err(); err != nil {
		SendException(err)
		time.Sleep(30 * time.Second)
		log.Fatalf("%s", err.Error())
		return nil, err
	}

	return config, nil
}

// CreateHTTPClient used to create the client for sending post requests to OMSEndpoint
func CreateHTTPClient() {
	cert, err := tls.LoadX509KeyPair(PluginConfiguration["cert_file_path"], PluginConfiguration["key_file_path"])
	if err != nil {
		message := fmt.Sprintf("Error when loading cert %s", err.Error())
		SendException(message)
		time.Sleep(30 * time.Second)
		Log(message)
		log.Fatalf("Error when loading cert %s", err.Error())
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	tlsConfig.BuildNameToCertificate()

	var proxyConfig *url.URL
	if _, err := os.Stat(PluginConfiguration["omsproxy_conf_path"]); err == nil {
		omsproxyConf, err := ioutil.ReadFile(PluginConfiguration["omsproxy_conf_path"])
		if err != nil {
			message := fmt.Sprintf("Error Reading omsproxy configuration %s\n", err.Error())
			Log(message)
			SendException(message)
			time.Sleep(30 * time.Second)
			log.Fatalln(message)
		} else {
			proxyEndpoint := strings.TrimSpace(string(omsproxyConf))
			proxyEndpointUrl, err := url.Parse(proxyEndpoint)	
			if err != nil {
				message := fmt.Sprintf("Error parsing omsproxy url %s\n", err.Error())
				Log(message)
				SendException(message)
				time.Sleep(30 * time.Second)
				log.Fatalln(message)
			} else {		
				Log("ProxyEndpoint %s", proxyEndpointUrl)
				proxyConfig = http.ProxyURL(proxyEndpointUrl)
		   }
		}
	}

	transport := &http.Transport{TLSClientConfig: tlsConfig, Proxy: proxyConfig}

	HTTPClient = http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	Log("Successfully created HTTP Client")
}

// ToString converts an interface into a string
func ToString(s interface{}) string {
	switch t := s.(type) {
	case []byte:
		// prevent encoding to base64
		return string(t)
	default:
		return ""
	}
}
