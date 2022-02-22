package docs

import (
	"encoding/json"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-contrib/cors"
	"os"
	"regexp"
)

const (
	DefaultTag       = "default"
	BindingTag       = "binding"
	HeaderTag        = "header"
	JsonTag          = "json"
	DefaultStatusTag = "default_status"
	StatusCodesTag   = "status_codes"
)

type Docs struct {
	OpenAPIPath     string
	OpenAPI         *openapi3.T
	OpenAPIUrl      string
	Title           string
	TermsOfService  string
	Description     string
	License         *openapi3.License
	Contact         *openapi3.Contact
	Version         string
	OpenAPIFilePath string
	InMemory        bool
	CORS            *cors.Config
}

func (d *Docs) NewOpenAPI() {
	d.OpenAPI = &openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:          d.Title,
			Description:    d.Description,
			TermsOfService: d.TermsOfService,
			Contact:        d.Contact,
			License:        d.License,
			Version:        d.Version,
		},
	}
}

func (d *Docs) Valid() error {
	if !d.InMemory && d.OpenAPIFilePath == "" {
		return fmt.Errorf("invalid docs configuration, OpenAPIFilePath field cannot to be empty without InMemory option")
	}
	return nil
}

func (d *Docs) OpenAPIContent() *openapi3.T {
	if !d.InMemory {
		doc, err := openapi3.NewLoader().LoadFromFile(d.OpenAPIFilePath + "openapi.json")
		if err != nil {
			panic(fmt.Sprintf("unable to load openapi.json file (path: %s), err: %v", d.OpenAPIPath, err))
		}
		return doc
	}
	return d.OpenAPI
}

func (d *Docs) SetPath(path string, method string, doc *Endpoint) {
	existingPathItem := d.PathItem(d.FixPath(path))
	existingPathItem.SetOperation(method, (*openapi3.Operation)(doc))
	if d.PathsIsEmpty() {
		d.OpenAPI.Paths = make(openapi3.Paths)
	}
	d.OpenAPI.Paths[d.FixPath(path)] = existingPathItem
}

func (d *Docs) SetPathItem(path string, pathItem *openapi3.PathItem) {
	d.OpenAPI.Paths[d.FixPath(path)] = pathItem
}

func (d *Docs) AddServer(addr string) {
	d.OpenAPI.Servers = append(d.OpenAPI.Servers, &openapi3.Server{URL: addr})
}

func (d *Docs) Build() error {
	data, err := json.Marshal(d.OpenAPI)
	if err != nil {
		return err
	}
	// I assume the folder will be created, if it doesn't mean that there is already such a folder.
	_ = os.Mkdir(d.OpenAPIFilePath, 0755)
	return os.WriteFile(d.OpenAPIFilePath+"openapi.json", data, 0755)
}

func (d *Docs) PathsIsEmpty() bool {
	return len(d.OpenAPI.Paths) == 0
}

func (d *Docs) FixPath(path string) string {
	reg := regexp.MustCompile("/:([0-9a-zA-Z]+)")
	return reg.ReplaceAllString(path, "/{${1}}")
}

func (d *Docs) PathItem(path string) *openapi3.PathItem {
	pathItem := d.OpenAPI.Paths.Find(path)
	if pathItem == nil {
		return &openapi3.PathItem{}
	}
	return pathItem
}
