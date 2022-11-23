package common

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

const (
	DefaultTimeout = 60

	DefaultIntFieldName = "int64"

	DefaultFloatFieldName = "float"

	DefaultVarcharFieldName = "varchar"

	DefaultFloatVecFieldName = "float_vec"

	DefaultBinaryVecFieldName = "binary_vec"

	DefaultPartition = "_default"

	DefaultDim    = 128
	DefaultDimStr = "128"

	MaxDim = 32768

	DefaultMaxLength = "65535"

	DefaultShards = int32(2)

	DefaultConsistencyLevel = entity.ConsistencyLevel(0)

	DefaultNb = 3000

	DefaultNq = 5

	DefaultTopK = 10

	MaxCollectionNameLen = 255

	RowCount = "row_count"
)

var r *rand.Rand

func init() {
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// gen random string
func GenRandomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[r.Intn(len(letterRunes))]
	}
	return string(b)
}

// gen scala field
func GenScalaField(name string, fieldType entity.FieldType, primaryKey bool, autoID bool) *entity.Field {
	var scaleField = new(entity.Field)
	scaleField.Name = name
	scaleField.PrimaryKey = primaryKey
	scaleField.AutoID = autoID
	scaleField.DataType = fieldType
	if fieldType == entity.FieldTypeVarChar {
		scaleField.TypeParams = map[string]string{entity.TypeParamMaxLength: DefaultMaxLength}
	}
	return scaleField
}

// gen vector field
func GenVectorField(name string, fieldType entity.FieldType, dim string) *entity.Field {
	if fieldType == entity.FieldTypeFloatVector || fieldType == entity.FieldTypeBinaryVector {
		var vecField = new(entity.Field)
		vecField.Name = name
		vecField.DataType = fieldType
		vecField.PrimaryKey = false
		vecField.TypeParams = map[string]string{entity.TypeParamDim: dim}
		return vecField
	}
	return nil
}

// gen default fields with int64, float, floatVector field
func GenDefaultFields(autoID bool) []*entity.Field {
	var fields = []*entity.Field{
		{
			Name:       DefaultIntFieldName,
			DataType:   entity.FieldTypeInt64,
			PrimaryKey: true,
			AutoID:     autoID,
		},
		{
			Name:     DefaultFloatFieldName,
			DataType: entity.FieldTypeFloat,
		},
		{
			Name:       DefaultFloatVecFieldName,
			DataType:   entity.FieldTypeFloatVector,
			TypeParams: map[string]string{entity.TypeParamDim: fmt.Sprintf("%d", DefaultDim)},
		},
	}
	return fields
}

// gen default binary fields with int64, float, binaryVector field
func GenDefaultBinaryFields(autoID bool, dim string) []*entity.Field {
	var fields = []*entity.Field{
		{
			Name:       DefaultIntFieldName,
			DataType:   entity.FieldTypeInt64,
			PrimaryKey: true,
			AutoID:     autoID,
		},
		{
			Name:     DefaultFloatFieldName,
			DataType: entity.FieldTypeFloat,
		},
		{
			Name:       DefaultBinaryVecFieldName,
			DataType:   entity.FieldTypeBinaryVector,
			TypeParams: map[string]string{entity.TypeParamDim: dim},
		},
	}
	return fields
}

//gen default fields with varchar, floatVector field
func GenDefaultVarcharFields(autoID bool) []*entity.Field {
	var fields = []*entity.Field{
		{
			Name:       DefaultVarcharFieldName,
			DataType:   entity.FieldTypeVarChar,
			PrimaryKey: true,
			AutoID:     autoID,
			TypeParams: map[string]string{entity.TypeParamMaxLength: DefaultMaxLength},
		},
		{
			Name:       DefaultBinaryVecFieldName,
			DataType:   entity.FieldTypeBinaryVector,
			TypeParams: map[string]string{entity.TypeParamDim: DefaultDimStr},
		},
	}
	return fields
}

// gen schema
func GenSchema(name string, autoID bool, fields []*entity.Field) *entity.Schema {
	schema := &entity.Schema{
		CollectionName: name,
		AutoID:         autoID,
		Fields:         fields,
	}
	return schema
}

