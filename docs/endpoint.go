package docs

import (
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin/binding"
	"math"
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
			Content:  openapi3.NewContentWithSchema(typeToSchema(bodyType), []string{binding.MIMEJSON}),
		},
	}
}

func (e *Endpoint) AddResponse(responseType reflect.Type) {
	e.addResponse(responseType, 200)
}

func (e *Endpoint) AddErrorResponse(responseType reflect.Type) {
	e.addResponse(responseType, 500)
}

func (e *Endpoint) addResponse(responseType reflect.Type, defaultStatus int) {
	if len(e.Responses) == 0 {
		e.Responses = make(openapi3.Responses, 1)
	}

	schema := typeToSchema(responseType)
	response := &openapi3.ResponseRef{
		Value: &openapi3.Response{Content: openapi3.NewContentWithJSONSchema(schema)},
	}

	statusCode := strconv.Itoa(DefaultStatus(responseType, defaultStatus))

	codes := append(getStatusCodes(responseType), statusCode)
	for _, code := range codes {
		e.Responses[code] = response
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

func DefaultStatus(type_ reflect.Type, default_ ...int) int {
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
	if len(default_) > 0 {
		return default_[0]
	}
	return 200
}

func typeToSchema(type_ reflect.Type) *openapi3.Schema {
	type_ = directType(type_)

	model := reflect.New(type_).Elem().Interface()
	switch model.(type) {
	case int:
		return openapi3.NewIntegerSchema().WithMin(math.MinInt).WithMax(math.MaxInt)
	case int8:
		return openapi3.NewIntegerSchema().WithMin(math.MinInt8).WithMax(math.MaxInt8)
	case int16:
		return openapi3.NewIntegerSchema().WithMin(math.MinInt16).WithMax(math.MaxInt16)
	case int32:
		return openapi3.NewInt32Schema().WithMin(math.MinInt32).WithMax(math.MaxInt32)
	case int64:
		return openapi3.NewInt64Schema().WithMin(math.MinInt64).WithMax(math.MaxInt64)
	case uint:
		return openapi3.NewIntegerSchema().WithMin(0).WithMax(math.MaxUint)
	case uint8:
		return openapi3.NewIntegerSchema().WithMin(0).WithMax(math.MaxUint8)
	case uint16:
		return openapi3.NewIntegerSchema().WithMin(0).WithMax(math.MaxUint16)
	case uint32:
		return openapi3.NewInt32Schema().WithMin(0).WithMax(math.MaxUint32)
	case uint64:
		return openapi3.NewInt64Schema().WithMin(0).WithMax(math.MaxUint64)
	case string:
		return openapi3.NewStringSchema()
	case time.Time:
		return openapi3.NewDateTimeSchema()
	case float32, float64:
		return openapi3.NewFloat64Schema()
	case bool:
		return openapi3.NewBoolSchema()
	case []byte:
		return openapi3.NewBytesSchema()
	case *multipart.FileHeader:
		return openapi3.NewStringSchema().WithFormat("binary")
	case []*multipart.FileHeader:
		return openapi3.NewArraySchema().WithItems(&openapi3.Schema{
			Type:   "string",
			Format: "binary",
		})
	default:
		switch type_.Kind() {
		case reflect.Int:
			return openapi3.NewIntegerSchema().WithMin(math.MinInt).WithMax(math.MaxInt)
		case reflect.Int8:
			return openapi3.NewIntegerSchema().WithMin(math.MinInt8).WithMax(math.MaxInt8)
		case reflect.Int16:
			return openapi3.NewIntegerSchema().WithMin(math.MinInt16).WithMax(math.MaxInt16)
		case reflect.Int32:
			return openapi3.NewInt32Schema().WithMin(math.MinInt32).WithMax(math.MaxInt32)
		case reflect.Int64:
			return openapi3.NewInt64Schema().WithMin(math.MinInt64).WithMax(math.MaxInt64)
		case reflect.Uint:
			return openapi3.NewIntegerSchema().WithMin(0).WithMax(math.MaxUint)
		case reflect.Uint8:
			return openapi3.NewIntegerSchema().WithMin(0).WithMax(math.MaxUint8)
		case reflect.Uint16:
			return openapi3.NewIntegerSchema().WithMin(0).WithMax(math.MaxUint16)
		case reflect.Uint32:
			return openapi3.NewInt32Schema().WithMin(0).WithMax(math.MaxUint32)
		case reflect.Uint64:
			return openapi3.NewInt64Schema().WithMin(0).WithMax(math.MaxUint64)
		case reflect.Float32, reflect.Float64:
			return openapi3.NewFloat64Schema()
		case reflect.Struct:
			return structToSchema(type_)
		case reflect.Slice, reflect.Array:
			return openapi3.NewArraySchema().WithItems(typeToSchema(type_.Elem()))
		case reflect.Map:
			return openapi3.NewObjectSchema().WithAdditionalProperties(typeToSchema(type_.Elem()))
		case reflect.Interface:
			return openapi3.NewSchema().WithDefault("any")
		case reflect.String:
			return openapi3.NewStringSchema()
		case reflect.Bool:
			return openapi3.NewBoolSchema()
		default:
			panic(fmt.Sprintf("not allowed type or kind: %s(%s)", type_, type_.Kind()))
		}
	}
}

func structToSchema(type_ reflect.Type) *openapi3.Schema {
	schema := openapi3.NewObjectSchema()
	for i := 0; i < type_.NumField(); i++ {
		field := type_.Field(i)
		tags := field.Tag
		fieldSchema := typeToSchema(field.Type)

		fieldName, exists := tags.Lookup(jsonTag)
		if !exists || fieldName == "-" {
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
	return schema
}

func getStatusCodes(type_ reflect.Type) []string {
	var codes []string
	type_ = directType(type_)
	if type_.Kind() == reflect.Struct {
		for i := 0; i < type_.NumField(); i++ {
			field := type_.Field(i)
			if codesTag, exists := field.Tag.Lookup(statusCodesTag); exists {
				codes = append(codes, strings.Split(codesTag, ",")...)
			}
		}
	}
	return codes
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
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}
