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
	Default       = "default"
	Binding       = "binding"
	Json          = "json"
	DefaultStatus = "default_status"
	StatusCodes   = "status_codes"
)

type PathDoc openapi3.Operation

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

func (d *Docs) SetOperationOnPath(path string, method string, operation openapi3.Operation) {
	existingPathItem := d.PathItem(d.FixPath(path))
	existingPathItem.SetOperation(method, &operation)
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

func (d *Docs) AddParametersToOperation(params openapi3.Parameters, operation *openapi3.Operation) {
	operation.Parameters = append(operation.Parameters, params...)
}

func (d *Docs) PathTags(path string) []string {
	var tags []string

	pathNormalWords := strings.Split(path, "/")
	for _, word := range pathNormalWords {
		if !strings.Contains(word, ":") && word != "" {
			tags = append(tags, word)
		}
	}

	return tags
}

func (d *Docs) AddParamToOperation(paramName string, paramType reflect.Type, operation *openapi3.Operation) {
	var parameter openapi3.ParameterRef
	parameter.Value = &openapi3.Parameter{
		Name:     paramName,
		In:       "path",
		Required: true,
		Schema: &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type: d.typeAsString(paramType),
			},
		},
	}

	d.AddParametersToOperation(openapi3.Parameters{&parameter}, operation)
}

func (d *Docs) ParseHeaderParams(queryModel interface{}) openapi3.Parameters {
	var params openapi3.Parameters

	t := reflect.TypeOf(queryModel).Elem().Elem()
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Tag.Get("header") != "" {
			params = append(params, &openapi3.ParameterRef{
				Value: &openapi3.Parameter{
					Name:            t.Field(i).Tag.Get("header"),
					In:              "header",
					Description:     "",
					Style:           "",
					Explode:         nil,
					AllowEmptyValue: false,
					AllowReserved:   false,
					Deprecated:      false,
					Required:        false,
					Schema:          nil,
					Example:         nil,
					Examples:        nil,
					Content:         nil,
				},
			})
		}
	}

	return params
}


func (d *Docs) ParseQueryParams(queryModel interface{}) openapi3.Parameters {
	var params openapi3.Parameters

	t := reflect.TypeOf(queryModel).Elem()
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Tag.Get("form") != "" {
			params = append(params, &openapi3.ParameterRef{
				Value: &openapi3.Parameter{
					Name:            t.Field(i).Tag.Get("form"),
					In:              "query",
					Description:     "",
					Style:           "",
					Explode:         nil,
					AllowEmptyValue: false,
					AllowReserved:   false,
					Deprecated:      false,
					Required:        false,
					Schema:          nil,
					Example:         nil,
					Examples:        nil,
					Content:         nil,
				},
			})
		}
	}

	return params
}

func (d *Docs) ConvertModelToRequestBody(model interface{}, contentType string) *openapi3.RequestBodyRef {
	body := &openapi3.RequestBodyRef{
		Value: openapi3.NewRequestBody(),
	}
	schema := d.modelSchema(model)
	body.Value.Required = true
	if contentType == "" {
		contentType = binding.MIMEJSON
	}
	body.Value.Content = openapi3.NewContentWithSchema(schema, []string{contentType})
	return body
}

func (d *Docs) ResponseDefaultStatus(model interface{}) int {
	type_ := reflect.TypeOf(model)
	if type_.Kind() == reflect.Ptr {
		type_ = type_.Elem()
	}
	if type_.Kind() == reflect.Struct {
		for i := 0; i < type_.NumField(); i++ {
			field := type_.Field(i)
			tags, err := structtag.Parse(string(field.Tag))
			if err != nil {
				panic(err)
			}
			defaultHttpCodeTag, err := tags.Get(DefaultStatus)
			if err == nil {
				status, err := strconv.Atoi(defaultHttpCodeTag.Name)
				if err != nil {
					panic(fmt.Sprintf("cannot parse default http code to integer %v", err))
				}
				return status
			}
		}
	}
	return 200
}