// gen float vector values
func GenFloatVector(nb, dim int) [][]float32 {
	rand.Seed(time.Now().Unix())
	floatVectors := make([][]float32, 0, nb)
	for i := 0; i < nb; i++ {
		vec := make([]float32, 0, dim)
		for j := 0; j < dim; j++ {
			vec = append(vec, rand.Float32())
		}
		floatVectors = append(floatVectors, vec)
	}
	return floatVectors
}

// gen binary vector values
func GenBinaryVector(nb int, dim int) [][]byte {
	binaryVectors := make([][]byte, 0, nb)
	for i := 0; i < nb; i++ {
		vec := make([]byte, dim/8)
		rand.Read(vec)
		binaryVectors = append(binaryVectors, vec)
	}
	return binaryVectors
}

// gen default column with data
func GenDefaultColumnData(nb int, dim int) (*entity.ColumnInt64, *entity.ColumnFloat, *entity.ColumnFloatVector) {
	int64Values := make([]int64, 0, nb)
	floatValues := make([]float32, 0, nb)
	vecFloatValues := make([][]float32, 0, nb)
	for i := 0; i < nb; i++ {
		int64Values = append(int64Values, int64(i))
		floatValues = append(floatValues, float32(i))
		vec := make([]float32, 0, DefaultDim)
		for j := 0; j < dim; j++ {
			vec = append(vec, rand.Float32())
		}
		vecFloatValues = append(vecFloatValues, vec)
	}
	intColumn := entity.NewColumnInt64(DefaultIntFieldName, int64Values)
	floatColumn := entity.NewColumnFloat(DefaultFloatFieldName, floatValues)
	vecColumn := entity.NewColumnFloatVector(DefaultFloatVecFieldName, DefaultDim, vecFloatValues)
	return intColumn, floatColumn, vecColumn
}

// gen default binary collection data
func GenDefaultBinaryData(nb int, dim int) (*entity.ColumnInt64, *entity.ColumnFloat, *entity.ColumnBinaryVector) {
	int64Values := make([]int64, 0, nb)
	floatValues := make([]float32, 0, nb)
	vecBinaryValues := make([][]byte, 0, nb)
	for i := 0; i < nb; i++ {
		int64Values = append(int64Values, int64(i))
		floatValues = append(floatValues, float32(i))
		vec := make([]byte, dim/8)
		rand.Read(vec)
		vecBinaryValues = append(vecBinaryValues, vec)
	}
	intColumn := entity.NewColumnInt64(DefaultIntFieldName, int64Values)
	floatColumn := entity.NewColumnFloat(DefaultFloatFieldName, floatValues)
	vecColumn := entity.NewColumnBinaryVector(DefaultBinaryVecFieldName, dim, vecBinaryValues)
	return intColumn, floatColumn, vecColumn
}

// gen search vectors
func GenSearchVectors(nq int, dim int) []entity.Vector {
	vectors := make([]entity.Vector, 0, nq)
	for i := 0; i < nq; i++ {
		vector := make([]float32, 0, dim)
		for j := 0; j < dim; j++ {
			vector = append(vector, rand.Float32())
		}
		vectors = append(vectors, entity.FloatVector(vector))
	}
	return vectors
}

// gen invalid long string
func GenLongString(n int) string {
	var builder strings.Builder
	longString := "a"
	for i := 0; i < n; i++ {
		builder.WriteString(longString)
	}
	return builder.String()
}

// gen fields with all scala field types
func GenAllFields() []*entity.Field {
	allFields := []*entity.Field{
		GenScalaField("int64", entity.FieldTypeInt64, true, false),                // int64
		GenScalaField("bool", entity.FieldTypeBool, false, false),                 // bool
		GenScalaField("int8", entity.FieldTypeInt8, false, false),                 // int8
		GenScalaField("int16", entity.FieldTypeInt16, false, false),               // int16
		GenScalaField("int32", entity.FieldTypeInt32, false, false),               // int32
		GenScalaField("float", entity.FieldTypeFloat, false, false),               // float
		GenScalaField("double", entity.FieldTypeDouble, false, false),             // double
		GenScalaField("varChar", entity.FieldTypeVarChar, false, false),           // varchar
		GenVectorField("floatVector", entity.FieldTypeFloatVector, DefaultDimStr), // float vector
	}
	return allFields
}
