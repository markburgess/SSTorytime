package main

import (
	"net/http"
	"net/url"
	"io"
	"fmt"
	"crypto/tls"
	"crypto/x509"
	"os"
)


func main() {

	uri := "https://127.0.0.1:8443/searchN4L"
	query := "brain \\chapter neuro"
	
	formdata := url.Values{
		"name": { query },
	}

	var body []byte
	
	fmt.Println("Test the certificate",uri,formdata)
	
	resp, err := http.PostForm(uri, formdata)
	
	if err != nil {
		fmt.Printf("POST: Unable to forward request: %s\n", "N4Lquery")
		body = SelfSignedForm(uri,query,formdata)
		return
	}
	
	defer resp.Body.Close()
	
	body, _ = io.ReadAll(resp.Body)
	
	fmt.Println("TRUSTED",string(body))
}

// *********************************************************************

func SelfSignedForm(uri,query string,formdata url.Values) []byte {
	
	// curl -Iv https://127.0.0.1:8443 --cacert cert.pem

	caCert, err := os.ReadFile("../server/cert.pem")
	
	if err != nil {
		fmt.Println("Couldn't load server's self-signed certificate file",err)
		return nil
	}
	
	// 2. Create a CertPool and add the CA certificate
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	
	// 3. Configure TLS with the custom CertPool
	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}
	
	// 4. Create an HTTP client with the custom TLS configuration
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}
	
	fmt.Println("Try to connect FORM",uri,formdata)
	
	resp, err2 := client.PostForm(uri, formdata)
	
	if err2 != nil {
		fmt.Printf("POST: Unable to forward request: %s\n", "N4Lquery")
		return nil
	}
	
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	
	fmt.Println("SELF_SIGNED",string(body))
	return body
}
