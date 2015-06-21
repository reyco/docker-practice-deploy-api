package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/swagger"

	"../database"
	"../gjwt"
	"../models"
)

// This example is functionally the same as the example in restful-user-resource.go
// with the only difference that is served using the restful.DefaultContainer

type ApiService struct {
	collection *mgo.Collection
	path       string
	jwtService *gjwt.JwtService
}

func NewApiService(ModelSettings *models.ModelSettings) *ApiService {

	as := new(ApiService)
	as.collection = database.GMyDb.GetCollection(ModelSettings.CollectionName)
	as.path = ModelSettings.Path

	ws := new(restful.WebService)
	ws.
		Path("/api"+as.path).
		Consumes(restful.MIME_JSON, restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_JSON) // you can specify this per route as well

	ws.Route(ws.GET("/").To(as.findAll).
		// docs
		Doc("get all "+ModelSettings.Noun).
		Operation("findAll"+ModelSettings.Noun+"s").
		Returns(200, "OK", nil))

	ws.Route(ws.GET("/{id}").To(as.find).
		// docs
		Doc("get a " + ModelSettings.Noun).
		Operation("find" + ModelSettings.Noun).
		Param(ws.PathParameter("id", "identifier of the "+ModelSettings.Noun).DataType("string")).
		Writes(ModelSettings.DataStruct)) // on the response

	ws.Route(ws.PUT("/{id}").To(as.update).
		// docs
		Doc("update a " + ModelSettings.Noun).
		Operation("update" + ModelSettings.Noun).
		Param(ws.PathParameter("id", "identifier of the "+ModelSettings.Noun).DataType("string")).
		Reads(ModelSettings.DataStruct)) // from the request

	ws.Route(ws.POST("").To(as.create).
		// docs
		Doc("create a " + ModelSettings.Noun).
		Operation("create" + ModelSettings.Noun).
		Reads(ModelSettings.DataStruct)) // from the request

	ws.Route(ws.DELETE("/{id}").To(as.remove).
		// docs
		Doc("delete a " + ModelSettings.Noun).
		Operation("remove" + ModelSettings.Noun).
		Param(ws.PathParameter("id", "identifier of the "+ModelSettings.Noun).DataType("string")))

	restful.Add(ws)

	return as

}

// GET http://localhost:8080/{noun_url}
//
func (as *ApiService) findAll(request *restful.Request, response *restful.Response) {

	pageOffset, err := strconv.Atoi(request.QueryParameter("page[offset]"))
	if err != nil {
		pageOffset = 0
	}

	pageLimit, err := strconv.Atoi(request.QueryParameter("page[limit]"))
	if err != nil {
		pageLimit = 10
	}

	query := bson.M{}

	authInfo := as.AuthInfo(request, response)
	fmt.Println("ang kita all")
	fmt.Printf("authInfo is %v", authInfo)
	if !as.IsAdmin(authInfo) {
		query = bson.M{"_id": bson.ObjectIdHex(authInfo["_id"].(string))}
	}
	data, err := models.FindAll(as.path, as.collection, query, pageOffset, pageLimit)
	if err != nil {
		response.WriteErrorString(http.StatusNotFound, "Empty data")
		return
	}

	// cors(response)
	response.WriteEntity(data)
}

// GET http://localhost:8080/{noun_url}/1
//
func (as *ApiService) find(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("id")
	if len(id) == 0 {
		response.WriteErrorString(http.StatusBadRequest, "Invalid Request")
		return
	}
	authInfo := as.AuthInfo(request, response)
	if !as.IsAdmin(authInfo) {
		if authInfo["_id"].(string) != id {
			response.WriteErrorString(http.StatusBadRequest, "Invalid Request")
			return
		}
	}
	data, err := models.FindId(as.collection, id)
	if err != nil {
		response.WriteErrorString(http.StatusNotFound, "User could not be found.")
		return
	}

	// cors(response)
	response.WriteEntity(bson.M{"data": data})
}

