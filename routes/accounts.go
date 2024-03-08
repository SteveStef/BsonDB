package route
import (
  "github.com/gin-gonic/gin"
  "net/http"
  "BsonDB-API/utils"
)

func FetchLoggedInStatus(c *gin.Context) {
  token := c.GetHeader("Authorization")
  if token == "" {
    c.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
    return
  }
  email, dbId, err := db.FetchLoggedInStatus(token)
  if err != nil { 
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }
  account := gin.H{"email": email, "id": dbId}
  c.JSON(http.StatusOK, account)
}

func Login(c *gin.Context) {
  var req map[string]string
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }
  if _, ok := req["email"]; !ok {
    c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
    return
  }
  if _, ok := req["password"]; !ok {
    c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
    return
  }
  dbId, token, err := db.Login(req["email"], req["password"])
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }
  c.JSON(http.StatusOK, gin.H{"id": dbId, "token": token})
}


func VerifyAccount(c *gin.Context) {
  var req map[string]string
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }
  if _, ok := req["email"]; !ok {
    c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
    return
  }
  if _, ok := req["code"]; !ok {
    c.JSON(http.StatusBadRequest, gin.H{"error": "Code is required"})
    return
  }
  err := db.VerifyAccount(req["email"], req["code"])
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }
  c.JSON(http.StatusOK, gin.H{"message": "Account verified"})
}

func SendVerificationCode(c *gin.Context) {
  var req map[string]string
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }
  if _, ok := req["email"]; !ok {
    c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
    return
  }

  emailRes := db.SendVerificationCode(req["email"])
  c.JSON(http.StatusOK, gin.H{"message": emailRes.Message})
}


func Signup(c *gin.Context) {
  var req map[string]string
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }

  if _, ok := req["email"]; !ok {
    c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
    return
  }

  if _, ok := req["password"]; !ok {
    c.JSON(http.StatusBadRequest, gin.H{"error": "password is required"})
    return
  }

  dbId, token, err := db.Signup(req["email"], req["password"])
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }
  c.JSON(http.StatusOK, gin.H{"id": dbId, "token": token})
}

