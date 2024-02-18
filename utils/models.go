package db

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

type table struct {
  name string `bson:"name"`
  requires []string `bson:"requires"`
  identifier string `bson:"identifier"`
  entrytemplate map[string]string `bson:"entrytemplate"`
  entries map[string]map[string]interface{}`bson:"entries"`
}

type model struct {
  tables []table `bson:"tables"`
}

type AdminData struct {
  UserAccounts []Account
  Size string
}

func DetermindType(i interface{}) string {
	var typestr string
	switch reflect.typeof(i).kind() {
	case reflect.string:
		typestr = "string"
	case reflect.int, reflect.float64:
		typestr = "number"
	case reflect.bool:
		typestr = "boolean"
	default:
		typestr = "object"
	}
	return typestr
}
