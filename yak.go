package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
	pam "yak/auth/PAM"

	"github.com/gorilla/securecookie"
	"libvirt.org/libvirt-go"
)

var domains []libvirt.Domain

const indexLoginPage = `
<!DOCTYPE html>
<head>
	<title> YAK </title>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bulma@0.9.2/css/bulma.min.css">
</head>
<body>
	<section class="section">
		<div class="container is-max-desktop">
			<div class="columns is-centered">
				<h1 class="title">YAK</h1>
				<figure class="image">
					<img src="images/yak.png">
				</figure>
			</div>
			<form class="box" method="post" action="/login">
				<div class="field">
					<label for="name">Username</label>
					<div class="control">
      					<input class="input" type="text" id="name" name="name" placeholder="e.g. jack">
    				</div>
				</div>

				<div class="field">
					<label for="password">Password</label>
					<div class="control">
      					<input class="input" type="password" id="password" name="password" placeholder="**********">
    				</div>
				</div>
				
				<input class="button is-primary" button type="submit" value="Login">
			</form>
		</div>
	</section>
</body>
`

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32))

func checkForAuthCookie(request *http.Request) (userName string) {
	cookie, err := request.Cookie("yak_session")
	if err == nil {
		cookieValue := make(map[string]string)
		err = cookieHandler.Decode("yak_session", cookie.Value, &cookieValue)
		if err == nil {
			userName = cookieValue["name"]
		}
	}

	return userName
}

func createAuthCookie(userName string, writer http.ResponseWriter, request *http.Request) {
	value := map[string]string{
		"name": userName,
	}
	encoded, err := cookieHandler.Encode("yak_session", value)
	if err == nil {
		cookie := &http.Cookie{
			Name:  "yak_session",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(writer, cookie)
	} else {
		log.Fatal(err)
	}
}

func indexHandleFunc(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/" {
		http.Error(writer, "404 not found.", http.StatusNotFound)
	}

	if request.Method != "GET" {
		http.Error(writer, "Method is not supported", http.StatusNotFound)
	}

	if checkForAuthCookie(request) == "" {
		fmt.Fprintf(writer, indexLoginPage)
	} else {
		fmt.Printf("%v : %v", request.RequestURI, request.RemoteAddr)

		fmt.Fprintf(writer, "Running VMs: \n")
		for _, domain := range domains {
			name, err := domain.GetName()
			if err == nil {
				fmt.Fprintf(writer, "%v\n", name)
			}
		}
	}
}

func loginHandleFunc(writer http.ResponseWriter, request *http.Request) {
	if request.Method == "POST" {
		name := request.FormValue("name")
		password := request.FormValue("password")
		if name != "" && password != "" {
			err := pam.PAMAuth(name, password)
			if err == nil {
				createAuthCookie(name, writer, request)
			} else {
				log.Printf("%v: %v\n", name, err.Error())
			}

			http.Redirect(writer, request, "/", 302)
		}
	} else {
		http.Error(writer, "Invalid request method.", 405)
	}
}

func logoutHandleFunc(writer http.ResponseWriter, request *http.Request) {
	if request.Method == "POST" {

	} else {
		http.Error(writer, "Invalid request method.", 405)
	}
}

func main() {
	connection, err := libvirt.NewConnect("qemu:///system")
	defer connection.Close() // NOTE(zak): Should we check err before we defer this. If defer doesnt check for null this could cause a crash
	if err == nil {
		fmt.Printf("Listening for connection on port 8082...\n")

		http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("./data/images"))))
		http.HandleFunc("/", indexHandleFunc)
		http.HandleFunc("/login", loginHandleFunc)
		http.HandleFunc("/logout", logoutHandleFunc)

		go func() {
			err = http.ListenAndServe(":8082", nil)
		}()

		for {
			domains, err = connection.ListAllDomains(0)
			time.Sleep(1 * time.Second)
		}
	} else {
		log.Fatal(err)
	}

}
