package docs

import (
	"github.com/getkin/kin-openapi/openapi3"
)

const (
	DefaultTag       = "default"
	BindingTag       = "binding"
	HeaderTag        = "header"
	JsonTag          = "json"
	DefaultStatusTag = "default_status"
	StatusCodesTag   = "status_codes"
	NoUrl            = "-"
)

type Options struct {
	// ExtensionProps OpenAPI extensions.
	// It reads/writes all properties with prefix "x-".
	// It is empty as default.
	ExtensionProps openapi3.ExtensionProps

	// Title of the documentation.
	// If not set, the default value is "Documentation".
	Title string

	// Description of the documentation.
	// It is empty as default.
	Description string

	// TermsOfService usually should contain the URL to terms of service.
	// It is empty as default.
	TermsOfService string

	// Contact information must be in openAPI format.
	// It is empty as default.
	Contact *openapi3.Contact

	// License information must be in openAPI format.
	// It is empty as default.
	License *openapi3.License

	// Version of the documentation.
	// If not set, the default value is "1.0.0".
	Version string

	// InteractiveUrl is the path where your interactive documentation will be placed. It must be an absolute path, started with "/".
	// If you run the server locally, then your interactive docs will be under "http://localhost:<port><InteractiveUrl>".
	// If set to NoUrl, the interactive documentation will not be served.
	// If not set, the default value is "/docs".
	//
	// Interactive documentation uses the JSON file, thus it requires the JsonUrl is set to a valid url.
	// If the JsonUrl is set to NoUrl, the interactive documentation will be disabled.
	InteractiveUrl string

	// JsonUrl is the path where your openAPI in JSON format will be placed. It must be an absolute path, started with "/".
	// If you run the server locally, then the file will be under "http://localhost:<port><JsonUrl>".
	// If set to NoUrl, the json file will not be served.
	// If not set, the default value is "/docs.json".
	JsonUrl string

	// YamlUrl is the path where your openAPI in YAML format will be placed. It must be an absolute path, started with "/".
	// If you run the server locally, then the file will be under "http://localhost:<port><YamlUrl>".
	// If set to NoUrl, the yaml file will not be served.
	// If not set, the default value is "/docs.yaml".
	YamlUrl string

	// Servers is the list of the API locations. In interactive documentation (see InteractiveUrl) you can try your API out using one of those servers.
	// In case it is an empty list, it will be empty.
	// In case it is nil, the default value is a list with one element "http://localhost:8080".
	Servers []string

	// Tags for the documentation.
	// It is empty as default.
	Tags []string
}

var defaultOptions = &Options{
	Title:          "Documentation",
	Version:        "1.0.0",
	InteractiveUrl: "/docs",
	JsonUrl:        "/docs.json",
	YamlUrl:        "/docs.yaml",
	Servers:        []string{"http://localhost:8080"},
}
