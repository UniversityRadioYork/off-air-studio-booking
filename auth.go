package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
)

func auth(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		http.ServeFile(w, r, "login.html")
		return
	} else if r.Method == "POST" {
		session, _ := cookiestore.Get(r, AuthRealm)

		memberid := r.FormValue("memberid")
		if memberid == "" {
			http.Redirect(w, r, "/auth", http.StatusFound)
			return
		}

		var err error
		session.Values["memberid"], err = strconv.Atoi(memberid)
		if err != nil {
			panic(err)
		}
		session.Save(r, w)
		http.Redirect(w, r, "/", http.StatusFound)

	}
	fmt.Fprint(w, "hmmm")
}

func AuthHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/auth" {
			h.ServeHTTP(w, r)
			return
		}

		session, _ := cookiestore.Get(r, AuthRealm)
		if auth, ok := session.Values["memberid"].(int); !ok || auth == 0 {
			// redirect to auth
			http.Redirect(w, r, "/auth", http.StatusFound)
		} else {
			ctx := context.WithValue(context.Background(), UserCtxKey, auth)
			h.ServeHTTP(w, r.WithContext(ctx))
		}

	})
}

func logout(w http.ResponseWriter, r *http.Request) {
	session, _ := cookiestore.Get(r, AuthRealm)
	session.Values["memberid"] = 0
	session.Save(r, w)
	http.Redirect(w, r, "https://ury.org.uk/myradio/MyRadio/logout", http.StatusFound)
}