func (d *Docs) CreateResponses(model interface{}, errorModel interface{}) openapi3.Responses {
	ret := openapi3.NewResponses()
	delete(ret, "default")
	schema := d.modelSchema(model)
	response := &openapi3.ResponseRef{
		Value: &openapi3.Response{Content: openapi3.NewContentWithJSONSchema(schema)},
	}

	defaultStatus := strconv.Itoa(d.ResponseDefaultStatus(model))

	codes := append(d.getStatusCodes(model), defaultStatus)
	for _, code := range codes {
		ret[code] = response
	}

	if errorModel != nil {
		errorSchema := d.modelSchema(errorModel)
		errorContent := openapi3.NewContentWithJSONSchema(errorSchema)
		errorCodes := d.getStatusCodes(errorModel)
		for _, code := range errorCodes {
			ret[code] = &openapi3.ResponseRef{
				Value: &openapi3.Response{Content: errorContent},
			}
		}
	}
	return ret
}

func (d *Docs) ConvertTypeToInterface(t reflect.Type) interface{} {
	p := reflect.New(t)
	return p.Interface()
}

func (d *Docs) defaultModelSchema(model interface{}) *openapi3.Schema {
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
		schema = d.modelSchema(model)
	}
	return schema
}

func (d *Docs) modelSchema(model interface{}) *openapi3.Schema {
	type_ := reflect.TypeOf(model)
	value_ := reflect.ValueOf(model)
	schema := openapi3.NewObjectSchema()
	if type_.Kind() == reflect.Ptr {
		type_ = type_.Elem()
	}
	if value_.Kind() == reflect.Ptr {
		value_ = value_.Elem()
	}
	if type_.Kind() == reflect.Struct {
		for i := 0; i < type_.NumField(); i++ {
			field := type_.Field(i)
			value := value_.Field(i)
			tags, err := structtag.Parse(string(field.Tag))
			if err != nil {
				panic(err)
			}
			if !value.CanInterface() {
				continue
			}
			fieldSchema := d.defaultModelSchema(value.Interface())
			bindingTag, err := tags.Get(Binding)
			if err == nil {
				if bindingTag.Name == "required" {
					schema.Required = append(schema.Required, bindingTag.Name)
				}
			}
			defaultTag, err := tags.Get(Default)
			if err == nil {
				fieldSchema.Default = defaultTag.Name
			}

			tag, err := tags.Get(Json)
			if err == nil {
				schema.Properties[tag.Name] = openapi3.NewSchemaRef("", fieldSchema)
			}
		}
	} else if type_.Kind() == reflect.Slice {
		schema = openapi3.NewArraySchema()
		schema.Items = &openapi3.SchemaRef{Value: d.modelSchema(reflect.New(type_.Elem()).Elem().Interface())}
	} else {
		schema = d.defaultModelSchema(model)
	}
	return schema
}

func (d *Docs) getStatusCodes(model interface{}) []string {
	var errorCodes []string
	type_ := reflect.TypeOf(model)
	if type_.Kind() == reflect.Ptr {
		type_ = type_.Elem()
	}
	if type_.Kind() == reflect.Struct {
		for i := 0; i < type_.NumField(); i++ {
			field := type_.Field(i)
			tags, err := structtag.Parse(string(field.Tag))
			if err != nil {
				panic(err)
			}
			errorCodesTag, err := tags.Get(StatusCodes)
			if err == nil {
				codes := strings.Split(errorCodesTag.Value(), ",")
				for _, code := range codes {
					errorCodes = append(errorCodes, code)
				}
			}
		}
	}
	return errorCodes
}

func (d *Docs) cleanPathParam(param string) string {
	wColon := strings.ReplaceAll(param, ":", "")
	wSlash := strings.ReplaceAll(wColon, "/", "")
	return wSlash
}

func (d *Docs) CORSConfig() *cors.Config {
	if d.CORS == nil {
		return &cors.Config{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"*"},
			AllowHeaders:     []string{"*"},
			AllowCredentials: true,
		}
	}
	return d.CORS
}

func (d *Docs) typeAsString(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Int:
		return "integer"
	case reflect.String:
		return "string"
	default:
		panic(fmt.Sprintf("unknown type: %v in path params, must be integer or string", t.Kind()))
	}
}


