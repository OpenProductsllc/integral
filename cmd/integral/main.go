package main

import (
    "fmt"
    "database/sql"
    "net/http"
    "log"
    "html/template"
    _ "github.com/lib/pq"
    "os"
    
    "github.com/OpenProductsllc/integral/internal/models"
)

var dbHost = os.Getenv("DB_HOST")
var dbPort = os.Getenv("DB_PORT")
var dbUser = os.Getenv("DB_USER")
var dbPassword = os.Getenv("DB_PASSWORD")
var dbName = os.Getenv("DB_NAME")

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
    db := connectDB()
    defer db.Close()
    imageRepo := &models.ImageRepository{DB: db}
    images, err := imageRepo.GetAllImages()
    // Parse the template
    tmpl, err := template.ParseFiles("./web/template/base.html", "./web/template/list.html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Execute the template
    err = tmpl.Execute(w, images)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func connectDB() *sql.DB{
    psqlconn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        dbHost, dbPort, dbUser, dbPassword, dbName)
    // open database
    db, _ := sql.Open("postgres", psqlconn)
    return db
}
