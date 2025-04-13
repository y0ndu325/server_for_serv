package main

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"
    "os"

    _ "github.com/lib/pq"
)

var (
    host     = getEnv("PGHOST", "localhost")
    user     = getEnv("PGUSER", "postgres")
    port     = getEnv("PGPORT", "5432")
    password = getEnv("PGPASSWORD", "password")
    dbname   = getEnv("PGDATABASE", "postgres")
)

func getEnv(key, defaultValue string) string {
    if value, exists := os.LookupEnv(key); exists {
        return value
    }
    return defaultValue
}

var db *sql.DB

func initDB() error {
    connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        host, port, user, password, dbname)
    var err error

    db, err = sql.Open("postgres", connStr)
    if err != nil {
        return fmt.Errorf("error opening database: %v", err)
    }

    if err := db.Ping(); err != nil {
        return fmt.Errorf("error pinging database: %v", err)
    }

    return nil
}

func ardHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodGet {
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        fmt.Fprintf(w, `
            <h1>Управление Arduino</h1>
            <form method="POST" action="/">
                <button type="submit" name="state" value="1">Включить (1)</button>
                <button type="submit" name="state" value="0">Выключить (0)</button>
            </form>
        `)
    } else if r.Method == http.MethodPost {
        state := r.FormValue("state")
        if state != "1" && state != "0" {
            http.Error(w, "Invalid state value", http.StatusBadRequest)
            return
        }

        query := `INSERT INTO arduino_control (state) VALUES ($1)`

        _, err := db.Exec(query, state)
        if err != nil {
            http.Error(w, "Error inserting data into database", http.StatusInternalServerError)
            return
        }
        fmt.Fprintf(w, "State: %s", state)
    }
}

func getStateHandler(w http.ResponseWriter, r *http.Request) {
    var state int
    err := db.QueryRow("SELECT state FROM arduino_control ORDER BY id DESC LIMIT 1").Scan(&state)
    if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }
    fmt.Fprintf(w, "%d", state)
}

func main() {
    err := initDB()
    if err != nil {
        log.Fatalf("error initializing database: %v", err)
    }
    defer db.Close()

    http.HandleFunc("/", ardHandler)
    http.HandleFunc("/get_state", getStateHandler)

    port := os.Getenv("PORT")
    if port == "" {
        port = "4000"
    }

    fmt.Printf("Server started on: http://localhost:%s\n", port)
    err = http.ListenAndServe("0.0.0.0:"+port, nil)
    if err != nil {
        fmt.Println("Error starting server:", err)
    }
}