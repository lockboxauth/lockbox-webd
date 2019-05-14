package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"darlinggo.co/trout"
)

func main() {
	var router trout.Router
	router.Endpoint("/login").Methods("GET").Handler(http.HandlerFunc(getLoginHandler))
	router.Endpoint("/login").Methods("POST").Handler(http.HandlerFunc(postLoginHandler))
	router.Endpoint("/register").Methods("GET").Handler(http.HandlerFunc(getRegisterHandler))
	router.Endpoint("/register").Methods("POST").Handler(http.HandlerFunc(postRegisterHandler))
	http.Handle("/", router)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	err := http.ListenAndServe(":9876", nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func getLoginHandler(w http.ResponseWriter, r *http.Request) {
	templates := template.Must(findAndParseTemplates("templates"))
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := templates.ExecuteTemplate(w, "login/login.html", nil)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

func postLoginHandler(w http.ResponseWriter, r *http.Request) {
	templates := template.Must(findAndParseTemplates("templates"))
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := templates.ExecuteTemplate(w, "login-confirm/login-confirm.html", nil)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

func getRegisterHandler(w http.ResponseWriter, r *http.Request) {
	templates := template.Must(findAndParseTemplates("templates"))
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := templates.ExecuteTemplate(w, "register/register.html", nil)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

func postRegisterHandler(w http.ResponseWriter, r *http.Request) {
	reqMap := map[string]interface{}{
		"id":             r.PostFormValue("email"),
		"isRegistration": true,
	}
	b, err := json.Marshal(reqMap)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	resp, err := http.Post("http://localhost:12345/accounts/v1", "application/json", bytes.NewReader(b))
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	if resp.StatusCode != http.StatusCreated {
		fmt.Println(resp.StatusCode)
		fmt.Println(string(body))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("%d\n%s", resp.StatusCode, body)))
		return
	}
	fmt.Println(string(body))
	templates := template.Must(findAndParseTemplates("templates"))
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = templates.ExecuteTemplate(w, "register-confirm/register-confirm.html", nil)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

func findAndParseTemplates(rootDir string) (*template.Template, error) {
	cleanRoot := filepath.Clean(rootDir)
	pfx := len(cleanRoot) + 1
	root := template.New("")

	err := filepath.Walk(cleanRoot, func(path string, info os.FileInfo, e1 error) error {
		if !info.IsDir() {
			if e1 != nil {
				return e1
			}

			b, e2 := ioutil.ReadFile(path)
			if e2 != nil {
				return e2
			}

			name := path[pfx:]
			t := root.New(name)
			t, e2 = t.Parse(string(b))
			if e2 != nil {
				return e2
			}
		}

		return nil
	})

	return root, err
}
