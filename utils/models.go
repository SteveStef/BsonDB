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

type Table struct {
  Name string `bson:"name"`
  Requires []string `bson:"requires"`
  Identifier string `bson:"identifier"`
  EntryTemplate map[string]interface{} `bson:"entryTemplate"`
  Entries map[string]map[string]interface{}`bson:"entries"`
}

type Model struct {
  Tables []Table `bson:"tables"`
}

type AdminData struct {
  UserAccounts []Account
  Size string
}
