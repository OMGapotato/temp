package main

import (
    "fmt"
    "io/ioutil"
    "encoding/binary"
    "encoding/json"
    "flag"
    "time"
    "encoding/base64"
    "mime"
    "net/http"
    "path/filepath"
    "strings"

    "github.com/ContinuumLLC/BBS/r1util"
)

var static_data map[string]([]byte)

type Key struct {
    Password string `json:"password"`
    KeyData string `json:"keyData"`
}

func staticHandler(resp http.ResponseWriter, req *http.Request, parsedURL map[string]string) *r1util.AppError {
    var (
        data       []byte
        path       string
        media_type string
        ok         bool
    )
    // Igonre URL parameters, for now. If needed we'll use a web framework.
    path = strings.Split(req.URL.Path, "?")[0]
    //

    if path == "/" {
        if _, ok := req.Header["Cookie"]; ok {
            http.Redirect(resp, req, "/static/index.html", http.StatusMovedPermanently)
        } else {
            http.Redirect(resp, req, "/static/login.html", http.StatusMovedPermanently)
        }
        return nil
    }

    if path == "/static/login.html" {
        if _, ok := req.Header["Cookie"]; ok {
            http.Redirect(resp, req, "/static/index.html", http.StatusMovedPermanently)
            return nil
        }
    }

    if path == "/static/index.html" {
        if _, ok := req.Header["Cookie"]; !ok {
            http.Redirect(resp, req, "/static/login.html", http.StatusMovedPermanently)
            return nil
        }
    }

    // We are ignoring the first char as it always starts with "/" for requests. But we don't
    // have "/" in data packed with "gopack.pl" tool.
    if data, ok = static_data[path[1:]]; !ok {
            r1util.LogError("Could not find static file from cache: " + path)
            http.NotFound(resp, req)
            return nil
    }

    media_type = mime.TypeByExtension(filepath.Ext(path))
    if media_type == "" {
            media_type = http.DetectContentType(data)
    }

    resp.Header().Set("Content-Type", media_type)
    binary.Write(resp, binary.BigEndian, data)
    return nil
}

func startRouter(port int) {
    router := r1util.NewRouter("/", fmt.Sprintf("0.0.0.0:%d", port))

    router.AddRoute("GET", "/", staticHandler)
    router.AddRoute("GET", "/static/.*", staticHandler)

    router.AddRoute("GET", "/rest/seed/checkin", seedCheckinHandler)
    router.AddRoute("GET", "/rest/seed/:id/start", startSeedHandler)
    router.AddRoute("GET", "/rest/seed/:id/progress", seedProgressHandler)
    router.AddRoute("GET", "/rest/seed/:id/cancel", cancelSeedHandler)
    router.AddRoute("GET", "/rest/seed/:id/retry", retrySeedHandler)
    router.AddRoute("GET", "/rest/seed/:id/logs", logsSeedHandler)
    router.AddRoute("GET", "/rest/auth", authHandler)
    router.AddRoute("GET", "/rest/auth/test", testHandler)

    router.Run()
}

func authHandler(resp http.ResponseWriter, req *http.Request, parsedURL map[string]string) *r1util.AppError {

    if _, ok := req.Header["Authorization"]; !ok {
        return &r1util.AppError{nil, "Invalid request header", 500}
    }

    auth := strings.SplitN(req.Header["Authorization"][0], " ", 2)

    if len(auth) != 2 || auth[0] != "Basic" {
        return &r1util.AppError{nil, "Invalid request header", 500}
    }

    payload, _ := base64.StdEncoding.DecodeString(auth[1])
    password := string(payload)

    b, err := ioutil.ReadFile("/home/jkwon/Git/BBS/sdLite/keys.json")
    if err != nil {
            return &r1util.AppError{err, "Error getting user: " + err.Error(), 500}
    }

    var keys []Key

    json.Unmarshal(b, &keys)

    var cookie *http.Cookie
    for i:= 0; i<len(keys); i++ {
        if keys[i].Password == password {
            expiration := time.Now().Add(365 * 24 * time.Hour)
            cookie = &http.Cookie{Name: "temp", Value:password , Expires: expiration}
            http.SetCookie(resp, cookie)
        }
    }

    if cookie == nil {
        return &r1util.AppError{nil, "Invalid password", 400}
    }

    return nil
}

func testHandler(resp http.ResponseWriter, req *http.Request, parsedURL map[string]string) *r1util.AppError {

    fmt.Printf("Testhandler: %v\n", req.Header["Cookie"])

    resp.Header().Set("Content-Type", "application/json")
    bytes, _ := json.Marshal("hey")
    resp.Write(bytes)

    return nil

}

func seedCheckinHandler(resp http.ResponseWriter, req *http.Request, parsedURL map[string]string) *r1util.AppError {

    bytes, err := ioutil.ReadFile("/home/jkwon/Git/BBS/sdLite/waiting.json")
    if err != nil {
            return &r1util.AppError{err, "Error creating spool list: " + err.Error(), 500}
    }

    resp.Header().Set("Content-Type", "application/json")
    resp.Write(bytes)

    return nil
}

func startSeedHandler(resp http.ResponseWriter, req *http.Request, parsedURL map[string]string) *r1util.AppError {

    resp.Write([]byte("seeding"))

    return nil
}

func seedProgressHandler(resp http.ResponseWriter, req *http.Request, parsedURL map[string]string) *r1util.AppError {

    bytes, err := ioutil.ReadFile("/home/jkwon/Git/BBS/sdLite/progress.json")
    if err != nil {
            return &r1util.AppError{err, "Error creating spool list: " + err.Error(), 500}
    }

    resp.Header().Set("Content-Type", "application/json")
    resp.Write(bytes)

    return nil

}

func cancelSeedHandler(resp http.ResponseWriter, req *http.Request, parsedURL map[string]string) *r1util.AppError {

    resp.Write([]byte("cancelling seed"))

    return nil
}

func retrySeedHandler(resp http.ResponseWriter, req *http.Request, parsedURL map[string]string) *r1util.AppError {

    resp.Write([]byte("retrying seed"))

    return nil
}

func logsSeedHandler(resp http.ResponseWriter, req *http.Request, parsedURL map[string]string) *r1util.AppError {

    bytes, err := ioutil.ReadFile("/home/jkwon/Git/BBS/sdLite/logs.json")
    if err != nil {
            return &r1util.AppError{err, "Error creating spool list: " + err.Error(), 500}
    }

    resp.Header().Set("Content-Type", "application/json")
    resp.Write(bytes)

    return nil

}

func main() {
    port := flag.Int("port", 4050, "run at port")
    static_data = GetFileMap()
    startRouter(*port)
}
