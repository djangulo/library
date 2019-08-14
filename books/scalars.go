package books

import (
	"database/sql"
	"encoding/json"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/pkg/errors"
	"strconv"
)

// NullString struct to represent sql.NullString in graphql
// queries and mutations
type NullString struct {
	sql.NullString
}

// MarshalJSON from the json.Marshaler interface
func (v NullString) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.String)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON from the json.Unmarshaler interface
func (v *NullString) UnmarshalJSON(data []byte) error {
	var x *string
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.String = *x
		v.Valid = true
	} else {
		v.Valid = false
	}
	return nil
}

// NewNullString create a new null string
func NewNullString(value string) NullString {
	var null NullString
	if value != "" {
		null.String = value
		null.Valid = true
		return null
	}
	null.Valid = false
	return null
}

// SerializeNullString serializes `NullString` to a string
func SerializeNullString(value interface{}) interface{} {
	switch value := value.(type) {
	case NullString:
		return value.String
	case *NullString:
		v := *value
		return v.String
	default:
		return nil
	}
}

// ParseNullString parses GraphQL variables from `string` to `CustomID`
func ParseNullString(value interface{}) interface{} {
	switch value := value.(type) {
	case string:
		return NewNullString(value)
	case *string:
		return NewNullString(*value)
	default:
		return nil
	}
}

// ParseLiteralNullString parses GraphQL AST value to `NullString`.
func ParseLiteralNullString(valueAST ast.Value) interface{} {
	switch valueAST := valueAST.(type) {
	case *ast.StringValue:
		return NewNullString(valueAST.Value)
	default:
		return nil
	}
}

// NullableString graphql *Scalar type based of NullString
var NullableString = graphql.NewScalar(graphql.ScalarConfig{
	Name:         "NullableString",
	Description:  "The `NullableString` type repesents a nullable SQL string.",
	Serialize:    SerializeNullString,
	ParseValue:   ParseNullString,
	ParseLiteral: ParseLiteralNullString,
})

// NullInt64 struct to represent sql.NullInt64 in graphql
// queries and mutations
type NullInt64 struct {
	sql.NullInt64
}

// MarshalJSON from the json.Marshaler interface
func (v NullInt64) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Int64)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON from the json.Unmarshaler interface
func (v *NullInt64) UnmarshalJSON(data []byte) error {
	var x *int64
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Int64 = *x
		v.Valid = true
	} else {
		v.Valid = false
	}
	return nil
}

// NewNullInt64 create a new null string
func NewNullInt64(value int64) NullInt64 {
	var null NullInt64
	if value == 0 {
		null.Valid = false
	} else {
		null.Valid = true
		null.Int64 = value
	}
	return null
}

// SerializeNullInt64 serializes `NullInt64` to a string
func SerializeNullInt64(value interface{}) interface{} {
	switch value := value.(type) {
	case NullInt64:
		return value.Int64
	case *NullInt64:
		v := *value
		return v.Int64
	case sql.NullInt64:
		return value.Int64
	case *sql.NullInt64:
		v := *value
		return v.Int64
	default:
		return nil
	}
}

// ParseNullInt64 parses GraphQL variables from `string` to `CustomID`
func ParseNullInt64(value interface{}) interface{} {
	switch value := value.(type) {
	case int64:
		return NewNullInt64(value)
	case *int64:
		return NewNullInt64(*value)
	default:
		return nil
	}
}

// ParseLiteralNullInt64 parses GraphQL AST value to `NullInt64`.
func ParseLiteralNullInt64(valueAST ast.Value) interface{} {
	switch valueAST := valueAST.(type) {
	case *ast.IntValue:
		inted, err := strconv.Atoi(valueAST.Value)
		if err != nil {
			return errors.Wrap(err, "could not parse AST value")
		}
		return NewNullInt64(int64(inted))
	default:
		return nil
	}
}

// NullableInt64 graphql *Scalar type based of NullInt64
var NullableInt64 = graphql.NewScalar(graphql.ScalarConfig{
	Name:         "NullableString",
	Description:  "The `NullableString` type repesents a nullable SQL string.",
	Serialize:    SerializeNullInt64,
	ParseValue:   ParseNullInt64,
	ParseLiteral: ParseLiteralNullInt64,
})