package main

import (
	"fmt"
	"log"
	"net/http"

	"libvirt.org/libvirt-go"
)

var domains []libvirt.Domain

func HTTPHandleFunc(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/" {
		http.Error(writer, "404 not found.", http.StatusNotFound)
	}

	if request.Method != "GET" {

		http.Error(writer, "Method is not supported", http.StatusNotFound)
	}

	fmt.Fprintf(writer, "Running VMs: \n")
	for _, domain := range domains {
		name, err := domain.GetName()
		if err == nil {
			fmt.Fprintf(writer, "%v\n", name)
		}
	}
}

func main() {
	connection, err := libvirt.NewConnect("qemu:///system")
	defer connection.Close() // NOTE(zak): Should we check err before we defer this. If defer doesnt check for null this could cause a crash
	if err == nil {
		domains, err := connection.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE)
		if err == nil {
			fmt.Printf("%d running domains:\n", len(domains))
			for _, domain := range domains {
				name, err := domain.GetName()
				if err == nil {
					fmt.Printf("  %s\n", name)
				}
			}
		} else {
			log.Fatal(err)
		}
	} else {
		log.Fatal(err)
	}

	fmt.Printf("Listening for connection on port 8082...\n")

	http.HandleFunc("/", HTTPHandleFunc)
	if err := http.ListenAndServe(":8082", nil); err != nil {
		log.Fatal(err)
	}
}
