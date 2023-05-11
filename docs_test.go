package gnext

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/meteran/gnext/docs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestInteractiveDocsHandlerReturnsSubstitutedHtml(t *testing.T) {
	interactiveUrl := "/path/to/interactive/docs"

	r := Router(&docs.Options{
		Title:          "Test",
		JsonUrl:        "/path/to/api.json",
		InteractiveUrl: interactiveUrl,
	})

	r.Docs.RegisterRoutes(r.rawRouter)

	response := makeRequest(t, r, http.MethodGet, interactiveUrl)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <title>Test - Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@3/swagger-ui.css">
    <script src="https://unpkg.com/swagger-ui-dist@3/swagger-ui-bundle.js" charset="UTF-8"></script>
</head>
<body>
<div id="swagger-ui"></div>
<script>
    const ui = SwaggerUIBundle({
        url: "\/path\/to\/api.json",
        dom_id: '#swagger-ui',
        presets: [
            SwaggerUIBundle.presets.apis,
        ],
        persistAuthorization: true,
    })
</script>
</body>
</html>`, response.Body.String())
}

func TestDocsTags(t *testing.T) {
	handler := func() string {
		return "Hello World!"
	}
	r := Router()
	r.POST("/my/example", handler, &docs.Endpoint{Tags: []string{}})
	r.Group("/my/shops").
		GET("/list", handler, &docs.Endpoint{Tags: []string{"shops"}}).
		GET("/shop/:name/", handler)

	r.Docs.RegisterRoutes(r.rawRouter)
	response := makeRequest(t, r, http.MethodGet, "/docs.json")
	assert.Equal(t, http.StatusOK, response.Code)
	doc, err := openapi3.NewLoader().LoadFromData(response.Body.Bytes())

	assert.Equal(t, nil, err)

	assert.Equal(t, []string(nil), doc.Paths["/my/example"].Post.Tags)
	assert.Equal(t, []string{"shops"}, doc.Paths["/my/shops/list"].Get.Tags)
	assert.Equal(t, []string{"my", "shops", "shop"}, doc.Paths["/my/shops/shop/{name}/"].Get.Tags)
}

func generateDocs(t *testing.T, r *RootRouter) *openapi3.T {
	r.Docs.RegisterRoutes(r.rawRouter)

	response := makeRequest(t, r, http.MethodGet, "/docs.json")
	require.Equal(t, http.StatusOK, response.Code)

	doc, err := openapi3.NewLoader().LoadFromData(response.Body.Bytes())
	require.NoError(t, err)

	return doc
}

func TestDocsWithoutSecuritySchema(t *testing.T) {
	handler := func() string {
		return "Hello World!"
	}
	r := Router()
	r.POST("/my/example", handler)

	doc := generateDocs(t, r)

	require.Nil(t, doc.Components)
	require.Nil(t, doc.Security)
}

func TestDocsWithGlobalSecuritySchema(t *testing.T) {
	handler := func() string {
		return "Hello World!"
	}
	securitySchema := &openapi3.SecuritySchemeRef{
		Ref:   "",
		Value: openapi3.NewJWTSecurityScheme(),
	}
	securityRequirements := openapi3.NewSecurityRequirements()
	securityRequirements.With(openapi3.SecurityRequirement{"HTTPBearer": []string{}})
	r := Router(
		&docs.Options{
			Components: &openapi3.Components{SecuritySchemes: openapi3.SecuritySchemes{"HTTPBearer": securitySchema}},
			Security:   *securityRequirements,
		},
	)
	r.POST("/my/example", handler)

	doc := generateDocs(t, r)

	require.NotNil(t, doc.Components)
	require.Equal(t, openapi3.Components{
		Extensions: map[string]interface{}{},
		SecuritySchemes: openapi3.SecuritySchemes{"HTTPBearer": &openapi3.SecuritySchemeRef{Value: &openapi3.SecurityScheme{
			Extensions:   map[string]interface{}{},
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
		}}},
	}, *doc.Components)
	require.NotNil(t, doc.Security)
	require.Equal(t, openapi3.SecurityRequirements{openapi3.SecurityRequirement{"HTTPBearer": []string{}}}, doc.Security)
}

func TestDocsWithEndpointSecuritySchema(t *testing.T) {
	handler := func() string {
		return "Hello World!"
	}
	securitySchema := &openapi3.SecuritySchemeRef{
		Ref:   "",
		Value: openapi3.NewJWTSecurityScheme(),
	}
	securityRequirements := openapi3.NewSecurityRequirements()
	securityRequirements.With(openapi3.SecurityRequirement{"HTTPBearer": []string{}})
	r := Router(
		&docs.Options{
			Components: &openapi3.Components{SecuritySchemes: openapi3.SecuritySchemes{"HTTPBearer": securitySchema}},
		},
	)
	r.POST("/my/example1", handler)
	r.POST("/my/example2", handler, &docs.Endpoint{Security: securityRequirements})

	doc := generateDocs(t, r)

	require.NotNil(t, doc.Components)
	require.Equal(t, openapi3.Components{
		Extensions: map[string]interface{}{},
		SecuritySchemes: openapi3.SecuritySchemes{"HTTPBearer": &openapi3.SecuritySchemeRef{Value: &openapi3.SecurityScheme{
			Extensions:   map[string]interface{}{},
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
		}}},
	}, *doc.Components)
	require.Nil(t, doc.Security)
	require.Nil(t, doc.Paths["/my/example1"].Post.Security)
	require.NotNil(t, doc.Paths["/my/example2"].Post.Security)
	require.Equal(t, openapi3.SecurityRequirements{openapi3.SecurityRequirement{"HTTPBearer": []string{}}}, *doc.Paths["/my/example2"].Post.Security)
}
