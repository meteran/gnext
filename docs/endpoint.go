package docs

import (
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin/binding"
	"mime/multipart"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Endpoint openapi3.Operation

func (e *Endpoint) SetTagsFromPath(path string) {
	pathNormalWords := strings.Split(path, "/")
	for _, word := range pathNormalWords {
		if !strings.Contains(word, ":") && word != "" {
			e.Tags = append(e.Tags, word)
		}
	}
}

func (e *Endpoint) SetBodyType(bodyType reflect.Type) {
	e.RequestBody = &openapi3.RequestBodyRef{
		Value: &openapi3.RequestBody{
			Required: true,
			Content:  openapi3.NewContentWithSchema(modelSchema(bodyType), []string{binding.MIMEJSON}),
		},
	}
}

func (e *Endpoint) SetResponses(responseType reflect.Type, errorType reflect.Type) {
	if len(e.Responses) == 0 {
		e.Responses = openapi3.NewResponses()
	}

	delete(e.Responses, "default")
	schema := modelSchema(responseType)
	response := &openapi3.ResponseRef{
		Value: &openapi3.Response{Content: openapi3.NewContentWithJSONSchema(schema)},
	}

	defaultStatus := strconv.Itoa(DefaultStatus(responseType))

	codes := append(getStatusCodes(responseType), defaultStatus)
	for _, code := range codes {
		e.Responses[code] = response
	}

	if errorType != nil {
		errorSchema := modelSchema(errorType)
		errorContent := openapi3.NewContentWithJSONSchema(errorSchema)
		for _, code := range getStatusCodes(errorType) {
			e.Responses[code] = &openapi3.ResponseRef{
				Value: &openapi3.Response{Content: errorContent},
			}
		}
	}
}

func (e *Endpoint) SetQueryType(queryType reflect.Type) {
	queryType = directType(queryType)

	for i := 0; i < queryType.NumField(); i++ {
		if name := queryType.Field(i).Tag.Get("form"); name != "" {
			e.Parameters = append(e.Parameters, &openapi3.ParameterRef{
				Value: &openapi3.Parameter{
					Name: name,
					In:   "query",
				},
			})
		}
	}
}

func (e *Endpoint) AddPathParam(name string, type_ reflect.Type) {
	e.Parameters = append(e.Parameters, &openapi3.ParameterRef{
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

func (e *Endpoint) AddHeadersType(headerType reflect.Type) {
	headerType = directType(headerType)

	for i := 0; i < headerType.NumField(); i++ {
		tag := headerType.Field(i).Tag
		if name := tag.Get("header"); name != "" {
			e.Parameters = append(e.Parameters, &openapi3.ParameterRef{
				Value: &openapi3.Parameter{
					Name:     name,
					In:       headerTag,
					Required: strings.Contains(tag.Get(bindingTag), "required"),
				},
			})
		}
	}
}

func DefaultStatus(type_ reflect.Type) int {
	if type_.Kind() == reflect.Ptr {
		type_ = type_.Elem()
	}
	if type_.Kind() == reflect.Struct {
		for i := 0; i < type_.NumField(); i++ {
			field := type_.Field(i)
			strStatus, exists := field.Tag.Lookup(defaultStatusTag)
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
			tags := field.Tag
			fieldSchema := defaultModelSchema(field.Type)

			fieldName, exists := tags.Lookup(jsonTag)
			if !exists {
				continue
			}

			schema.Properties[fieldName] = openapi3.NewSchemaRef("", fieldSchema)

			bindingTag, exists := tags.Lookup(bindingTag)
			if exists {
				for _, validation := range strings.Split(bindingTag, ",") {
					if validation == "required" {
						schema.Required = append(schema.Required, fieldName)
					}
				}
			}

			defaultValue, exists := tags.Lookup(defaultTag)
			if exists {
				fieldSchema.Default = defaultValue
			}
		}
	case reflect.Slice:
		schema = openapi3.NewArraySchema()
		schema.Items = &openapi3.SchemaRef{Value: modelSchema(type_.Elem())}
	case reflect.Map:
		schema.AdditionalProperties = &openapi3.SchemaRef{Value: modelSchema(type_.Elem())}
	case reflect.Interface:
		schema = openapi3.NewSchema()
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
			if errorCodesTag, exists := field.Tag.Lookup(statusCodesTag); exists {
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
