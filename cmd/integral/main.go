package main

import (
    "io"
    "golang.org/x/oauth2"
    "context"
    "path/filepath"
	"database/sql"
	"fmt"
	"html/template"
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

var clientid = os.Getenv("CLIENT_ID")
var clientSecret = os.Getenv("CLIENT_SECRET")
var realm = os.Getenv("REALM")

var oauthConfig = &oauth2.Config{
    RedirectURL:  "http://localhost:8080/callback",
    ClientID:     clientid,
    ClientSecret: clientSecret,
    Scopes:       []string{"openid", "profile", "email"},
    Endpoint: oauth2.Endpoint{
        AuthURL:  "http://localhost:8069/realms/myrealm/protocol/openid-connect/auth",
        TokenURL: "http://localhost:8069/realms/myrealm/protocol/openid-connect/token",
    },
}

var oauthStateString = "random"

func main() {
    // Serve static files
    fs := http.FileServer(http.Dir("web/static"))
    http.Handle("/static/", http.StripPrefix("/static/", fs))

    // Set up routes
    http.HandleFunc("/db", homeHandler)
    http.HandleFunc("/", testHandler)
    http.HandleFunc("/upload", uploadFileHandler)
    http.HandleFunc("/delete", deletePhotosHandler)
    http.HandleFunc("/servedelete", serveDelete)
    http.HandleFunc("/update", serveUpdate)
    http.HandleFunc("/create", serveCreate)
    http.HandleFunc("/login", loginHandler)
    http.HandleFunc("/callback", callbackHandler)

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

func loginHandler(w http.ResponseWriter, r *http.Request) {
    url := oauthConfig.AuthCodeURL(oauthStateString, oauth2.AccessTypeOffline)
    http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
    // Check that the state parameter matches
    state := r.FormValue("state")
    if state != oauthStateString {
        fmt.Printf("invalid oauth state, expected %s, got %s\n", oauthStateString, state)
        http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
        return
    }

    code := r.FormValue("code")
    token, err := oauthConfig.Exchange(context.Background(), code)
    if err != nil {
        fmt.Printf("oauthConfig.Exchange() failed with %s\n", err)
        http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
        return
    }
    // Create a new HTTP client
        client := oauthConfig.Client(context.Background(), token)

        // Define the user info endpoint of the OAuth provider
        // This will vary depending on the provider and should be replaced with the actual endpoint
        userInfoURL := "http://localhost:8069/realms/myrealm/protocol/openid-connect/userinfo"

        // Make a request to the user info endpoint
        resp, err := client.Get(userInfoURL)
        if err != nil {
            // Handle error
            fmt.Printf("Error retrieving user info: %s\n", err)
            http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
            return
        }
        defer resp.Body.Close()

        // Read and process the response
        // The exact processing will depend on the structure of the response
        userInfo, err := io.ReadAll(resp.Body)
        if err != nil {
            // Handle error
            fmt.Printf("Error reading user info response: %s\n", err)
            http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
            return
        }

        // Do something with the user info, e.g., display it or create a user session
        fmt.Fprintf(w, "User Info: %s", string(userInfo))
}

func serveCreate(w http.ResponseWriter, r *http.Request) {
    // Parse the template
    images2 := listPhotos()
    fmt.Println(images2)
    tmpl, err := template.ParseFiles("./web/template/base.html", "./web/template/create.html")
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
func serveDelete(w http.ResponseWriter, r *http.Request) {
    // Parse the template
    images2 := listPhotos()
    fmt.Println(images2)
    tmpl, err := template.ParseFiles("./web/template/base.html", "./web/template/delete.html")
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

func serveUpdate(w http.ResponseWriter, r *http.Request) {
    // Parse the template
    images2 := listPhotos()
    fmt.Println(images2)
    tmpl, err := template.ParseFiles("./web/template/base.html", "./web/template/edit.html")
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
