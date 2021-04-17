package main

import (
	"fmt"
	"log"
	"net/http"
	pam "yak/auth/PAM"

	"github.com/gorilla/securecookie"
	"libvirt.org/libvirt-go"
)

var domains []libvirt.Domain

const indexLoginPage = `
<h1>Login</h1>
<form method="post" action="/login">
    <label for="name">User name</label>
    <input type="text" id="name" name="name">
	<br>
    <label for="password">Password</label>
    <input type="password" id="password" name="password">
	<br><br>
    <button type="submit">Login</button>
</form>
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
		domains, err = connection.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE)
	} else {
		log.Fatal(err)
	}

	fmt.Printf("Listening for connection on port 8082...\n")

	http.HandleFunc("/", indexHandleFunc)
	http.HandleFunc("/login", loginHandleFunc)
	http.HandleFunc("/logout", logoutHandleFunc)

	if err := http.ListenAndServe(":8082", nil); err != nil {
		log.Fatal(err)
	}
}
