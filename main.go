// lens project main.go
package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"runtime"
)

func main() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	path := ""
	switch runtime.GOOS {
	case "windows":
		path = "\\"
	case "linux":
		path = "/"
	}
	filename := fmt.Sprintf("%s%s.ip.address", usr.HomeDir, path)

	args := url.Values{}
	args.Set("tkn", "example")
	args.Set("email", "foo@example.com")
	args.Set("a", "rec_edit")
	args.Set("z", "example.com")
	args.Set("type", "A")
	args.Set("id", "1234567890")
	args.Set("name", "foo.example.com")
	args.Set("ttl", "1")
	args.Set("service_mode", "1")

	cloudflare_url := "https://www.cloudflare.com/api_json.html"

	//Get local external IP address.
	resp, err := http.Get("http://ipecho.net/plain")
	if err != nil {
		fmt.Printf("Request failed.. %s\n", err)
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	externalip := string(body)
	//TODO: Check the http status codes.

	//Open the ip address file.
	fh, err := os.Open(filename)
	if err != nil && os.IsNotExist(err) {
		//If the file did not open, create it.
		fh, err = os.Create(filename)
		if err != nil {
			//If that did not work then panic
			fmt.Printf("Error opening file handle: %s\n", err)
			panic(err)
		}
	} else if err != nil {
		fmt.Printf("Unhandled error: %s\n", err)
	}
	defer fh.Close()

	//Get last known external IP address.
	r := bufio.NewReader(fh)
	oldip, err := r.ReadString('\n')
	if err != nil && err != io.EOF {
		fmt.Printf("Opening IP file failed. %s\n", err)
	}

	fmt.Printf("Old ip: %s\nNew Ip: %s\n", oldip, externalip)
	//Check to see if the IP address has changed.
	if externalip != oldip {
		fmt.Printf("Ip has changed updating...\n")
		//Update cloudflare.
		args.Set("content", externalip)
		resp, err = http.PostForm(cloudflare_url, args)
		defer resp.Body.Close()
		if err != nil {
			fmt.Printf("Error posting to cloudflare %s", err)
			panic(err)
		}

	} else {
		fmt.Printf("Nothing has changed. Exiting...\n")
		os.Exit(0)
	}

	//Save the new IP to disk
	fh, err = os.Create(filename)
	if err != nil {
		fmt.Printf("Unhandled error: %s\n", err)
	}
	defer fh.Close()
	w := bufio.NewWriter(fh)
	w.WriteString(externalip)
	w.Flush()
}
