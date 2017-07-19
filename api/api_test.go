/*
 * Copyright Morpheo Org. 2017
 *
 * contact@morpheo.co
 *
 * This software is part of the Morpheo project, an open-source machine
 * learning platform.
 *
 * This software is governed by the CeCILL license, compatible with the
 * GNU GPL, under French law and abiding by the rules of distribution of
 * free software. You can  use, modify and/ or redistribute the software
 * under the terms of the CeCILL license as circulated by CEA, CNRS and
 * INRIA at the following URL "http://www.cecill.info".
 *
 * As a counterpart to the access to the source code and  rights to copy,
 * modify and redistribute granted by the license, users are provided only
 * with a limited warranty  and the software's author,  the holder of the
 * economic rights,  and the successive licensors  have only  limited
 * liability.
 *
 * In this respect, the user's attention is drawn to the risks associated
 * with loading,  using,  modifying and/or developing or reproducing the
 * software by the user in light of its specific status of free software,
 * that may mean  that it is complicated to manipulate,  and  that  also
 * therefore means  that it is reserved for developers  and  experienced
 * professionals having in-depth computer knowledge. Users are therefore
 * encouraged to load and test the software's suitability as regards their
 * requirements in conditions enabling the security of their systems and/or
 * data to be ensured and,  more generally, to use and operate it in the
 * same conditions as regards security.
 *
 * The fact that you are presently reading this means that you have had
 * knowledge of the CeCILL license and that you accept its terms.
 */

package main_test

import (
	"fmt"
	"math/rand"
	"os"
	"path"
	"strconv"
	"testing"

	"github.com/MorpheoOrg/go-packages/common"
	. "github.com/MorpheoOrg/storage/api"
	"github.com/satori/go.uuid"
	"gopkg.in/kataras/iris.v6"
	"gopkg.in/kataras/iris.v6/adaptors/httprouter"
	"gopkg.in/kataras/iris.v6/httptest"
)

var (
	app          *iris.Framework
	objectRoutes = []string{
		ProblemListRoute, ProblemRoute, ProblemBlobRoute,
		DataListRoute, DataRoute, DataBlobRoute,
		AlgoListRoute, AlgoRoute, AlgoBlobRoute,
		ModelListRoute, ModelRoute, ModelBlobRoute,
	}
	listObjectRoutes = []string{DataListRoute, ProblemListRoute, AlgoListRoute, ModelListRoute}

	RandomUUID           uuid.UUID
	READMEPath           string
	MultipartFormMap     map[string]map[string]string
	MultipartFormUUIDMap map[string]map[string]string
)

func TestMain(m *testing.M) {
	fmt.Printf("Test starting bitch!\n")
	app = setTestApp()

	// Set Random UUID & valid Multipart/form-data fields
	RandomUUID = uuid.NewV4()
	MultipartFormMap, MultipartFormUUIDMap = NewMultipartFormMap(RandomUUID)

	// Get Readme.md path for description form field
	wd, _ := os.Getwd()
	pd, _ := path.Split(wd)
	READMEPath = path.Join(pd, "README.md")
	os.Exit(m.Run())
}

// Test valid public request returns Success
func TestPublicRoute(t *testing.T) {
	e := httptest.New(app, t)

	e.GET(RootRoute).Expect().Status(200)
	e.GET(HealthRoute).Expect().Status(200).JSON().Equal(map[string]interface{}{"status": "ok"})
}

func TestRouteAuthentication(t *testing.T) {
	e := httptest.New(app, t)

	for _, url := range objectRoutes {
		t.Logf(url)

		// Test access unauthorized without valid authentication
		e.GET(url).Expect().Status(401)
		e.GET(url).WithBasicAuth("invalid", "invalid").Expect().Status(401)
	}
}

func TestGetListObject(t *testing.T) {
	e := httptest.New(app, t)

	for _, url := range listObjectRoutes {
		t.Logf(url)

		// Test valid request returns Success
		e.GET(url).WithBasicAuth("u", "p").Expect().Status(200)
	}
}

