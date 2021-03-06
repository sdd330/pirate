package controllers

/*
 * The docker API controller to access docker unix socket and reponse JSON data
 *
 * Refer to https://docs.docker.com/reference/api/docker_remote_api_v1.14/ for docker remote API
 * Refer to https://github.com/Soulou/curl-unix-socket to know how to access unix docket
 */

import (
	"github.com/astaxie/beego"
    "os"
	"fmt"
	"io/ioutil"
	"net/http"
	"errors"
	"strings"
	"encoding/json"
	"os/exec"
	"regexp"
)

/* Give address and method to request docker unix socket */
func RequestRegistry(address, method string) string {
	REGISTRY_URL := "registry:5000"
	if os.Getenv("REGISTRY_URL") != "" {
	   REGISTRY_URL = os.Getenv("REGISTRY_URL")
	} else {
	   // set varible to be used in configuration
	   os.Setenv("REGISTRY_URL", REGISTRY_URL)
	}
    // fmt.Println(os.Environ())
	registry_url := "http://" + REGISTRY_URL + "/v1" + address
    //fmt.Println(registry_url)

	reader := strings.NewReader("")

	request, err := http.NewRequest(method, registry_url, reader)
	if err != nil {
		fmt.Println("Error to create http request", err)
		return ""
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		fmt.Println("Error to achieve http request over unix socket", err)
		return ""
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error, get invalid body in answer")
		return ""
	}
	//fmt.Println(body)

	/* An example to get continual stream from events, but it's for stdout
		_, err = io.Copy(os.Stdout, res.Body)
		if err != nil && err != io.EOF {
			fmt.Println("Error, get invalid body in answer")
			return ""
	   }
	*/

	defer response.Body.Close()
	// fmt.Println("http result body:",string(body))
	return string(body)
}

/* It's a beego controller */
type DockerregistryapiController struct {
	beego.Controller
}

type image struct {
    Name string
    Id string
    Tag string
}

type repository struct {
    Description string
    Name string
}

type search struct {
    Num_results int
    Query string
    Results []repository
}

func getTags(name string) string {
	address := "/repositories/" + name + "/tags"
	fmt.Println(address)
	result := RequestRegistry(address, "GET")
    return result
}

func getAncestry(id string) string {
	address := "/images/" + id + "/ancestry"
	fmt.Println(address)
	result := RequestRegistry(address, "GET")
    return result
}


const README_MD_FILE = "app/README.md"
const DOCKERFILE_FILE = "app/Dockerfile"
const BUILD_LOG_FILE = "app/BUILD.log"
const PIRATE_INI_FILE = "app/PIRATE.ini"
const REGISTRY_PATH = "/registry/images"

func readFileFromTar(filelist string, tarfile string, extracted_file string)(string, error) {
   
    if strings.Index(string(filelist),extracted_file) != -1 {
        fmt.Printf("found the file %s\n", extracted_file)
        cmd := fmt.Sprintf("tar -xOf %s %s",tarfile, extracted_file)
        dat1, err1 := exec.Command("sh","-c",cmd).Output()
        return string(dat1),err1
   }
   return "",errors.New("can't find the file")
}

// ini part is referred https://github.com/vaughan0/go-ini/blob/master/ini.go
var assignRegex  = regexp.MustCompile(`^([^=]+)=(.*)$`)

func getIni(ini string)(map[string]string) {
    var abc=map[string]string{}
    lines := strings.Split(ini, "\n")
    for i, line := range lines {
		fmt.Println(i, line)
				line = strings.TrimSpace(line)
		if len(line) == 0 {
			// Skip blank lines
			continue
		}
		if line[0] == ';' || line[0] == '#' {
			// Skip comments
			continue
		}

		if groups := assignRegex.FindStringSubmatch(line); groups != nil {
			key, val := groups[1], groups[2]
			key, val = strings.TrimSpace(key), strings.TrimSpace(val)
			abc[key] = val
		}

    }
    return abc
}

