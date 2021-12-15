package docs

import (
	"github.com/fatih/structtag"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin/binding"
	"mime/multipart"
	"reflect"
	"regexp"
	"strings"
	"time"
)

const (
	DEFAULT         = "default"
	BINDING         = "binding"
	JSON            = "json"
	DefaultHttpCode = "default_http_code"
	ErrorCodes      = "error_codes"
)

type Helper struct {
	docs *Docs
}

func NewHelper(docs *Docs) *Helper {
	return &Helper{docs: docs}
}

func (h *Helper) FixPath(path string) string {
	reg := regexp.MustCompile("/:([0-9a-zA-Z]+)")
	return reg.ReplaceAllString(path, "/{${1}}")
}

func (h *Helper) GetPathItem(path string) *openapi3.PathItem {
	pathItem := h.docs.OpenAPI.Paths.Find(path)
	if pathItem == nil {
		return &openapi3.PathItem{}
	}
	return pathItem
}

func (h *Helper) AddParametersToOperation(params openapi3.Parameters, operation *openapi3.Operation) {
	operation.Parameters = append(operation.Parameters, params...)
}

func (h *Helper) getPathParams(path string) []string {
	reg := regexp.MustCompile(":([0-9a-zA-Z]+)")
	params := reg.FindAllString(path, -1)
	var newParams []string
	for _, param := range params {
		newParams = append(newParams, strings.ReplaceAll(param, ":", ""))
	}
	return newParams
}

func (h *Helper) GetTagsFromPath(path string) []string {
	var tags []string

	pathNormalWords := strings.Split(path, "/")
	for _, word := range pathNormalWords {
		if !strings.Contains(word, ":") && word != "" {
			tags = append(tags, word)
		}
	}

	return tags
}

func (h *Helper) ParsePathParams(path string) openapi3.Parameters {
	var params openapi3.Parameters

	pathParams := h.getPathParams(path)

	for _, param := range pathParams {
		params = append(params, &openapi3.ParameterRef{
			Value: &openapi3.Parameter{
				Name:            param,
				In:              "path",
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

	return params
}

func (h *Helper) ParseQueryParams(queryModel interface{}) openapi3.Parameters {
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

func (h *Helper) defaultModelSchema(model interface{}) *openapi3.Schema {
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
		schema = h.modelSchema(model)
	}
	return schema
}

func (h *Helper) modelSchema(model interface{}) *openapi3.Schema {
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
			fieldSchema := h.defaultModelSchema(value.Interface())
			bindingTag, err := tags.Get(BINDING)
			if err == nil {
				if bindingTag.Name == "required" {
					schema.Required = append(schema.Required, bindingTag.Name)
				}
			}
			defaultTag, err := tags.Get(DEFAULT)
			if err == nil {
				fieldSchema.Default = defaultTag.Name
			}

			tag, err := tags.Get(JSON)
			if err == nil {
				schema.Properties[tag.Name] = openapi3.NewSchemaRef("", fieldSchema)
			}
		}
	} else if type_.Kind() == reflect.Slice {
		schema = openapi3.NewArraySchema()
		schema.Items = &openapi3.SchemaRef{Value: h.modelSchema(reflect.New(type_.Elem()).Elem().Interface())}
	} else {
		schema = h.defaultModelSchema(model)
	}
	return schema
}

func (h *Helper) ConvertModelToRequestBody(model interface{}, contentType string) *openapi3.RequestBodyRef {
	body := &openapi3.RequestBodyRef{
		Value: openapi3.NewRequestBody(),
	}
	schema := h.modelSchema(model)
	body.Value.Required = true
	if contentType == "" {
		contentType = binding.MIMEJSON
	}
	body.Value.Content = openapi3.NewContentWithSchema(schema, []string{contentType})
	return body
}

func (h *Helper) getResponseDefaultHttpCode(model interface{}) string {
	type_ := reflect.TypeOf(model)
	value_ := reflect.ValueOf(model)
	if type_.Kind() == reflect.Ptr {
		type_ = type_.Elem()
	}
	if value_.Kind() == reflect.Ptr {
		value_ = value_.Elem()
	}
	if type_.Kind() == reflect.Struct {
		for i := 0; i < type_.NumField(); i++ {
			field := type_.Field(i)
			tags, err := structtag.Parse(string(field.Tag))
			if err != nil {
				panic(err)
			}
			defaultHttpCodeTag, err := tags.Get(DefaultHttpCode)
			if err == nil {
				return defaultHttpCodeTag.Name
			}
		}
	}
	return "200"
}

func (h *Helper) CreateResponses(model interface{}, errorModel interface{}) openapi3.Responses {
	ret := openapi3.NewResponses()
	schema := h.modelSchema(model)
	content := openapi3.NewContentWithJSONSchema(schema)
	ret[h.getResponseDefaultHttpCode(model)] = &openapi3.ResponseRef{
		Value: &openapi3.Response{
			Content: content,
		},
	}
	if errorModel != nil {
		errorSchema := h.modelSchema(errorModel)
		errorContent := openapi3.NewContentWithJSONSchema(errorSchema)
		errorCodes := h.getResponseErrorCodes(model)
		for _, code := range errorCodes {
			ret[code] = &openapi3.ResponseRef{
				Value: &openapi3.Response{Content: errorContent},
			}
		}
	}
	return ret
}

func (h *Helper) getResponseErrorCodes(model interface{}) []string {
	var errorCodes []string
	type_ := reflect.TypeOf(model)
	value_ := reflect.ValueOf(model)
	if type_.Kind() == reflect.Ptr {
		type_ = type_.Elem()
	}
	if value_.Kind() == reflect.Ptr {
		value_ = value_.Elem()
	}
	if type_.Kind() == reflect.Struct {
		for i := 0; i < type_.NumField(); i++ {
			field := type_.Field(i)
			tags, err := structtag.Parse(string(field.Tag))
			if err != nil {
				panic(err)
			}
			errorCodesTag, err := tags.Get(ErrorCodes)
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

func (h *Helper) ConvertTypeToInterface(t reflect.Type) interface{} {
	p := reflect.New(t)
	return p.Interface()
}