func TestGetObject(t *testing.T) {
	e := httptest.New(app, t)

	for _, url := range listObjectRoutes {
		t.Logf(url)

		// Test valid request returns Success
		e.GET(url+"/"+RandomUUID.String()).WithBasicAuth("u", "p").Expect().Status(200)

		// Test invalid uuid returns BadRequest
		e.GET(url+"/666devil").WithBasicAuth("u", "p").Expect().Status(400).Body().Match("(.*)Impossible to parse UUID(.*)")

		// Test uuid not in db returns NotFound
		e.GET(url+"/"+DevilMockUUID).WithBasicAuth("u", "p").Expect().Status(404).Body().Match("{(.*)sql: no rows in result set\"}")
	}
}

func TestGetObjectBlob(t *testing.T) {
	e := httptest.New(app, t)

	for _, url := range listObjectRoutes {
		t.Logf(url)

		// Test valid request returns Success
		e.GET(url+"/"+RandomUUID.String()+"/blob").WithBasicAuth("u", "p").Expect().Status(200)

		// Test invalid uuid returns BadRequest and error 'Impossible to parse'
		e.GET(url+"/666devil/blob").WithBasicAuth("u", "p").Expect().Status(400).Body().Match("(.*)Impossible to parse UUID(.*)")

		// Test uuid not in db returns NotFound
		e.GET(url+"/"+DevilMockUUID+"/blob").WithBasicAuth("u", "p").Expect().Status(404).Body().Match("{(.*)sql: no rows in result set\"}")

		// test download failed returns internalServerError
		e.GET(url+"/"+common.ViciousDevilUUID+"/blob").WithBasicAuth("u", "p").Expect().Status(500)
	}
}