func getReadme(id string)(string,string,string,string,error) {
    layerfile := REGISTRY_PATH + "/" + id + "/layer"
	
	cmd := "tar -tf "+ layerfile
    out, err := exec.Command("sh","-c",cmd).Output()
    if err != nil {
       fmt.Printf("shall not happen !! %s", err)
    } 
   
    // fmt.Printf("========================  %s", out)
	
    dat1, err1 := readFileFromTar(string(out), layerfile,README_MD_FILE)
	dat2, err2 := readFileFromTar(string(out), layerfile,DOCKERFILE_FILE)
	dat3, err3 := readFileFromTar(string(out), layerfile,BUILD_LOG_FILE)
	dat4, err4 := readFileFromTar(string(out), layerfile,PIRATE_INI_FILE)
	
	if err1 != nil && err2 != nil && err3 != nil && err4 != nil {
		return "","","","",errors.New("can't find the file")
		
	} 
    return string(dat1),string(dat2),string(dat3),string(dat4),nil
}

/* Wrap docker remote API to get images */
func (this *DockerregistryapiController) GetImages() {
	// https://github.com/docker/docker-registry/issues/63
	
    // search for all repository
	address := "/search"	
	result := RequestRegistry(address, "GET")
	fmt.Println("result:",result)
	
	var searchResult search
    json.Unmarshal([]byte(result),&searchResult)
    fmt.Println(searchResult)
    fmt.Println(searchResult.Results)
  
    var f interface{}
    images := make([]image,0)
    var oneimage image
    for _, repo := range searchResult.Results {
        fmt.Println(repo)
        tags := getTags(repo.Name)
        json.Unmarshal([]byte(tags), &f)
        m := f.(map[string]interface{})
        for k, v := range m {
          fmt.Println("tag name:",k,",id:",v)
          oneimage.Name =  repo.Name
          oneimage.Tag = k
          oneimage.Id = v.(string)
          images = append(images,oneimage)
        }
    }
    fmt.Println(images)
    all,_ := json.Marshal(images)
    fmt.Println(string(all))
	
	this.Ctx.WriteString(string(all))
}

/* Wrap docker remote API to get data of image */
func (this *DockerregistryapiController) GetImage() {
	id := this.GetString(":id")
	address := "/images/" + id + "/json"
	result := RequestRegistry(address, "GET")
	this.Ctx.WriteString(result)
}

/* Wrap docker remote API to get data of user image */
func (this *DockerregistryapiController) GetUserImage() {
	user := this.GetString(":user")
	repo := this.GetString(":repo")
	address := "/repositories/" + user + "/" + repo + "/tags"
	fmt.Println(address)
	result := RequestRegistry(address, "GET")
	this.Ctx.WriteString(result)
}

/* Wrap docker remote API to delete image */
func (this *DockerregistryapiController) DeleteImage() {
	name := this.GetString(":name")
	repo := this.GetString(":repo")
	tag := this.GetString(":tag")
	address := "/repositories/" + name + "/" + repo + "/tags/" + tag
	var result string
	if  os.Getenv("PIRATE_MODE") == "readonly" {
	    fmt.Println("readonly mode, delete is not allowed")
	    result = "" 
	} else {
		result = RequestRegistry(address, "DELETE")
	}
	this.Ctx.WriteString(result)
}

type version struct {
    ApiVersion string 
	Arch string
	GitCommit string
	GoVersion string
	KernelVersion string
	Os string
	Version string
}

type configuration struct {
    KernelVersion string
	RegistryServer string
	PirateMode string
	PirateUrlAlias string
    Os string
    Version string
	ApiVersion string
	GoVersion string
	Arch string
	GitCommit string
	Host string
	Launch string
}

type ping struct {
    Host []string
    Launch []string
	Versions interface{}
}

/* Wrap docker registry API to get version info */
func (this *DockerregistryapiController) GetVersion() {
	address := "/_ping"
	result := RequestRegistry(address, "GET")
	
	// unmarshall the docker ping result to internal data
    var pingResult ping
    json.Unmarshal([]byte(result),&pingResult)
	m := pingResult.Versions.(map[string]interface{})

/*	
    fmt.Println(pingResult)
    fmt.Println("Host:",pingResult.Host)
	fmt.Println("Host:",pingResult.Host[2])
	fmt.Println("Launch:",pingResult.Launch)
	fmt.Println("Versions:",pingResult.Versions)
*/
    // fill the internal data structure
	var config configuration
//	config.Host          = pingResult.Host
//	config.Launch        = pingResult.Launch
//	config.Versions      = pingResult.Versions
	config.Os            = pingResult.Host[0]
	config.GitCommit     = pingResult.Host[1]
	config.KernelVersion = pingResult.Host[2]	
	config.Version       = m["docker_registry.server"].(string)
	
	// add pirate environment
	config.RegistryServer = os.Getenv("REGISTRY_URL")
	config.PirateMode = os.Getenv("PIRATE_MODE")
	config.PirateUrlAlias = os.Getenv("PIRATE_URL_ALIAS")
	if(config.PirateUrlAlias == ""){
		config.PirateUrlAlias = config.RegistryServer
	}

    configJson,_ := json.Marshal(config)
    // fmt.Println("config:",string(configJson))
	
	this.Ctx.WriteString(string(configJson))
}

