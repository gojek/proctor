package docs

import "net/http"

func APIDocHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/docs" {
		http.Redirect(w, r, "/docs/", http.StatusFound)
		return
	}
}
