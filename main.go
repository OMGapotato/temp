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
    "errors"
    "strconv"

    "github.com/ContinuumLLC/BBS/r1util"
)

var static_data map[string]([]byte)

type RestClient struct {
    Hostname        string
    Port            string
}

type Key struct {
    Password string `json:"password"`
    KeyData string `json:"keyData"`
}

type Job struct {
    Name string `json:"name"`
    Builds []int64 `json:"builds"`
}

type JobHistory struct {
    Name string `json:"Name"`
    BuildHistory []Build `json:"BuildHistory"`
}

type JenkinsBuild struct {
    BuildNum    string
    Successful bool `json:"successful"`
    Url string `json:"url"`
    Artifacts []string `json:"artifacts"`
}

type Build struct {
    BuildNum string
    Successful bool
    Url string
    Artifacts []Artifact
}

type Artifact struct {
    Name string
    Url string
}

func startRouter(port int) {
    router := r1util.NewRouter("/", fmt.Sprintf("0.0.0.0:%d", port))

    router.AddRoute("GET", "/", staticHandler)
    router.AddRoute("GET", "/static/.*", staticHandler)

    router.AddRoute("GET", "/rest/auth", authHandler)
    router.AddRoute("GET", "/rest/systems", systemsHandler)
    router.AddRoute("GET", "/rest/artifacts/initialize", initializeArtifactsHandler)
    router.AddRoute("GET", "/rest/artifacts", artifactsHandler)
    router.AddRoute("GET", "/rest/system/artifacts/initialize", initializeSystemArtifactsHandler)
    router.AddRoute("GET", "/rest/:system/artifacts", systemArtifactsHandler)

    router.Run()
}

func CreateRestClient(hostname string, port string) (RestClient, error) {
    var retVal RestClient

    retVal.Hostname = hostname
    retVal.Port = port

    return retVal, nil
}

