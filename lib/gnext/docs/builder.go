package docs

import (
	"encoding/json"
	"github.com/getkin/kin-openapi/openapi3"
	"os"
	"path"
)

type Builder struct {
	docs   *Docs
	Helper *Helper
}

func NewBuilder(docs *Docs) *Builder {
	b := &Builder{docs: docs, Helper: NewHelper(docs)}
	b.init()

	return b
}

func (b *Builder) SetOperationOnPath(path string, method string, operation openapi3.Operation) {
	existingPathItem := b.Helper.GetPathItem(b.Helper.FixPath(path))
	existingPathItem.SetOperation(method, &operation)
	if b.docs.PathsIsEmpty() {
		b.docs.OpenAPI.Paths = make(openapi3.Paths)
	}
	b.docs.OpenAPI.Paths[b.Helper.FixPath(path)] = existingPathItem
}

func (b *Builder) GetPathItem(path string) *openapi3.PathItem {
	return b.Helper.GetPathItem(b.Helper.FixPath(path))
}

func (b *Builder) SetPathItem(path string, pathItem *openapi3.PathItem) {
	b.docs.OpenAPI.Paths[b.Helper.FixPath(path)] = pathItem
}

func (b *Builder) Build() error {
	data, err := json.Marshal(b.docs.OpenAPI)
	if err != nil {
		return err
	}
	return os.WriteFile(path.Join("docs/", "openapi.json"), data, 0664)
}

func (b *Builder) init() {
	if b.docs.OpenAPI == nil {
		b.docs.OpenAPI = &openapi3.T{
			OpenAPI: "3.0.0",
			Info: &openapi3.Info{
				Title:          b.docs.Title,
				Description:    b.docs.Description,
				TermsOfService: b.docs.TermsOfService,
				Contact:        b.docs.Contact,
				License:        b.docs.License,
				Version:        b.docs.Version,
			},
		}
	}
}