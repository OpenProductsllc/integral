package main

import (
    "net/http"
    "log"
    "html/template"
)

func main() {
    // Serve static files
    fs := http.FileServer(http.Dir("web/static"))
    http.Handle("/static/", http.StripPrefix("/static/", fs))

    // Set up routes
    http.HandleFunc("/", homeHandler)

    // Start the server
    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
    content := []string {
        "IMAGE1",
        "IMAGE2"}
    // Parse the template
    tmpl, err := template.ParseFiles("./web/template/base.html", "./web/template/list.html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Execute the template
    err = tmpl.Execute(w, content)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}