func (r RestClient) Execute(restUrl string, method string) ([]byte, error) {
    client := new(http.Client)
    req, err := http.NewRequest(method, "http://" + r.Hostname + ":" + r.Port + restUrl, nil)
    if err != nil {
            r1util.LogError("Failed to create request: " + err.Error())
            return nil, errors.New("Failed to create request: " + err.Error())
    }
    resp, err := client.Do(req)
    if err != nil {
            r1util.LogError("Failed to execute rest call: " + err.Error())
            return nil, err
    }
    defer resp.Body.Close()

    bytes, err := ioutil.ReadAll(resp.Body)
    if err != nil {
            r1util.LogError("Failed to execute rest call: " + err.Error())
            return nil, errors.New("Failed to execute rest call: " + err.Error())
    }
    return bytes, nil
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

func systemsHandler(resp http.ResponseWriter, req *http.Request, parsedURL map[string]string) *r1util.AppError {
    client, err := CreateRestClient("10.80.199.100", "9292")
    if err != nil {
        r1util.LogError("Failed to create rest client")
        return &r1util.AppError{err, "Failed to create rest client: " + err.Error(), 500}
    }

    result, err := client.Execute("/systems", "GET")
    if err != nil {
        r1util.LogError("Failed to execute rest call")
        return &r1util.AppError{err, "Failed to execute rest call: " + err.Error(), 500}
    }

    resp.Header().Set("Content-Type", "application/json")
    resp.Write(result)
    return nil
}

func initializeArtifactsHandler(resp http.ResponseWriter, req *http.Request, parsedURL map[string]string) *r1util.AppError {
    
    client, err := CreateRestClient("10.80.199.100", "9292")
    if err != nil {
        r1util.LogError("Failed to create rest client")
        return &r1util.AppError{err, "Failed to create rest client: " + err.Error(), 500}
    }

    result, err := client.Execute("/jobs", "GET")
    if err != nil {
        r1util.LogError("Failed to execute rest call")
        return &r1util.AppError{err, "Failed to execute rest call: " + err.Error(), 500}
    }

    var jobList []string = make([]string, 0, 0)
    err = json.Unmarshal(result, &jobList)
    
    result, err = client.Execute("/systems", "GET")
    if err != nil {
        r1util.LogError("Failed to execute rest call")
        return &r1util.AppError{err, "Failed to execute rest call: " + err.Error(), 500}
    }

    var systemList []string = make([]string, 0, 0)
    err = json.Unmarshal(result, &systemList)

    var jobHistoryList []JobHistory

    for _, job := range jobList {

        jobHistory := new(JobHistory)

        result, err := client.Execute("/jobs" + "/" + job, "GET")
        if err != nil {
            r1util.LogError("Failed to execute rest call")
            return &r1util.AppError{err, "Failed to execute rest call: " + err.Error(), 500}
        }

        var jobObj Job 
        json.Unmarshal(result, &jobObj)

        jobHistory.Name = jobObj.Name

        for _, build := range jobObj.Builds {
            buildNum := strconv.FormatInt(build, 10)
            result, err := client.Execute("/jobs" + "/" + job + "/" + buildNum, "GET")
            if err != nil {
                r1util.LogError("Failed to execute rest call")
                return &r1util.AppError{err, "Failed to execute rest call: " + err.Error(), 500}
            }
           
            var buildObj Build
            var artifactList []Artifact
            var jenkinsBuildObj JenkinsBuild

            jenkinsBuildObj.BuildNum = buildNum
            json.Unmarshal(result, &jenkinsBuildObj)

            for _, artifact := range jenkinsBuildObj.Artifacts {
                var artifactObj Artifact
                splitUrl := strings.Split(artifact, "/")
                tmpArtifact := splitUrl[len(splitUrl)-1]
                newTmpArtifact := strings.Replace(tmpArtifact, ".deb", "", -1)
                splitArtifactName := strings.Split(newTmpArtifact, "-")
                
                // artifact name without '.deb', '-', '/', and job/version
                splitArtifactName = splitArtifactName[:len(splitArtifactName)-1] 
                var artifactName string
                for i:=0; i<len(splitArtifactName); i++ {
                    if i == 0 {
                        artifactName = splitArtifactName[i]
                    } else {
                        artifactName = artifactName + "-" + splitArtifactName[i]
                    }
                }

                artifactObj.Name = artifactName
                artifactObj.Url = artifact
                artifactList = append(artifactList, artifactObj)
            }

            buildObj.BuildNum = jenkinsBuildObj.BuildNum
            buildObj.Successful = jenkinsBuildObj.Successful
            buildObj.Url = jenkinsBuildObj.Url
            buildObj.Artifacts = artifactList

            jobHistory.BuildHistory = append(jobHistory.BuildHistory, buildObj)
        }
      
        jobHistoryList = append(jobHistoryList, *jobHistory)
 
        bytes, err := json.Marshal(jobHistory)
        if err != nil {
            r1util.LogError("Failed to create JSON for job history")
            return &r1util.AppError{err, "Failed to create JSON for job history: " + err.Error(), 500}
        }
 
        fileName := "/home/jkwon/Git/releaseBuilder/BuildHistory/" + jobHistory.Name + ".BuildHistory.json"
        err = ioutil.WriteFile(fileName, bytes, 0644)
        if err != nil {
            r1util.LogError("Failed to create JSON for job history")
            return &r1util.AppError{err, "Failed to create JSON for job history: " + err.Error(), 500}
        }
    }

    fmt.Printf("%v", jobHistoryList)

    bytes, err := json.Marshal(jobHistoryList)
    if err != nil {
        r1util.LogError("Failed to create JSON for complete job history")
        return &r1util.AppError{err, "Failed to create JSON for complete job history: " + err.Error(), 500}
    }

    fileName := "/home/jkwon/Git/releaseBuilder/BuildHistory/CompleteList.json"
    err = ioutil.WriteFile(fileName, bytes, 0644)
    if err != nil {
        r1util.LogError("Failed to create JSON for complete job history")
        return &r1util.AppError{err, "Failed to create JSON for complete job history: " + err.Error(), 500}
    }

    resp.Header().Set("Content-Type", "application/json")
    resp.Write(result)
    return nil
}

func artifactsHandler(resp http.ResponseWriter, req *http.Request, parsedURL map[string]string) *r1util.AppError {
    
    var jobHistoryList []JobHistory 

    bytes, err := ioutil.ReadFile("/home/jkwon/Git/releaseBuilder/BuildHistory/CompleteList.json")
    if err != nil {
        r1util.LogError("Failed to read complete job history")
        return &r1util.AppError{err, "Failed to read complete job history: " + err.Error(), 500}
    }

    json.Unmarshal(bytes, &jobHistoryList)
    
    resp.Header().Set("Content-Type", "application/json")
    resp.Header().Set("Access-Control-Allow-Origin", "*")
    resp.Write(bytes)
    
    return nil
}

func initializeSystemArtifactsHandler(resp http.ResponseWriter, req *http.Request, parsedURL map[string]string) *r1util.AppError {

    client, err := CreateRestClient("10.80.199.100", "9292")
    if err != nil {
        r1util.LogError("Failed to create rest client")
        return &r1util.AppError{err, "Failed to create rest client: " + err.Error(), 500}
    }

    result, err := client.Execute("/systems", "GET")
    if err != nil {
        r1util.LogError("Failed to execute rest call")
        return &r1util.AppError{err, "Failed to execute rest call: " + err.Error(), 500}
    }
    
    var systemList []string = make([]string, 0, 0)
    err = json.Unmarshal(result, &systemList)
    
    var jobHistoryList []JobHistory 

    bytes, err := ioutil.ReadFile("/home/jkwon/Git/releaseBuilder/BuildHistory/CompleteList.json")
    if err != nil {
        r1util.LogError("Failed to read complete job history")
        return &r1util.AppError{err, "Failed to read complete job history: " + err.Error(), 500}
    }

    json.Unmarshal(bytes, &jobHistoryList)

    

    return nil
}

func systemArtifactsHandler(resp http.ResponseWriter, req *http.Request, parsedURL map[string]string) *r1util.AppError {
    return nil
}

func main() {
    port := flag.Int("port", 4030, "run at port")
    static_data = GetFileMap()
    startRouter(*port)
}
