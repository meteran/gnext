package gnext

import (
	"github.com/meteran/gnext/docs"
	"github.com/stretchr/testify/assert"
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
