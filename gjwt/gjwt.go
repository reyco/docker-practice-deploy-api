package gjwt

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/dgrijalva/jwt-go"
	"github.com/emicklei/go-restful"
)

var (
	Env map[string]interface{}
)

type JwtService struct {
	// This struct is copied from other JWT of Antoine
	// Realm name to display to the user. Required.
	Realm string

	// signing algorithm - possible values are HS256, HS384, HS512
	// Optional, default is HS256.
	SigningAlgorithm string

	// Secret key used for signing. Required.
	Key []byte

	// Duration that a jwt token is valid. Optional, defaults to one hour.
	Timeout time.Duration

	// This field allows clients to refresh their token until MaxRefresh has passed.
	// Note that clients can refresh their token in the last moment of MaxRefresh.
	// This means that the maximum validity timespan for a token is MaxRefresh + Timeout.
	// Optional, defaults to 0 meaning not refreshable.
	MaxRefresh time.Duration

	// Callback function that should perform the authentication of the user based on userId and
	// password. Must return true on success, false on failure. Required.
	Authenticator func(userId string, password string) (bson.M, bool)

	// Callback function that should perform the authorization of the authenticated user. Called
	// only after an authentication success. Must return true on success, false on failure.
	// Optional, default to success.
	Authorizator func(userId string, request *restful.Request) bool

	// Callback function that will be called during login.
	// Using this function it is possible to add additional payload data to the webtoken.
	// The data is then made available during requests via request.Env["JWT_PAYLOAD"].
	// Note that the payload is not encrypted.
	// The attributes mentioned on jwt.io can't be used as keys for the map.
	// Optional, by default no additional data will be set.
	PayloadFunc func(userId string) map[string]interface{}
}

type AuthUser struct {
	username, password string
}

// Handler that clients can use to get a jwt token.
// Payload needs to be json in the form of {"username": "USERNAME", "password": "PASSWORD"}.
// Reply will be of the form {"token": "TOKEN"}.
func (jwts *JwtService) LoginHandler(request *restful.Request, response *restful.Response) {
	fmt.Println("Logging In")

	data := bson.M{}
	if err := request.ReadEntity(&data); err != nil {
		fmt.Println("Empty Username or Password")
		fmt.Println(err)

		jwts.unauthorized(response)
		return
	}

	fmt.Println("From user inputs:", data)
	usr, ok := jwts.Authenticator(data["username"].(string), data["password"].(string))
	if !ok {
		fmt.Println("Wrong Username or Password")
		jwts.unauthorized(response)
		return
	}

	token := jwt.New(jwt.GetSigningMethod(jwts.SigningAlgorithm))

	if jwts.PayloadFunc != nil {
		for key, value := range jwts.PayloadFunc(data["username"].(string)) {
			token.Claims[key] = value
		}
	}

	fmt.Printf("The user from logindb %v", usr)
	token.Claims["id"] = usr["_id"].(bson.ObjectId)
	token.Claims["username"] = usr["username"].(string)
	token.Claims["exp"] = time.Now().Add(jwts.Timeout).Unix()
	if jwts.MaxRefresh != 0 {
		token.Claims["orig_iat"] = time.Now().Unix()
	}
	tokenString, err := token.SignedString(jwts.Key)

	if err != nil {
		response.WriteErrorString(http.StatusUnauthorized, "Unauthorized Access")
		jwts.unauthorized(response)
		return
	}

	response.WriteEntity(&bson.M{"meta": bson.M{"token": tokenString}})
}

// Handler that clients can use to get a jwt token.
// Payload needs to be json in the form of {"username": "USERNAME", "password": "PASSWORD"}.
// Reply will be of the form {"token": "TOKEN"}.
func (jwts *JwtService) SignupToken(request *restful.Request, response *restful.Response, usr bson.M) string {
	fmt.Println("Getting signup token")

	data := &AuthUser{}
	if err := request.ReadEntity(data); err != nil {
		fmt.Println("Empty Username or Password")
		fmt.Println(err)

		jwts.unauthorized(response)
		return ""
	}

	token := jwt.New(jwt.GetSigningMethod(jwts.SigningAlgorithm))

	if jwts.PayloadFunc != nil {
		for key, value := range jwts.PayloadFunc(data.username) {
			token.Claims[key] = value
		}
	}

	token.Claims["id"] = usr["_id"].(bson.ObjectId)
	token.Claims["username"] = usr["username"].(string)
	token.Claims["exp"] = time.Now().Add(jwts.Timeout).Unix()
	if jwts.MaxRefresh != 0 {
		token.Claims["orig_iat"] = time.Now().Unix()
	}
	tokenString, err := token.SignedString(jwts.Key)

	if err != nil {
		response.WriteErrorString(http.StatusUnauthorized, "Unauthorized Access")
		jwts.unauthorized(response)
		return ""
	}

	return tokenString
}

