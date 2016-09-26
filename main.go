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
    "reflect"

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
    Name string `json:"Name"`
    Builds []int64 `json:"Builds"`
}

type JobHistory struct {
    JobName string `json:"JobName"`
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
    Job string
    Version string
    Build string
}

func startRouter(port int) {
    router := r1util.NewRouter("/", fmt.Sprintf("0.0.0.0:%d", port))

    router.AddRoute("GET", "/", staticHandler)
    router.AddRoute("GET", "/static/.*", staticHandler)

    router.AddRoute("GET", "/rest/auth", authHandler)
    router.AddRoute("GET", "/rest/systems", systemsHandler)
    router.AddRoute("GET", "/rest/systems/:system", systemsPackagesHandler)
    router.AddRoute("GET", "/rest/artifacts/initialize", initializeArtifactsHandler)
    router.AddRoute("GET", "/rest/artifacts", artifactsHandler)
    router.AddRoute("GET", "/rest/system/artifacts/initialize", initializeSystemArtifactsHandler)
    router.AddRoute("GET", "/rest/:system/artifacts", systemArtifactsHandler)

    //really need to get rid of these eventually
    router.AddRoute("GET", "/rest/jobs", jobsHandler)
    router.AddRoute("GET", "/rest/jobs/:job", jobBuildHandler)
    router.AddRoute("GET", "/rest/jobs/:job/:build" jobBuildHandler)

    router.AddRoute("POST", "/rest/build", initializeBuildHandler)

    router.AddRoute("GET", "/rest/test", testHandler)

    setSettings()
    router.Run()
}

var settings Settings

type Settings struct {
    BuildMasters []string `json:"BuildMasters"`
    JobFilters []string `json:"JobFilters"`
    Systems []string `json:"Systems"`
}

func setSettings() {
    b, err := ioutil.ReadFile("/home/jkwon/Git/releaseBuilder/settings.json")
    if err != nil {
        fmt.Println(err)
    }
       
    err = json.Unmarshal(b, &settings)
    if err != nil {
        fmt.Println(err)
    }
}

func jobsHandler(resp http.ResponseWriter, req *http.Request, parsedURL map[string]string) *r1util.AppError {

    b, err := json.Marshal(settings.JobFilters)
    if err != nil {
        return &r1util.AppError{err, "Error sending jobs: " + err.Error(), 500}
    }

    resp.Header().Set("Content-Type", "application/json")
    resp.Write(b)
    return nil
}

func jobBuildHandler(resp http.ResponseWriter, req *http.Request, parsedURL map[string]string) *r1util.AppError {

    job := parsedURL["job"]
    var port string
    var buildServer string
    
    var jobObj Job
    jobObj.Name = job
    

    isSbmJob := strings.Contains(job, "ServerBackup")
    if isSbmJob {
        buildServer = "sbci.do.r1soft.com"
        port = "8080"
    } else {
        buildServer = "ci.do.r1soft.com"
        port = ""
    }

    client, err := CreateRestClient(buildServer, port)
    if err != nil {
        r1util.LogError("Failed to create rest client")
        return &r1util.AppError{err, "Failed to create rest client: " + err.Error(), 500}
    } 

    result, err := client.Execute("/job/" + job + "/api/json", "GET")
    if err != nil {
        r1util.LogError("Failed to execute rest call")
        return &r1util.AppError{err, "Failed to execute rest call: " + err.Error(), 500}
    }

    var jobDetails map[string]interface{}
    err = json.Unmarshal(result, &jobDetails)

    // configure this in case we need to use windows and linux 32 and 64 bit, right now it
    // only uses 64 bit
    if isSbmJob {
        var buildNumList []int64
        activeConfigurations := reflect.ValueOf(jobDetails["activeConfigurations"])

        for i:=0; i<activeConfigurations.Len(); i++ {
            var tmp interface{} = activeConfigurations.Index(i).Interface()
            configList, ok := tmp.(map[string]interface{})
            if ok {
                config := configList["name"].(string)
                if strings.Contains(config, "linux-64") {

                    result, err := client.Execute("/job/" + job +"/" + config + "/api/json", "GET")
                    if err != nil {
                        r1util.LogError("Failed to execute rest call")
                        return &r1util.AppError{err, "Failed to execute rest call: " + err.Error(), 500}
                    }

                    var hello map[string]interface{}
                    err = json.Unmarshal(result, &hello)

                    tmpBuildList := reflect.ValueOf(hello["builds"])
                    for j:=0; j<tmpBuildList.Len(); j++ {
                        tmp = tmpBuildList.Index(j).Interface()
                        buildList, ok := tmp.(map[string]interface{})
                        if ok {
                            // don't ask me why...fix it later
                            floatNum := strconv.FormatFloat(buildList["number"].(float64), 'f', -1, 64)
                            intNum, _ := strconv.ParseInt(floatNum, 10, 64)
                            buildNumList = append(buildNumList, intNum)
                        }
                    }
                }
            }
        }
        jobObj.Builds = buildNumList
    }

    bytes, err := json.Marshal(jobObj)
    if err != nil {
        r1util.LogError("Failed to create JSON for complete job history")
        return &r1util.AppError{err, "Failed to create JSON for complete job history: " + err.Error(), 500}
    }

    resp.Header().Set("Content-Type", "application/json")
    resp.Write(bytes)
    
    return nil
}

