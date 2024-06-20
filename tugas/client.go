package main

import (
	"bufio"
	"bytes"
	
	"crypto/tls"
	"crypto/x509"

	"encoding/json"
	"fmt"
	"io"
	"main/data"
	"main/tools"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
)


func waitServer(url string, duration time.Duration, client *http.Client) bool {
	deadline := time.Now().Add(duration)
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		tools.ErrorHandler(err)
		if resp != nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return false
		}
	}
	return true
}


func main() {
	
	certPool := x509.NewCertPool()
	certData, err := os.ReadFile("./cert/cert.pem")
	if err != nil {
		fmt.Print(err)
	}
	if !certPool.AppendCertsFromPEM(certData) {
		fmt.Errorf("failed to append certificate")
	}
	tools.ErrorHandler(err)
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certPool,
			},
		},
	}

	if waitServer("https://localhost:9876", 5*time.Second, client) {
		//jika selama 5 detik server tidak ada response maka akan dimatikan dari sisi client
		fmt.Println("server missing")
		return
	}
	var choice int
	for {

		fmt.Println("Main Menu")
		fmt.Println("1. Get message")
		fmt.Println("2. Send file")
		fmt.Println("3. Check TLS")
		fmt.Println("4. Quit")
		fmt.Print(">> ")
		fmt.Scanf("%d\n", &choice)
		if choice == 1 {
			getMessage(client)
		} else if choice == 2 {
			sendFile(client)
		} else if choice == 3 {
			checkTLS(client)
		} else if choice == 4 {
			break
		} else {
			fmt.Println("Invalid choice")
		}
	}
}

func getMessage(client *http.Client) {
	resp, err := client.Get("https://localhost:9876")
	tools.ErrorHandler(err)
	// langsung menutup setiap kali menggunakan response guna mencegah data leak
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	tools.ErrorHandler(err)

	fmt.Println("Server:", string(data))
}

func sendFile(client *http.Client) {
	var name string
	var age int

	scanner := bufio.NewReader(os.Stdin)
	fmt.Print("Input name: ")
	name, _ = scanner.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Print("Input age: ")
	fmt.Scanf("%d\n", &age)

	person := data.Person{
		Name: name,
		Age:  age,
	}

	jsonData, err := json.Marshal(person)
	tools.ErrorHandler(err)

	temp := new(bytes.Buffer)
	w := multipart.NewWriter(temp)

	personField, err := w.CreateFormField("Person")
	tools.ErrorHandler(err)

	_, err = personField.Write(jsonData)
	tools.ErrorHandler(err)

	file, err := os.Open("./file.txt")
	tools.ErrorHandler(err)
	defer file.Close()

	fileField, err := w.CreateFormFile("file", file.Name())
	tools.ErrorHandler(err)

	_, err = io.Copy(fileField, file)
	tools.ErrorHandler(err)

	err = w.Close()
	tools.ErrorHandler(err)

	req, err := http.NewRequest("POST", "https://localhost:9876/sendFile", temp)
	tools.ErrorHandler(err)

	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := client.Do(req)
	tools.ErrorHandler(err)
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	tools.ErrorHandler(err)

	fmt.Println("Server:", string(data))
}

func checkTLS(client *http.Client) {
    // Make a request
    resp, err := client.Get("https://localhost:9876")
    if err != nil {
        fmt.Println("Failed to connect:", err)
        return
    }
    defer resp.Body.Close()

    // Check if the connection used TLS
    if resp.TLS == nil {
        fmt.Println("No TLS connection state found")
        return
    }

    // Print the TLS version
    switch resp.TLS.Version {
    case tls.VersionSSL30:
        fmt.Println("SSL 3.0")
    case tls.VersionTLS10:
        fmt.Println("TLS 1.0")
    case tls.VersionTLS11:
        fmt.Println("TLS 1.1")
    case tls.VersionTLS12:
        fmt.Println("TLS 1.2")
    case tls.VersionTLS13:
        fmt.Println("TLS 1.3")
    default:
        fmt.Println("Unknown TLS version")
    }

    // Print the Cipher Suite Name
    fmt.Println(tls.CipherSuiteName(resp.TLS.CipherSuite))

    // Print the Issuer Organization
    for _, cert := range resp.TLS.PeerCertificates {
        fmt.Println(cert.Issuer.Organization)
    }
}
