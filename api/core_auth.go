package api

import (
	"fmt"

	"net/http"

	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/emicklei/go-restful"

	"crypto/sha1"
	"encoding/hex"

	"../database"
	"../gjwt"
	"../models"
)

// PUT http://localhost:8080/{noun_url}/1
// <User><Id>1</Id><Name>Melissa</Name></User>
//
func (as *ApiService) signup(request *restful.Request, response *restful.Response) {

	data := bson.M{"_id": bson.NewObjectId()}
	if err := request.ReadEntity(&data); err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	if _, ok := data["password"]; !ok {
		response.WriteErrorString(http.StatusInternalServerError, "Empty Password")
		return
	}
	if _, ok := data["username"]; !ok {
		response.WriteErrorString(http.StatusInternalServerError, "Empty Username")
		return
	}

	if models.IsExists(as.collection, &bson.M{"username": data["username"]}) {
		response.WriteErrorString(http.StatusInternalServerError, "Empty Password")
		return
	}

	data["password"] = GenPasswordHash(data["password"].(string))

	fmt.Println(data["password"])

	if err := models.Create(as.collection, &data); err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	tokenString := gJwtService.SignupToken(request, response, data)

	response.WriteEntity(bson.M{"data": data, "meta": bson.M{"token": tokenString}})

	// response.WriteEntity(bson.M{"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0MzQ0NDQ0OTQsImlkIjoiIiwib3JpZ19pYXQiOjE0MzQ0NDA4OTR9.4mQvwOmA2sf5pbHiIXAD-1003MDsQ6WYdo_J1nzgeaY"})
}

type AuthUser struct {
	username, password string
}

type SignupStruct struct {
	username, password string
}

func (as *ApiService) Authenticator(userId string, password string) (bson.M, bool) {
	fmt.Println(userId, password)
	usr, err := models.FindOne(as.collection, &bson.M{"username": userId})
	if err != nil {
		return nil, false
	}
	fmt.Printf("From Database %v", usr)

	return usr, CheckPassword(password, usr["password"].(string))

}

var (
	gJwtService *gjwt.JwtService
)

func NewAuthService() *ApiService {
	as := new(ApiService)
	as.collection = database.GMyDb.GetCollection("users")

	gJwtService = &gjwt.JwtService{
		SigningAlgorithm: "HS256",
		Key:              []byte("secret key"),
		Realm:            "jwt auth",
		Timeout:          time.Hour,
		MaxRefresh:       time.Hour * 24,
		Authenticator:    as.Authenticator}

	gJwtService.Init()

	ws := new(restful.WebService)

	ws.
		Path("/api/auth").
		Consumes(restful.MIME_JSON, restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_JSON) // you can specify this per route as well

	ws.Route(ws.POST("/login").To(gJwtService.LoginHandler).
		// docs
		Doc("login to create token").
		Operation("createTokenViaLogin").
		Reads(AuthUser{})) // from the request

	ws.Route(ws.GET("/test").To(as.AuthTest).
		// docs
		Doc("test the token").
		Operation("testToken"))

	ws.Route(ws.POST("/signup").To(as.signup).
		// docs
		Doc("for signing up").
		Operation("signup").
		Reads(SignupStruct{})) // from the request

	ws.Route(ws.GET("/refresh").To(gJwtService.RefreshHandler).
		// docs
		Doc("refresh the token").
		Operation("refreshToken"))

	restful.Add(ws)
	return as

}

func (as *ApiService) AuthTest(request *restful.Request, response *restful.Response) {
	tokenUsr, err := gJwtService.Guard(request, response)
	if err != nil {
		return
	}
	response.WriteEntity(&bson.M{"data": bson.M{"_id": tokenUsr["id"], "username": tokenUsr["username"]}})
}

func (as *ApiService) AuthInfo(request *restful.Request, response *restful.Response) bson.M {
	fmt.Println("Ni sud in AuthInfo")
	tokenUsr, err := gJwtService.Guard(request, response)

	fmt.Println("Lampus in AuthInfo")
	if err != nil {
		return nil
	}
	fmt.Printf("In AuthInfo Fx : %v", tokenUsr)
	return bson.M{"_id": tokenUsr["id"], "username": tokenUsr["username"]}
}

func (as *ApiService) IsAdmin(authInfo bson.M) bool {
	return authInfo["username"] == "admin"
}

func HashPassword(password string) string {
	h := sha1.New()
	h.Write([]byte("fZJ9MYnzeaW7q3DY" + password))
	return hex.EncodeToString(h.Sum(nil))
}

func GenNonce() string {
	return HashPassword(time.Now().String())
}

func CheckPassword(passwordInputted string, passwordCorrectHash string) bool {

	nonce := passwordCorrectHash[0:40]
	the_hash := passwordCorrectHash[40:]

	return HashPassword(passwordInputted+nonce) == the_hash
}

func GenPasswordHash(passwordInputted string) string {
	nonce := GenNonce()
	return nonce + HashPassword(passwordInputted+nonce)
}
