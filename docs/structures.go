package docs

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/structtag"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin/binding"
	"mime/multipart"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultTag       = "default"
	BindingTag       = "binding"
	HeaderTag        = "header"
	JsonTag          = "json"
	DefaultStatusTag = "default_status"
	StatusCodesTag   = "status_codes"
)

type PathDoc openapi3.Operation

func (d *PathDoc) SetTagsFromPath(path string) {
	pathNormalWords := strings.Split(path, "/")
	for _, word := range pathNormalWords {
		if !strings.Contains(word, ":") && word != "" {
			d.Tags = append(d.Tags, word)
		}
	}
}

func (d *PathDoc) SetBodyType(bodyType reflect.Type) {
	d.RequestBody = &openapi3.RequestBodyRef{
		Value: &openapi3.RequestBody{
			Required: true,
			Content:  openapi3.NewContentWithSchema(modelSchema(bodyType), []string{binding.MIMEJSON}),
		},
	}
}

func (d *PathDoc) SetResponses(responseType reflect.Type, errorType reflect.Type) {
	if len(d.Responses) == 0 {
		d.Responses = openapi3.NewResponses()
	}

	delete(d.Responses, "default")
	schema := modelSchema(responseType)
	response := &openapi3.ResponseRef{
		Value: &openapi3.Response{Content: openapi3.NewContentWithJSONSchema(schema)},
	}

	defaultStatus := strconv.Itoa(DefaultStatus(responseType))

	codes := append(getStatusCodes(responseType), defaultStatus)
	for _, code := range codes {
		d.Responses[code] = response
	}

	if errorType != nil {
		errorSchema := modelSchema(errorType)
		errorContent := openapi3.NewContentWithJSONSchema(errorSchema)
		for _, code := range getStatusCodes(errorType) {
			d.Responses[code] = &openapi3.ResponseRef{
				Value: &openapi3.Response{Content: errorContent},
			}
		}
	}
}

func (d *PathDoc) SetQueryType(queryType reflect.Type) {
	queryType = directType(queryType)

	for i := 0; i < queryType.NumField(); i++ {
		if name := queryType.Field(i).Tag.Get("form"); name != "" {
			d.Parameters = append(d.Parameters, &openapi3.ParameterRef{
				Value: &openapi3.Parameter{
					Name: name,
					In:   "query",
				},
			})
		}
	}
}

func (d *PathDoc) AddPathParam(name string, type_ reflect.Type) {
	d.Parameters = append(d.Parameters, &openapi3.ParameterRef{
		Value: &openapi3.Parameter{
			Name:     name,
			In:       "path",
			Required: true,
			Schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: typeAsString(type_),
				},
			},
		},
	})
}

func (d *PathDoc) AddHeadersType(headerType reflect.Type) {
	headerType = directType(headerType)

	for i := 0; i < headerType.NumField(); i++ {
		tag := headerType.Field(i).Tag
		if name := tag.Get("header"); name != "" {
			d.Parameters = append(d.Parameters, &openapi3.ParameterRef{
				Value: &openapi3.Parameter{
					Name:     name,
					In:       HeaderTag,
					Required: strings.Contains(tag.Get(BindingTag), "required"),
				},
			})
		}
	}
}

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

func (d *Docs) SetPath(path string, method string, doc *PathDoc) {
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

func DefaultStatus(type_ reflect.Type) int {
	if type_.Kind() == reflect.Ptr {
		type_ = type_.Elem()
	}
	if type_.Kind() == reflect.Struct {
		for i := 0; i < type_.NumField(); i++ {
			field := type_.Field(i)
			strStatus, exists := field.Tag.Lookup(DefaultStatusTag)
			if exists {
				status, err := strconv.Atoi(strStatus)
				if err != nil {
					panic(fmt.Sprintf("cannot parse default http code to integer %v", err))
				}
				return status
			}
		}
	}
	return 200
}

func defaultModelSchema(type_ reflect.Type) *openapi3.Schema {
	model := reflect.New(type_).Elem().Interface()
	var schema *openapi3.Schema
	var m float64
	m = float64(0)
	switch model.(type) {
	case int, int8, int16:
		schema = openapi3.NewIntegerSchema()
	case uint, uint8, uint16:
		schema = openapi3.NewIntegerSchema()
		schema.Min = &m
	case int32:
		schema = openapi3.NewInt32Schema()
	case uint32:
		schema = openapi3.NewInt32Schema()
		schema.Min = &m
	case int64:
		schema = openapi3.NewInt64Schema()
	case uint64:
		schema = openapi3.NewInt64Schema()
		schema.Min = &m
	case string:
		schema = openapi3.NewStringSchema()
	case time.Time:
		schema = openapi3.NewDateTimeSchema()
	case float32, float64:
		schema = openapi3.NewFloat64Schema()
	case bool:
		schema = openapi3.NewBoolSchema()
	case []byte:
		schema = openapi3.NewBytesSchema()
	case *multipart.FileHeader:
		schema = openapi3.NewStringSchema()
		schema.Format = "binary"

	case []*multipart.FileHeader:
		schema = openapi3.NewArraySchema()
		schema.Items = &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type:   "string",
				Format: "binary",
			},
		}
	default:
		schema = modelSchema(type_)
	}
	return schema
}

func modelSchema(type_ reflect.Type) *openapi3.Schema {
	type_ = directType(type_)

	schema := openapi3.NewObjectSchema()
	switch type_.Kind() {
	case reflect.Struct:
		for i := 0; i < type_.NumField(); i++ {
			field := type_.Field(i)
			tags, err := structtag.Parse(string(field.Tag))
			if err != nil {
				panic(err)
			}
			fieldSchema := defaultModelSchema(field.Type)
			bindingTag, err := tags.Get(BindingTag)
			if err == nil {
				if bindingTag.Name == "required" {
					schema.Required = append(schema.Required, bindingTag.Name)
				}
			}
			defaultTag, err := tags.Get(DefaultTag)
			if err == nil {
				fieldSchema.Default = defaultTag.Name
			}

			tag, err := tags.Get(JsonTag)
			if err == nil {
				schema.Properties[tag.Name] = openapi3.NewSchemaRef("", fieldSchema)
			}
		}
	case reflect.Slice:
		schema = openapi3.NewArraySchema()
		schema.Items = &openapi3.SchemaRef{Value: modelSchema(type_.Elem())}
	case reflect.Map:
		schema.Items = &openapi3.SchemaRef{Value: modelSchema(type_.Elem())}
	case reflect.Interface:
		schema.Default = "any"
	default:
		schema = defaultModelSchema(type_)
	}
	return schema
}

func getStatusCodes(type_ reflect.Type) []string {
	var errorCodes []string
	type_ = directType(type_)
	if type_.Kind() == reflect.Struct {
		for i := 0; i < type_.NumField(); i++ {
			field := type_.Field(i)
			if errorCodesTag, exists := field.Tag.Lookup(StatusCodesTag); exists {
				errorCodes = append(errorCodes, strings.Split(errorCodesTag, ",")...)
			}
		}
	}
	return errorCodes
}

func typeAsString(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Int:
		return "integer"
	case reflect.String:
		return "string"
	default:
		panic(fmt.Sprintf("unknown type: %s in path params, must be integer or string", t))
	}
}

func directType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		return t.Elem()
	}
	return t
}