/* Wrap docker remote API to get docker info */
func (this *DockerregistryapiController) GetInfo() {
	address := "/_ping"
	result := RequestRegistry(address, "GET")
	this.Ctx.WriteString(result)
}

type imageInfo struct {
    Id string
	ParentId string
    Name string
	Tag string
	Readme string
	Layers string
	BuildInfo string
	Dockerfile string
	Author string
	Architecture string
	Created string
	Comment string
	DockerVersion string
	Os string
	Size string
	PirateFile string
	Tags string
	Pirate2 map[string]string
	Layers2 []string
    Tags2 interface{}
	
}

const DOCKERHUB_URL="https://registry.hub.docker.com/u"
/* Wrap docker remote API to get data of image */
func (this *DockerregistryapiController) GetImageInfo() {
	var id, name, tag string

	this.Ctx.Input.Bind(&id, "id")
	this.Ctx.Input.Bind(&name, "name")
	this.Ctx.Input.Bind(&tag, "tag")
	fmt.Printf("id:%s,name:%s,tag:%s", id, name, tag)
	
	// send message for image
	address := "/images/" + id + "/json"
	result := RequestRegistry(address, "GET")
	
	var objmap map[string]json.RawMessage
	json.Unmarshal([]byte(result), &objmap)

	var info imageInfo

	info.ParentId = string(objmap["parent"])	
	info.Architecture = string(objmap["architecture"])	
	info.Created=string(objmap["created"])
	info.Author = string(objmap["author"])
	info.Os = string(objmap["os"])
	info.Comment = ""
	info.DockerVersion = string(objmap["docker_version"])
	info.Size = string(objmap["Size"])
	
    readme := ""
	dockerfile := ""
	buildlogfile := ""
	piratefile := ""
	var parentId []string
    for i:=0;; {      // i is used to control the depth of the layer
        fmt.Println("Check id:" + id)
        dat1, dat2,dat3,dat4, err := getReadme(id)
        if err == nil || i > 5 {
            // readme = contents
			readme = dat1
			dockerfile = dat2
			buildlogfile = dat3
			piratefile= dat4
            break  // found README
        }
        i = i+1 
        address := "/images/" + id + "/json"
        result := RequestRegistry(address, "GET")
        i := strings.Index(result,`"parent":`)
        if i==-1 {
            break // to the end
        }
        id = result[i+10:i+74] // TODO: hacked solution to get parent
		parentId = append(parentId, id)
        //fmt.Println(parentId)
    }
    if readme == "" {
	    url := DOCKERHUB_URL + "/" + name 
	    readme = fmt.Sprintf("Not find in docker image, mostly you could try [%s](%s)",url,url)
    }
	
	if dockerfile == "" {
	    url := DOCKERHUB_URL + "/" + name + "/dockerfile"
	    dockerfile = fmt.Sprintf("Not found in docker image, mostly you want to check %s",url)
	}
    
    if buildlogfile == "" {
	    buildlogfile ="no build log is attached"
	}	

    info.Readme = readme
	info.Id = id
	info.Name = name
	info.Tag = tag
	info.BuildInfo = buildlogfile
	info.Dockerfile = dockerfile
	info.PirateFile = piratefile
	
	tags := getTags(name)
	info.Tags = tags

	layers := getAncestry(id)
	//fmt.Println("layers:", layers)
	info.Layers = layers
	
	// http://play.golang.org/p/6b1buUfE7y
	var f interface{}
	json.Unmarshal([]byte(tags), &f)
    info.Tags2 = f.(map[string]interface{})

	var f2 []string
	json.Unmarshal([]byte(layers), &f2)
	
	info.Layers2 = f2
	
	info.Pirate2 = getIni(piratefile)
	
    all,_ := json.MarshalIndent(info,"","  ")
    fmt.Println(string(all))
	
	this.Ctx.WriteString(string(all))
}

