package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

func auth(w http.ResponseWriter, r *http.Request) {

	jwtString := r.URL.Query().Get("jwt")

	if jwtString == "" {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "No JWT")
		return
	}

	// parse the token
	token, err := jwt.Parse(jwtString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return []byte(os.Getenv("MYRADIO_SIGNING_KEY")), nil
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}

	session, _ := cookiestore.Get(r, AuthRealm)

	memberid := claims["uid"].(float64)
	name := claims["name"].(string)

	myRadioNameCache[int(memberid)] = myRadioNameCacheObject{
		name:      name,
		cacheTime: time.Now(),
	}

	session.Values["memberid"] = int(memberid)
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)

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
			http.Redirect(w, r, fmt.Sprintf("https://ury.org.uk/myradio/MyRadio/jwt?redirectto=%s/auth", os.Getenv("HOST")), http.StatusFound)
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
