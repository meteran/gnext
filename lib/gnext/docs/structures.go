package docs

import "github.com/getkin/kin-openapi/openapi3"

type Docs struct {
	OpenAPIPath        string
	OpenAPI            *openapi3.T
	OpenAPIUrl         string
	Title              string
	TermsOfService     string
	Description        string
	License            *openapi3.License
	Contact            *openapi3.Contact
	Version            string
}

func (d *Docs) PathsIsEmpty() bool {
	return len(d.OpenAPI.Paths) == 0
}

func (d *Docs) MarshalJSON() ([]byte, error) {
	return d.OpenAPI.MarshalJSON()
}