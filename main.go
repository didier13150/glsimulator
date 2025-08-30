package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var groups = []string{
	"CPP Language",
	"Go Language",
	"Perl Language",
	"Python Language",
	"PHP Language",
	"C Sharp Language",
}
var names = []string{
	"Wolfgang Amadeus Mozart",
	"Johann Sebastian Bach",
	"Ludwig Van Beethoven",
	"Johannes Brahms",
	"Giuseppe Verdi",
	"Piotr Ilitch Tchaïkovski",
	"Robert Schumann",
	"Franz Joseph Haydn",
}

type GLProject struct {
	Id                int    `json:"id"`
	Name              string `json:"name"`
	Description       any    `json:"description"`
	Path              string `json:"path"`
	NameWithNamespace string `json:"name_with_namespace"`
	PathWithNamespace string `json:"path_with_namespace"`
	SshUrlToRepo      string `json:"ssh_url_to_repo"`
	HttpUrlToRepo     string `json:"http_url_to_repo"`
	WebUrl            string `json:"web_url"`
	Visibility        string `json:"visibility"`
}

type GLVar struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description any    `json:"description"`
	Env         string `json:"environment_scope"`
	IsRaw       bool   `json:"raw"`
	IsHidden    bool   `json:"hidden"`
	IsProtected bool   `json:"protected"`
	IsMasked    bool   `json:"masked"`
}

type GLEnv struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	State       any    `json:"state"`
	Url         any    `json:"external_url"`
	Description any    `json:"description"`
}

func main() {
	http.HandleFunc("/", defaultdhandler)
	port := os.Getenv("APP_PORT")
	if len(port) == 0 {
		port = "8080"
	}
	log.Println("Listening on port " + port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalln(err)
	}
}

func defaultdhandler(w http.ResponseWriter, r *http.Request) {
	var payload []byte
	var err error
	var total int
	args := r.URL.Query()
	path := r.URL.Path
	log.Printf("Request: %s", path)

	previousPage := ""
	nextPage := ""
	per_page := 20
	page := 1
	if args["per_page"] != nil {
		per_page, _ = strconv.Atoi(args["per_page"][0])
	}
	if args["page"] != nil {
		page, _ = strconv.Atoi(args["page"][0])
	}

	pagination := false
	if len(path) >= 16 && path[:16] == "/api/v4/projects" {
		parts := strings.Split(path[1:], "/")
		kind := "projects"
		projectId := 0
		if len(parts) >= 4 {
			projectId, _ = strconv.Atoi(parts[3])
			// log.Printf("Project ID: %d", projectId)
		}
		if len(parts) >= 5 {
			kind = parts[4]
			// log.Printf("Kind: %s", kind)
		}
		switch kind {
		case "environments":
			envs := generateEnvList(projectId)
			total = len(envs)
			payload, err = json.Marshal(envs)
			if err != nil {
				log.Println(err)
			}
		case "variables":
			vars := generateVarList()
			total = len(vars)
			payload, err = json.Marshal(vars)
			if err != nil {
				log.Println(err)
			}
		case "projects":
			pagination = true
			//total = len(projects)
			total = len(groups) * len(names)
			min := 1 + ((page - 1) * per_page)
			max := per_page + ((page - 1) * per_page)
			if projectId != 0 {
				min = projectId
				max = projectId
				pagination = false
			}
			projects := generateProjectList(min, max)
			payload, err = json.Marshal(projects)
			if err != nil {
				log.Println(err)
			}
		default:
			payload = []byte(`{"error":"404 Not Found"}`)
		}
	} else {
		payload = []byte(`{"error":"404 Not Found"}`)
	}

	nbPage := int((total / per_page) + 1)
	pPage := page - 1
	if pPage <= 0 {
		pPage = 0
	}
	nPage := page + 1
	if nPage > nbPage {
		nPage = 0
	}

	if pPage > 0 {
		previousPage = strconv.Itoa(pPage)
	}
	if nPage > 0 {
		nextPage = strconv.Itoa(nPage)
	}
	if page > nbPage {
		page = nbPage
	}

	correlation := randomString(26, "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	duration := "0.8" + randomString(5, "0123456789")

	//w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Nel", "{\"max_age\": 0}")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
	w.Header().Set("Server", "nginx")
	w.Header().Set("Strict-transport-security", "max-age=31536000; preload")
	w.Header().Set("Vary", "Origin")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "SAMEORIGIN")
	w.Header().Set("X-Gitlab-Meta", "{\"correlation_id\":\""+correlation+"\",\"version\":\"1\"}")
	w.Header().Set("Link", "<http://localhost:8080/"+path+">")
	w.Header().Set("X-Request-Id", correlation)
	w.Header().Set("X-Runtime", duration)
	if pagination {
		w.Header().Set("X-Next-Page", nextPage)
		w.Header().Set("X-Prev-Page", previousPage)
		w.Header().Set("X-Total", strconv.Itoa(total))
		w.Header().Set("X-Total-Pages", strconv.Itoa(nbPage))
		w.Header().Set("X-Page", strconv.Itoa(page))
		w.Header().Set("X-Per-Page", strconv.Itoa(per_page))
	}
	written, err := w.Write(payload)
	if err != nil {
		log.Fatalln(err)
	}
	if written == 0 {
		log.Fatal("No data written")
	}
}

func randomString(length int, charset string) string {
	seededRand := rand.New(
		rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func generateProjectList(min int, max int) []GLProject {
	var projects []GLProject

	port := os.Getenv("APP_PORT")
	if len(port) == 0 {
		port = "8080"
	}
	// log.Printf("project min: %d, max %d\n", min, max)
	id := 0
	for _, group := range groups {
		groupPath := strings.ToLower(strings.ReplaceAll(group, " ", "_"))
		for _, name := range names {
			id++
			if id < min {
				continue
			}
			// log.Printf("generate project #%d\n", id)
			var project GLProject
			path := strings.ToLower(strings.ReplaceAll(name, " ", "_"))
			project.Id = id
			project.Name = name
			project.Path = path
			project.NameWithNamespace = group + " / " + name
			project.PathWithNamespace = groupPath + "/" + path
			project.SshUrlToRepo = "git@localhost:" + groupPath + "/" + path + ".git"
			project.HttpUrlToRepo = "http://localhost:" + port + "/" + groupPath + "/" + path + ".git"
			project.WebUrl = "http://localhost:" + port + "/" + groupPath + "/" + path
			project.Visibility = "public"
			projects = append(projects, project)
			if id >= max {
				break
			}
		}
		if id >= max {
			break
		}
	}
	return projects
}

func generateVarList() []GLVar {
	var vars []GLVar
	var envs = []string{"*", "Staging", "Production"}
	for _, env := range envs {
		var variable GLVar
		variable.Key = "DEBUG_ENABLED"
		variable.Value = "1"
		variable.Description = nil
		variable.Env = strings.ToLower(env)
		variable.IsRaw = true
		variable.IsHidden = false
		variable.IsMasked = false
		variable.IsProtected = false
		vars = append(vars, variable)
	}
	return vars
}

func generateEnvList(projectId int) []GLEnv {
	var environments []GLEnv
	var envs = []string{"Staging", "Production"}
	id := projectId*2 - 1
	for _, env := range envs {
		var environment GLEnv
		environment.Id = id
		environment.Name = env
		environment.State = "available"
		environment.Url = nil
		environment.Description = nil
		environments = append(environments, environment)
		id++
	}
	return environments
}