func TestPostObjectMultipart(t *testing.T) {
	e := httptest.New(app, t)

	for _, url := range []string{ProblemListRoute} {
		t.Logf(url)

		// Test valid request with UUID returns Success
		e.POST(url).WithBasicAuth("u", "p").WithMultipart().WithForm(MultipartFormUUIDMap[url]).WithFile("description", READMEPath).WithFormField("size", "666").WithFile("blob", "config.go").Expect().Status(201).Body().Match("(.*)" + RandomUUID.String() + "(.*)")

		// Test valid request without UUID returns Success
		e.POST(url).WithBasicAuth("u", "p").WithMultipart().WithForm(MultipartFormMap[url]).WithFile("description", READMEPath).WithFormField("size", "666").WithFile("blob", "config.go").Expect().Status(201)

		// Test size omission returns BadRequest
		e.POST(url).WithBasicAuth("u", "p").WithMultipart().WithForm(MultipartFormUUIDMap[url]).WithFile("description", READMEPath).WithFile("blob", "config.go").Expect().Status(400).Body().Match("(.*)'Size' unset(.*)")

		// Test failed file upload returns InternalServerError
		e.POST(url).WithBasicAuth("u", "p").WithMultipart().WithForm(MultipartFormUUIDMap[url]).WithFile("description", READMEPath).WithFormField("size", common.NaughtySize).WithFile("blob", "config.go").Expect().Status(500)

		// Test invalid description returns BadRequest
		e.POST(url).WithBasicAuth("u", "p").WithMultipart().WithForm(MultipartFormUUIDMap[url]).WithFile("description", "main.go").WithFormField("size", "666").WithFile("blob", "config.go").Expect().Status(400).Body().Match("(.*)description should be a '.md' file(.*)")
		e.POST(url).WithBasicAuth("u", "p").WithMultipart().WithForm(MultipartFormUUIDMap[url]).WithFormField("size", "666").WithFile("blob", "config.go").Expect().Status(400).Body().Match("(.*)'Description' unset(.*)")
		e.POST(url).WithBasicAuth("u", "p").WithMultipart().WithForm(MultipartFormUUIDMap[url]).WithFormField("description", "great description").WithFormField("size", "666").WithFile("blob", "config.go").Expect().Status(400).Body().Match("(.*)description should be a '.md' file(.*)")
	}

	for _, url := range []string{AlgoListRoute, DataListRoute} {
		t.Logf(url)

		// Test valid request with UUID returns Success
		e.POST(url).WithBasicAuth("u", "p").WithMultipart().WithForm(MultipartFormUUIDMap[url]).WithFormField("size", "666").WithFile("blob", "config.go").Expect().Status(201).Body().Match("(.*)" + RandomUUID.String() + "(.*)")

		// Test valid request without UUID returns Success
		e.POST(url).WithBasicAuth("u", "p").WithMultipart().WithForm(MultipartFormMap[url]).WithFormField("size", "666").WithFile("blob", "config.go").Expect().Status(201)

		// Test size omission returns BadRequest
		e.POST(url).WithBasicAuth("u", "p").WithMultipart().WithForm(MultipartFormUUIDMap[url]).WithFile("blob", "config.go").Expect().Status(400).Body().Match("(.*)'Size' unset(.*)")

		// Test failed file upload returns InternalServerError
		e.POST(url).WithBasicAuth("u", "p").WithMultipart().WithForm(MultipartFormUUIDMap[url]).WithFormField("size", common.NaughtySize).WithFile("blob", "config.go").Expect().Status(500)
	}

	for _, url := range []string{AlgoListRoute, DataListRoute, ProblemListRoute} {
		t.Logf(url)

		// Test request with invalid Content-Type header returns BadRequest
		e.POST(url).WithBasicAuth("u", "p").Expect().Status(400).Body().Match("(.*)Error parsing header(.*)")
		e.POST(url).WithBasicAuth("u", "p").WithHeader("Content-Type", "invalid").Expect().Status(400).Body().Match("(.*)Invalid media type(.*)")

		// Test invalid form fields returns BadRequest
		e.POST(url).WithBasicAuth("u", "p").WithMultipart().WithFormField("invalid", "aze").Expect().Status(400).Body().Match("(.*)Unknown field(.*)")
		e.POST(url).WithBasicAuth("u", "p").WithMultipart().WithFormField("uuid", "invalid").WithFile("blob", "config.go").Expect().Status(400).Body().Match("(.*)Error parsing UUID uuid(.*)")
		e.POST(url).WithBasicAuth("u", "p").WithMultipart().WithFormField("size", "invalid").Expect().Status(400).Body().Match("(.*)Error parsing size(.*)")

		// Test field blob not at the end returns BadRequest
		e.POST(url).WithBasicAuth("u", "p").WithMultipart().WithFile("blob", "config.go").WithForm(MultipartFormUUIDMap[url]).WithFormField("size", "666").Expect().Status(400)
	}

	// Test valid form field but not suited for Object returns BadRequest
	e.POST(DataListRoute).WithBasicAuth("u", "p").WithMultipart().WithForm(MultipartFormUUIDMap[ProblemListRoute]).WithFormField("size", "666").WithFile("blob", "config.go").Expect().Status(400)

	// Test big name/size returns BadRequest
	buf := make([]byte, StrFieldMaxLength+1)
	rand.Read(buf)
	e.POST(AlgoListRoute).WithBasicAuth("u", "p").WithMultipart().WithFormField("size", buf).Expect().Status(400).Body().Match("(.*)Buffer overflow reading size(.*)")
	e.POST(ProblemListRoute).WithBasicAuth("u", "p").WithMultipart().WithFormField("name", buf).Expect().Status(400).Body().Match("(.*)Buffer overflow reading name(.*)")
}

