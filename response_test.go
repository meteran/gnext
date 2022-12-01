package gnext

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCustomTypesInResponse(t *testing.T) {
	type CustomBool bool
	type CustomInt int
	type CustomInt8 int8
	type CustomInt16 int16
	type CustomInt32 int32
	type CustomInt64 int64
	type CustomUint uint
	type CustomUint8 uint8
	type CustomUint16 uint16
	type CustomUint32 uint32
	type CustomUint64 uint64
	type CustomFloat32 float32
	type CustomFloat64 float64
	type CustomArray [3]string
	type CustomString string
	type CustomInterface interface{}
	type CustomMap map[string]interface{}
	type CustomSlice []string
	type CustomStruct struct {
		Collection   []int   `json:"collection"`
		CustomString *string `json:"custom_string"`
	}

	type payload struct {
		Bool      CustomBool      `json:"bool"`
		Int       CustomInt       `json:"int"`
		Int8      CustomInt8      `json:"int8"`
		Int16     CustomInt16     `json:"int16"`
		Int32     CustomInt32     `json:"int32"`
		Int64     CustomInt64     `json:"int64"`
		Uint      CustomUint      `json:"uint"`
		Uint8     CustomUint8     `json:"uint8"`
		Uint16    CustomUint16    `json:"uint16"`
		Uint32    CustomUint32    `json:"uint32"`
		Uint64    CustomUint64    `json:"uint64"`
		Float32   CustomFloat32   `json:"float32"`
		Float64   CustomFloat64   `json:"float64"`
		Array     CustomArray     `json:"array"`
		String    CustomString    `json:"string"`
		Interface CustomInterface `json:"interface"`
		Map       CustomMap       `json:"map"`
		Slice     CustomSlice     `json:"slice"`
		Struct    CustomStruct    `json:"struct"`

		// the same fields provided by pointers
		BoolPtr      *CustomBool      `json:"bool_ptr"`
		IntPtr       *CustomInt       `json:"int_ptr"`
		Int8Ptr      *CustomInt8      `json:"int8_ptr"`
		Int16Ptr     *CustomInt16     `json:"int16_ptr"`
		Int32Ptr     *CustomInt32     `json:"int32_ptr"`
		Int64Ptr     *CustomInt64     `json:"int64_ptr"`
		UintPtr      *CustomUint      `json:"uint_ptr"`
		Uint8Ptr     *CustomUint8     `json:"uint8_ptr"`
		Uint16Ptr    *CustomUint16    `json:"uint16_ptr"`
		Uint32Ptr    *CustomUint32    `json:"uint32_ptr"`
		Uint64Ptr    *CustomUint64    `json:"uint64_ptr"`
		Float32Ptr   *CustomFloat32   `json:"float32_ptr"`
		Float64Ptr   *CustomFloat64   `json:"float64_ptr"`
		ArrayPtr     *CustomArray     `json:"array_ptr"`
		StringPtr    *CustomString    `json:"string_ptr"`
		InterfacePtr *CustomInterface `json:"interface_ptr"`
		MapPtr       *CustomMap       `json:"map_ptr"`
		SlicePtr     *CustomSlice     `json:"slice_ptr"`
		StructPtr    *CustomStruct    `json:"struct_ptr"`
	}

	handler := func() *payload {
		return nil
	}

	r := Router()
	r.GET("/full", handler)

	docs, err := json.Marshal(r.Docs.OpenApi)
	require.NoError(t, err)
	assert.JSONEq(t, fullCustomKindRequestResponse, string(docs))
}