// PUT http://localhost:8080/{noun_url}/1
// <User><Id>1</Id><Name>Melissa Raspberry</Name></User>
//
func (as *ApiService) update(request *restful.Request, response *restful.Response) {

	id := request.PathParameter("id")

	authInfo := as.AuthInfo(request, response)
	if !as.IsAdmin(authInfo) {
		if authInfo["_id"].(string) != id {
			response.WriteErrorString(http.StatusBadRequest, "Invalid Request")
			return
		}
	}

	data := bson.M{}
	if err := request.ReadEntity(&data); err != nil {
		fmt.Println("blank data")
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	if _, ok := data["password"]; ok {
		data["password"] = GenPasswordHash(data["password"].(string))
	}

	err := models.Update(as.collection, id, &data)
	if err != nil {
		fmt.Println("can't update")
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	data["_id"] = id

	// cors(response)
	response.WriteEntity(bson.M{"data": data})
}

// PUT http://localhost:8080/{noun_url}/1
// <User><Id>1</Id><Name>Melissa</Name></User>
//
func (as *ApiService) create(request *restful.Request, response *restful.Response) {
	authInfo := as.AuthInfo(request, response)
	if !as.IsAdmin(authInfo) {
		response.WriteErrorString(http.StatusBadRequest, "Invalid Request")
		return
	}

	data := bson.M{"_id": bson.NewObjectId()}
	if err := request.ReadEntity(&data); err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	if _, ok := data["username"]; ok {
		data["password"] = GenPasswordHash(data["password"].(string))
	}

	if err := models.Create(as.collection, &data); err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	// cors(response)
	response.WriteEntity(bson.M{"data": data})
}

// DELETE http://localhost:8080/{noun_url}/1
//
func (as *ApiService) remove(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("id")

	if len(id) == 0 {
		response.WriteErrorString(http.StatusBadRequest, "Invalid Request")
		return
	}

	models.Remove(as.collection, id)

	// cors(response)
	response.WriteHeader(200)
}

func enableCORS(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	resp.AddHeader("Access-Control-Allow-Origin", "http://"+strings.Replace(req.Request.Host, ":8080", "", 1))
	resp.AddHeader("Access-Control-Allow-Credentials", "true")
	resp.AddHeader("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE")
	resp.AddHeader("Access-Control-Allow-Headers", "Content-Type, Accept, Authorization, X-Requested-With")
	resp.AddHeader("Access-Control-Max-Age", "28800")
	resp.AddHeader("Content-Type", "application/json")
	chain.ProcessFilter(req, resp)
}

func enableOptions(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	if "OPTIONS" != req.Request.Method {
		chain.ProcessFilter(req, resp)
		return
	}
	resp.AddHeader(restful.HEADER_Allow, "POST, GET, PUT, DELETE")
}

func Run() {
	database.Init()
	defer database.GMyDb.Destroy()

	registerAll()

	restful.Filter(enableCORS)
	restful.Filter(enableOptions)

	// Optionally, you can install the Swagger Service which provides a nice Web UI on your REST API
	// You need to download the Swagger HTML5 assets and change the FilePath location in the config below.
	// Open http://localhost:8080/apidocs and enter http://localhost:8080/apidocs.json in the api input field.
	config := swagger.Config{
		WebServices:    restful.RegisteredWebServices(), // you control what services are visible
		WebServicesUrl: "/",
		ApiPath:        "/apidocs.json",

		// Optionally, specifiy where the UI is located
		SwaggerPath:     "/apidocs/",
		SwaggerFilePath: "./swagger-ui2/dist"}
	swagger.InstallSwaggerService(config)

	// log.Printf("start listening on :8080")
	// server := &http.Server{Addr: "10.10.1.94:8080", Handler: wsContainer}
	// log.Fatal(server.ListenAndServe())

	log.Printf("start listening on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
