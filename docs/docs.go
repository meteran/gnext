package docs

import (
	"encoding/json"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
	"os"
	"regexp"
)

func New(options *Options) *Docs {
	if options.Title == "" {
		options.Title = defaultOptions.Title
	}
	if options.Version == "" {
		options.Version = defaultOptions.Version
	}
	if options.InteractiveUrl == "" {
		options.InteractiveUrl = defaultOptions.InteractiveUrl
	}
	if options.JsonUrl == "" {
		options.JsonUrl = defaultOptions.JsonUrl
	}
	if options.YamlUrl == "" {
		options.YamlUrl = defaultOptions.YamlUrl
	}
	if options.Servers == nil {
		options.Servers = defaultOptions.Servers
	}

	servers := openapi3.Servers{}
	for _, url := range options.Servers {
		servers = append(servers, &openapi3.Server{URL: url})
	}

	tags := openapi3.Tags{}
	for _, tag := range options.Tags {
		tags = append(tags, &openapi3.Tag{Name: tag})
	}

	d := Docs{
		OpenApi: &openapi3.T{
			OpenAPI: "3.0.0",
			Info: &openapi3.Info{
				Extensions:     options.ExtensionProps,
				Title:          options.Title,
				Description:    options.Description,
				TermsOfService: options.TermsOfService,
				Contact:        options.Contact,
				License:        options.License,
				Version:        options.Version,
			},
			Components: options.Components,
			Security:   options.Security,
			Paths:      make(openapi3.Paths),
			Servers:    servers,
			Tags:       tags,
		},
		InteractiveUrl: options.InteractiveUrl,
		JsonUrl:        options.JsonUrl,
		YamlUrl:        options.YamlUrl,
	}
	return &d
}

type Docs struct {
	OpenApi        *openapi3.T
	InteractiveUrl string
	JsonUrl        string
	YamlUrl        string
}

func (d *Docs) SetPath(path string, method string, doc *Endpoint) {
	existingPathItem := d.PathItem(d.NormalizePath(path))
	existingPathItem.SetOperation(method, (*openapi3.Operation)(doc))
	d.OpenApi.Paths[d.NormalizePath(path)] = existingPathItem
}

func (d *Docs) SaveAsJson(path string) error {
	data, err := json.Marshal(d.OpenApi)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (d *Docs) SaveAsYaml(path string) error {
	data, err := yaml.Marshal(d.OpenApi)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (d *Docs) NormalizePath(path string) string {
	reg := regexp.MustCompile("/:([0-9a-zA-Z]+)")
	return reg.ReplaceAllString(path, "/{${1}}")
}

func (d *Docs) PathItem(path string) *openapi3.PathItem {
	pathItem := d.OpenApi.Paths.Find(path)
	if pathItem == nil {
		return &openapi3.PathItem{}
	}
	return pathItem
}

func (d *Docs) RegisterRoutes(router gin.IRouter) {
	handler := NewHandler(d)

	if d.YamlUrl != NoUrl {
		router.GET(d.YamlUrl, handler.YamlFile)
	}
	if d.JsonUrl != NoUrl {
		router.GET(d.JsonUrl, handler.JsonFile)
		if d.InteractiveUrl != NoUrl {
			router.GET(d.InteractiveUrl, handler.Docs)
		}
	}
}
