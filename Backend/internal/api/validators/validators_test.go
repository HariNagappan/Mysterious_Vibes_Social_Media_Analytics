package validators

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestParsePaginationBounds(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest("GET", "/x?limit=5000&offset=10", nil)

	page := ParsePagination(c, 20, 100)
	if page.Limit != 100 {
		t.Fatalf("limit should clamp to max 100, got %d", page.Limit)
	}
	if page.Offset != 10 {
		t.Fatalf("offset = %d", page.Offset)
	}
}

func TestParsePaginationDefaults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest("GET", "/x", nil)

	page := ParsePagination(c, 25, 100)
	if page.Limit != 25 || page.Offset != 0 {
		t.Fatalf("unexpected defaults: %+v", page)
	}
}

func TestParseIDParamRejectsInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Params = gin.Params{{Key: "id", Value: "abc"}}

	if _, ok := ParseIDParam(c, "id"); ok {
		t.Fatal("expected invalid id to be rejected")
	}
	if recorder.Code != 400 {
		t.Fatalf("expected HTTP 400, got %d", recorder.Code)
	}
}

func TestEnumValidators(t *testing.T) {
	if !ValidRole("analyst") || ValidRole("superuser") {
		t.Fatal("role validation is wrong")
	}
	if !ValidPlatform("telegram") || ValidPlatform("myspace") {
		t.Fatal("platform validation is wrong")
	}
}
