package db
import (
  "reflect"
  "strings"
)

type EmailResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

type Accounts struct {
  AccountData []Account `bson:"accounts"`
}

type Account struct {
  Email string `bson:"email"`
  Database string `bson:"database"`
  Size string
}

type Table struct {
  Name string `bson:"name"`
  Requires []string `bson:"requires"`
  Identifier string `bson:"identifier"`
  EntryTemplate map[string]string `bson:"entrytemplate"`
  Entries map[string]map[string]interface{}`bson:"entries"`
}

var SpecialCharactersToLetters map[string]string
func init() {
  SpecialCharactersToLetters = map[string]string{
    "\\": "backslash",
    "/": "slash",
    ":": "colon",
    "*": "asterisk",
    "?": "questionmark",
    "\"": "quote",
    "<": "lessthan",
    ">": "greaterthan",
    "|": "pipe",
    ".": "period",
  }
}

func ValidateIdentifier(identifier string) string {
  for k, v := range SpecialCharactersToLetters {
    identifier = strings.ReplaceAll(identifier, k, v)
  }
  return identifier
}

type TableDefinition struct {
  Name string
  Identifier string
  Requires []string
  EntryTemplate map[string]string
}

type DBAccount struct {
  Email string
  Password string
  DatabaseId string
}

type DBAccounts struct {
  Accounts []DBAccount
}

func DetermindType(i interface{}) string {
	var typestr string
	switch reflect.TypeOf(i).Kind() {
	case reflect.String:
		typestr = "string"
	case reflect.Int, reflect.Float64:
		typestr = "number"
	case reflect.Bool:
		typestr = "boolean"
	default:
		typestr = "object"
	}
	return typestr
}