func TestPostModel(t *testing.T) {
	e := httptest.New(app, t)

	// Test valid request returns StatusCreated
	e.POST(ModelListRoute).WithQuery("algo", RandomUUID.String()).WithBasicAuth("u", "p").WithHeader("Content-Length", "15").WithBytes([]byte("fakefilecontent")).Expect().Status(201)

	// Test request with unvalid algo uuid returns BadRequest
	e.POST(ModelListRoute).WithQuery("algo", "7-Batman").WithBasicAuth("u", "p").WithHeader("Content-Length", "15").WithBytes([]byte("fakefilecontent")).Expect().Status(400)

	// Test request with unexistant algo uuid returns NotFound
	e.POST(ModelListRoute).WithQuery("algo", DevilMockUUID).WithBasicAuth("u", "p").WithHeader("Content-Length", "15").WithBytes([]byte("fakefilecontent")).Expect().Status(404).Body().Match(`{\"error\":\"Error uploading model: algorithm (.+) not found: Error retrieving algo (.+): (.*)\"}`)

	// Test failed file upload returns InternalServerError
	e.POST(ModelListRoute).WithQuery("algo", RandomUUID.String()).WithBasicAuth("u", "p").WithHeader("Content-Length", strconv.Itoa(common.NaughtySize)).WithBytes([]byte("fakefilecontent")).Expect().Status(500).Body().Match("(.*)What a naughty size(.*)")
}

func TestPatchProblem(t *testing.T) {
	e := httptest.New(app, t)

	// Test valid patch returns Success
	e.PATCH(ProblemListRoute+"/"+ProblemMockUUIDStr).WithBasicAuth("u", "p").WithMultipart().WithFile("description", READMEPath).WithFormField("uuid", uuid.NewV4()).Expect().Status(200)

	// Test used UUID returns Conflict
	e.PATCH(ProblemListRoute+"/"+ProblemMockUUIDStr).WithBasicAuth("u", "p").WithMultipart().WithFormField("uuid", ProblemMockUUIDStr).Expect().Status(409)

	// 	Test invalid name returns BadRequest
	e.PATCH(ProblemListRoute+"/"+ProblemMockUUIDStr).WithBasicAuth("u", "p").WithMultipart().WithFormField("name", "").Expect().Status(400).Body().Match("(.*)'Name' unset(.*)")
}

// setTestApp set up the Iris App for testing
func setTestApp() *iris.Framework {
	_ = NewStorageConfig()
	app := iris.New()
	app.Adapt(iris.DevLogger(), httprouter.New())
	auth := SetAuthentication("u", "p")

	// Set models configuration
	problemModel, _ := NewMockedModel(ProblemModelName)
	algoModel, _ := NewMockedModel(AlgoModelName)
	modelModel, _ := NewMockedModel(ModelModelName)
	dataModel, _ := NewMockedModel(DataModelName)

	// set Blobstore
	blobStore, _ := SetBlobStore("fake", "fake", "fake")
	// blobStore, _ := SetBlobStore("test_data", "", "")

	api := &APIServer{
		BlobStore:    blobStore,
		ProblemModel: problemModel,
		AlgoModel:    algoModel,
		ModelModel:   modelModel,
		DataModel:    dataModel,
	}
	api.ConfigureRoutes(app, auth)
	return app
}

// NewMultipartFormUUIDMap creates valid Multipart/form-data fields for each Resource
func NewMultipartFormMap(id uuid.UUID) (m map[string]map[string]string, mUUID map[string]map[string]string) {
	m = map[string]map[string]string{
		ProblemListRoute: map[string]string{
			"name":  "testName",
			"owner": uuid.NewV4().String(),
		},
		AlgoListRoute: map[string]string{
			"name":  "testName",
			"owner": uuid.NewV4().String(),
		},
		ModelListRoute: map[string]string{
			"owner": uuid.NewV4().String(),
		},
		DataListRoute: map[string]string{
			"owner": uuid.NewV4().String(),
		},
	}
	mUUID = m
	mUUID[ProblemListRoute]["uuid"] = id.String()
	mUUID[AlgoListRoute]["uuid"] = id.String()
	mUUID[ModelListRoute]["uuid"] = id.String()
	mUUID[DataListRoute]["uuid"] = id.String()
	return
}
