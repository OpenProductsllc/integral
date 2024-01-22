package main

import (
    "path/filepath"
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"

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
    http.HandleFunc("/db", homeHandler)
    http.HandleFunc("/", testHandler)
    http.HandleFunc("/upload", uploadFileHandler)
    http.HandleFunc("/delete", deletePhotosHandler)

    // Start the server
    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
    db := connectDB()
    defer db.Close()
    imageRepo := &models.ImageRepository{DB: db}
    images, err := imageRepo.GetAllImages()
    fmt.Println(images)
    // Parse the template
    images2 := listPhotos()
    tmpl, err := template.ParseFiles("./web/template/base.html", "./web/template/list.html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Execute the template
    err = tmpl.Execute(w, images2)
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

func uploadFileHandler(w http.ResponseWriter, r *http.Request) {
    // Limit the size of the file to prevent memory overload
    r.ParseMultipartForm(10 << 20) // 10 MB

    // Retrieve the file from the form data
    file, handler, err := r.FormFile("file")
    if err != nil {
        fmt.Println("Error Retrieving the File")
        fmt.Println(err)
        return
    }
    defer file.Close()

    // Save the file to a destination
    dst, err := os.Create("./web/static/uploads/" + handler.Filename)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer dst.Close()

    // Copy the uploaded file to the destination file
    _, err = io.Copy(dst, file)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    //fmt.Fprintf(w, "Successfully Uploaded File\n")
    http.Redirect(w, r, "/", 301)
}

func deletePhotosHandler(w http.ResponseWriter, r *http.Request) {
    deleteFiles("./web/static/uploads")

    http.Redirect(w, r, "/", 301)
}

func listPhotos() []string {

    photos := []string {}
    // Open the directory
    dir, err := os.Open("./web/static/uploads")
    if err != nil {
        log.Fatal(err)
    }
    defer dir.Close()

    // Read directory entries
    files, err := dir.Readdir(-1)
    if err != nil {
        log.Fatal(err)
    }

    // Print file names
    for _, file := range files {
        fmt.Println(file.Name())
        photos = append(photos, file.Name())
    }
    return photos
}

func testHandler(w http.ResponseWriter, r *http.Request) {
    // Parse the template
    images2 := listPhotos()
    fmt.Println(images2)
    tmpl, err := template.ParseFiles("./web/template/base.html", "./web/template/test.html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Execute the template
    err = tmpl.Execute(w, images2)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func deleteFiles(directory string) error {
    files, err := os.ReadDir(directory)
    if err != nil {
        return err
    }

    for _, file := range files {
        // Skip directories
        if file.IsDir() {
            continue
        }

        // Construct full file path
        filePath := filepath.Join(directory, file.Name())

        // Delete the file
        err := os.Remove(filePath)
        if err != nil {
            return err
        }

        fmt.Printf("Deleted file: %s\n", filePath)
    }

    return nil
}