func CreateRestClient(hostname string, port string) (RestClient, error) {
    var retVal RestClient

    retVal.Hostname = hostname
    if port != "" {
        retVal.Port = port
    }

    return retVal, nil
}

func (r RestClient) Execute(restUrl string, method string) ([]byte, error) {
    client := new(http.Client)
    req, err := http.NewRequest("", "", nil)
    if r.Port != "" {
        req, err = http.NewRequest(method, "http://" + r.Hostname + ":" + r.Port + restUrl, nil)
    } else {
        req, err = http.NewRequest(method, "http://" + r.Hostname + restUrl, nil)
    }

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

func systemsPackagesHandler(resp http.ResponseWriter, req *http.Request, parsedURL map[string]string) *r1util.AppError {

    system := parsedURL["system"]

    client, err := CreateRestClient("10.80.199.100", "9292")
    if err != nil {
        r1util.LogError("Failed to create rest client")
        return &r1util.AppError{err, "Failed to create rest client: " + err.Error(), 500}
    }

    result, err := client.Execute("/systems/" + system, "GET")
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

        jobHistory.JobName = jobObj.Name

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

                artifactObj = createArtifactObject(buildNum, artifact, job) 

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
 
        fileName := "/home/jkwon/Git/releaseBuilder/BuildHistory/" + jobHistory.JobName + ".BuildHistory.json"
        err = ioutil.WriteFile(fileName, bytes, 0644)
        if err != nil {
            r1util.LogError("Failed to create JSON for job history")
            return &r1util.AppError{err, "Failed to create JSON for job history: " + err.Error(), 500}
        }
    }

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

func createArtifactObject(buildNum string, artifactUrl string, jobName string) (Artifact) {
    splitUrl := strings.Split(artifactUrl, "/")
    // only artifact name
    tmpArtifact := splitUrl[len(splitUrl)-1]
    // strip '.deb'
    newTmpArtifact := strings.Replace(tmpArtifact, ".deb", "", -1)
   
    var tmpArtObj Artifact

    sbm := [2]string{"ServerBackup-6.0.x", "ServerBackup-6.2.x"}

    isSbm := false
    for i:=0; i<2; i++ {
        if strings.EqualFold(jobName, sbm[i]) {
            isSbm = true
        }
    } 

    //if (strings.Contains(jobName, "ServerBackup")) {
    if (isSbm) {
        // artifact name without '.deb', '-', '/'
        splitArtifactName := strings.Split(newTmpArtifact, "-")
        
        if splitArtifactName[0] == "idera" {
            tmpArtObj.Name = splitArtifactName[0] + "-" + splitArtifactName[1]
            tmpArtObj.Build = buildNum
            tmpArtObj.Version = splitArtifactName[3] + "-" + buildNum
        } else if splitArtifactName[1] == "docstore" {
            tmpArtObj.Build = buildNum
            tmpArtObj.Name = splitArtifactName[0] + "-" + splitArtifactName[1]
            tmpArtObj.Version = splitArtifactName[2] + "-" + buildNum
        } else {
            tmpArtObj.Build = splitArtifactName[len(splitArtifactName)-1]
            tmpArtObj.Version = splitArtifactName[len(splitArtifactName)-2] + "-" + tmpArtObj.Build
            splitArtifactName = splitArtifactName[:len(splitArtifactName)-3]
            for i:=0; i<len(splitArtifactName); i++ {
                if i == 0 {
                    tmpArtObj.Name = splitArtifactName[i]
                } else {
                    tmpArtObj.Name = tmpArtObj.Name + "-" + splitArtifactName[i]
                }
            }
        }
    } else /*if other jobs i.e. BBS, c247, r1scsi*/ {
        // artifact name without '.deb', '_', '/'
        if (strings.Contains(newTmpArtifact,"c247tools") || strings.Contains(newTmpArtifact,"c247mon")) {
            splitArtifactName := strings.Split(newTmpArtifact, "-")
            tmpArtObj.Name = splitArtifactName[0]
            tmpArtObj.Build = buildNum
            tmpArtObj.Version = splitArtifactName[1] + "-" + buildNum
        } else {
            splitArtifactName := strings.Split(newTmpArtifact, "_")
            tmpArtObj.Name = splitArtifactName[0]
            tmpArtObj.Version = splitArtifactName[1]
            tmpString := strings.Split(splitArtifactName[1], "-")
            tmpArtObj.Build = tmpString[1]
        }
    }
    
    tmpArtObj.Job = jobName
    tmpArtObj.Url = artifactUrl

    return tmpArtObj

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
    if err != nil {
        r1util.LogError("Failed to create system list")
        return &r1util.AppError{err, "Failed to create system list: " + err.Error(), 500}
    }
    
    var jobHistoryList []JobHistory 

    bytes, err := ioutil.ReadFile("/home/jkwon/Git/releaseBuilder/BuildHistory/CompleteList.json")
    if err != nil {
        r1util.LogError("Failed to read complete job history")
        return &r1util.AppError{err, "Failed to read complete job history: " + err.Error(), 500}
    }

    json.Unmarshal(bytes, &jobHistoryList)

    var artifactList []Artifact

    for _, system := range systemList {
        result, err := client.Execute("/systems/" + system, "GET")
        if err != nil {
            r1util.LogError("Failed to execute rest call")
            return &r1util.AppError{err, "Failed to execute rest call: " + err.Error(), 500}
        }

        var sysPackageList []string = make([]string, 0, 0)
        err = json.Unmarshal(result, &sysPackageList)
        if err != nil {
            r1util.LogError("Failed to system package list")
            return &r1util.AppError{err, "Failed to create system package list: " + err.Error(), 500}
        }

        for _, sysPackage := range sysPackageList {
            for _, job := range jobHistoryList {
                for _, buildHistory := range job.BuildHistory {
                    for _, artifact := range buildHistory.Artifacts {
                        if (strings.EqualFold(artifact.Name, sysPackage)) {
                            //fmt.Println(sysPackage + " " + job.JobName + " "+ artifact.Name + " " + artifact.Job + " " + artifact.Version)
                            tmp := exists(artifactList, artifact)
                            if !tmp {
                                //fmt.Println(sysPackage + " " + job.JobName + " "+ artifact.Name + " " + artifact.Job + " " + artifact.Version)
                                artifactList = append(artifactList, artifact)
                            }
                        }
                    }
                }
            }
        }
        
        bytes, err := json.Marshal(artifactList)
        if err != nil {
            r1util.LogError("Failed to create JSON for system artifact list")
            return &r1util.AppError{err, "Failed to create JSON for system artifact list: " + err.Error(), 500}
        }
        fileName := "/home/jkwon/Git/releaseBuilder/BuildHistory/" + system + ".Artifacts.json"
        err = ioutil.WriteFile(fileName, bytes, 0644)
    }

    return nil
}

func exists(artifactList []Artifact, artifact Artifact) bool {
    for _, tmpArt := range artifactList {
        // need to ask about this one.....
        /*
        if (strings.Contains(artifact.Job, "c247ufw-master") && strings.Contains(tmpArt.Job, "c247ufw-master") ) {
            if (artifact.Version == tmpArt.Version) {
                return true
            }
        }
        */
        if tmpArt == artifact {
            return true
        }
    }

    return false
}

func systemArtifactsHandler(resp http.ResponseWriter, req *http.Request, parsedURL map[string]string) *r1util.AppError {

    var artifactList []Artifact
    system := parsedURL["system"]

    bytes, err := ioutil.ReadFile("/home/jkwon/Git/releaseBuilder/BuildHistory/" + system + ".Artifacts.json")
    if err != nil {
        r1util.LogError("Failed to read system artifacts list")
        return &r1util.AppError{err, "Failed to read system artifacts list: " + err.Error(), 500}
    }

    json.Unmarshal(bytes, &artifactList)

    resp.Header().Set("Content-Type", "application/json")
    resp.Header().Set("Access-Control-Allow-Origin", "*")
    resp.Write(bytes)
    
    return nil
}

func initializeBuildHandler(resp http.ResponseWriter, req *http.Request, parsedURL map[string]string) *r1util.AppError {
    fmt.Printf("%v", req);
    bytes, err := ioutil.ReadAll(req.Body)
    if err != nil {
        r1util.LogError("Failed to execute rest call: " + err.Error())
    }
    fmt.Printf("%v", string(bytes))
    /*
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
    */
    return nil
}

func main() {
    port := flag.Int("port", 4030, "run at port")
    static_data = GetFileMap()
    startRouter(*port)
}