func (jwts *JwtService) IsValidToken(request *restful.Request) bool {

	token, err := jwts.parseToken(request)

	// Token should be valid anyway as the RefreshHandler is authed
	if err != nil {
		return false
	}

	origIat := int64(token.Claims["orig_iat"].(float64))

	if origIat < time.Now().Add(-jwts.MaxRefresh).Unix() {
		return false
	}
	return true
}

func (jwts *JwtService) RefreshHandler(request *restful.Request, response *restful.Response) {
	token, err := jwts.parseToken(request)

	// Token should be valid anyway as the RefreshHandler is authed
	if err != nil {
		jwts.unauthorized(response)
		return
	}

	origIat := int64(token.Claims["orig_iat"].(float64))

	if origIat < time.Now().Add(-jwts.MaxRefresh).Unix() {
		jwts.unauthorized(response)
		return
	}

	newToken := jwt.New(jwt.GetSigningMethod(jwts.SigningAlgorithm))

	for key := range token.Claims {
		newToken.Claims[key] = token.Claims[key]
	}

	newToken.Claims["id"] = token.Claims["id"]
	newToken.Claims["exp"] = time.Now().Add(jwts.Timeout).Unix()
	newToken.Claims["orig_iat"] = origIat
	tokenString, err := newToken.SignedString(jwts.Key)

	if err != nil {
		jwts.unauthorized(response)
		return
	}

	response.WriteEntity(&map[string]string{"token": tokenString})
}

func (jwts *JwtService) parseToken(request *restful.Request) (*jwt.Token, error) {
	authHeader := request.HeaderParameter("Authorization")

	if authHeader == "" {
		return nil, errors.New("Auth header empty")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		return nil, errors.New("Invalid auth header")
	}

	return jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
		if jwt.GetSigningMethod(jwts.SigningAlgorithm) != token.Method {
			return nil, errors.New("Invalid signing algorithm")
		}
		return jwts.Key, nil
	})
}

func (jwts *JwtService) unauthorized(response *restful.Response) {
	response.AddHeader("WWW-Authenticate", "JWT realm="+jwts.Realm)
	response.WriteErrorString(http.StatusUnauthorized, "Not Authorized")
}

type HandlerFunc func(request *restful.Request, response *restful.Response)

// Init
func (jwts *JwtService) Init() {

	if jwts.Realm == "" {
		log.Fatal("Realm is required")
	}
	if jwts.SigningAlgorithm == "" {
		jwts.SigningAlgorithm = "HS256"
	}
	if jwts.Key == nil {
		log.Fatal("Key required")
	}
	if jwts.Timeout == 0 {
		jwts.Timeout = time.Hour
	}
	if jwts.Authenticator == nil {
		log.Fatal("Authenticator is required")
	}
	if jwts.Authorizator == nil {
		jwts.Authorizator = func(userId string, request *restful.Request) bool {
			return true
		}
	}
}

func (jwts *JwtService) Guard(request *restful.Request, response *restful.Response) (bson.M, error) {
	token, err := jwts.parseToken(request)

	if err != nil {
		jwts.unauthorized(response)
		return nil, err
	}

	origIat := int64(token.Claims["orig_iat"].(float64))

	if origIat < time.Now().Add(-jwts.MaxRefresh).Unix() {
		return nil, err
	}

	return bson.M{"id": token.Claims["id"], "username": token.Claims["username"]}, nil
}

// MiddlewareFunc makes JWTMiddleware implement the Middleware interface.
func (jwts *JwtService) MiddlewareFunc(handler HandlerFunc) HandlerFunc {

	if jwts.Realm == "" {
		log.Fatal("Realm is required")
	}
	if jwts.SigningAlgorithm == "" {
		jwts.SigningAlgorithm = "HS256"
	}
	if jwts.Key == nil {
		log.Fatal("Key required")
	}
	if jwts.Timeout == 0 {
		jwts.Timeout = time.Hour
	}
	if jwts.Authenticator == nil {
		log.Fatal("Authenticator is required")
	}
	if jwts.Authorizator == nil {
		jwts.Authorizator = func(userId string, request *restful.Request) bool {
			return true
		}
	}

	return func(request *restful.Request, response *restful.Response) {
		jwts.middlewareImpl(request, response, handler)
	}
}

func (jwts *JwtService) middlewareImpl(request *restful.Request, response *restful.Response, handler HandlerFunc) {
	token, err := jwts.parseToken(request)

	if err != nil {
		jwts.unauthorized(response)
		return
	}

	id := token.Claims["id"].(string)

	Env["REMOTE_USER"] = id
	Env["JWT_PAYLOAD"] = token.Claims

	if !jwts.Authorizator(id, request) {
		jwts.unauthorized(response)
		return
	}

	handler(request, response)
}

// Helper function to extract the JWT claims
func ExtractClaims(request *rest.Request) map[string]interface{} {
	if request.Env["JWT_PAYLOAD"] == nil {
		empty_claims := make(map[string]interface{})
		return empty_claims
	}
	jwt_claims := request.Env["JWT_PAYLOAD"].(map[string]interface{})
	return jwt_claims

}
