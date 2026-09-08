package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

type extendedEducationHandlerStub struct {
	educationHandlerStub
	listErr     error
	getErr      error
	lastSchoolQ interfaces.PrimaryAndSecondarySchoolQuery
}

func (s *extendedEducationHandlerStub) ListPolytechnics(context.Context) ([]models.Polytechnic, error) {
	return []models.Polytechnic{{ID: "sample-polytechnic", Name: "Sample Polytechnic", OwnershipType: "federal", StateID: "lagos", CountryCode: "NG"}}, s.listErr
}
func (s *extendedEducationHandlerStub) GetPolytechnic(context.Context, string) (models.Polytechnic, error) {
	return models.Polytechnic{ID: "sample-polytechnic", Name: "Sample Polytechnic", OwnershipType: "federal", StateID: "lagos", CountryCode: "NG"}, s.getErr
}
func (s *extendedEducationHandlerStub) ListMonotechnics(context.Context) ([]models.Monotechnic, error) {
	return []models.Monotechnic{{ID: "sample-monotechnic", Name: "Sample Monotechnic", OwnershipType: "state", StateID: "lagos", CountryCode: "NG"}}, s.listErr
}
func (s *extendedEducationHandlerStub) GetMonotechnic(context.Context, string) (models.Monotechnic, error) {
	return models.Monotechnic{ID: "sample-monotechnic", Name: "Sample Monotechnic", OwnershipType: "state", StateID: "lagos", CountryCode: "NG"}, s.getErr
}
func (s *extendedEducationHandlerStub) ListCollegesOfAgriculture(context.Context) ([]models.CollegeOfAgriculture, error) {
	return []models.CollegeOfAgriculture{{ID: "sample-college-of-agriculture", Name: "Sample College of Agriculture", OwnershipType: "private", StateID: "lagos", CountryCode: "NG"}}, s.listErr
}
func (s *extendedEducationHandlerStub) GetCollegeOfAgriculture(context.Context, string) (models.CollegeOfAgriculture, error) {
	return models.CollegeOfAgriculture{ID: "sample-college-of-agriculture", Name: "Sample College of Agriculture", OwnershipType: "private", StateID: "lagos", CountryCode: "NG"}, s.getErr
}
func (s *extendedEducationHandlerStub) ListCollegesOfHealthSciencesAndTechnology(context.Context) ([]models.CollegeOfHealthSciencesAndTechnology, error) {
	return []models.CollegeOfHealthSciencesAndTechnology{{ID: "sample-health-college", Name: "Sample Health College", OwnershipType: "federal", StateID: "lagos", CountryCode: "NG"}}, s.listErr
}
func (s *extendedEducationHandlerStub) GetCollegeOfHealthSciencesAndTechnology(context.Context, string) (models.CollegeOfHealthSciencesAndTechnology, error) {
	return models.CollegeOfHealthSciencesAndTechnology{ID: "sample-health-college", Name: "Sample Health College", OwnershipType: "federal", StateID: "lagos", CountryCode: "NG"}, s.getErr
}
func (s *extendedEducationHandlerStub) ListCollegesOfNursingAndMidwifery(context.Context) ([]models.CollegeOfNursingAndMidwifery, error) {
	return []models.CollegeOfNursingAndMidwifery{{ID: "sample-nursing-college", Name: "Sample Nursing College", OwnershipType: "state", StateID: "lagos", CountryCode: "NG"}}, s.listErr
}
func (s *extendedEducationHandlerStub) GetCollegeOfNursingAndMidwifery(context.Context, string) (models.CollegeOfNursingAndMidwifery, error) {
	return models.CollegeOfNursingAndMidwifery{ID: "sample-nursing-college", Name: "Sample Nursing College", OwnershipType: "state", StateID: "lagos", CountryCode: "NG"}, s.getErr
}
func (s *extendedEducationHandlerStub) ListVocationalEnterpriseInstitutions(context.Context) ([]models.VocationalEnterpriseInstitution, error) {
	return []models.VocationalEnterpriseInstitution{{ID: "sample-vei", Name: "Sample VEI", OwnershipType: "private", StateID: "lagos", CountryCode: "NG"}}, s.listErr
}
func (s *extendedEducationHandlerStub) GetVocationalEnterpriseInstitution(context.Context, string) (models.VocationalEnterpriseInstitution, error) {
	return models.VocationalEnterpriseInstitution{ID: "sample-vei", Name: "Sample VEI", OwnershipType: "private", StateID: "lagos", CountryCode: "NG"}, s.getErr
}
func (s *extendedEducationHandlerStub) ListTechnicalColleges(context.Context) ([]models.TechnicalCollege, error) {
	return []models.TechnicalCollege{{ID: "sample-technical-college", Name: "Sample Technical College", OwnershipType: "state", StateID: "lagos", CountryCode: "NG"}}, s.listErr
}
func (s *extendedEducationHandlerStub) GetTechnicalCollege(context.Context, string) (models.TechnicalCollege, error) {
	return models.TechnicalCollege{ID: "sample-technical-college", Name: "Sample Technical College", OwnershipType: "state", StateID: "lagos", CountryCode: "NG"}, s.getErr
}
func (s *extendedEducationHandlerStub) ListPrimaryAndSecondarySchools(_ context.Context, query interfaces.PrimaryAndSecondarySchoolQuery) (interfaces.PrimaryAndSecondarySchoolListResult, error) {
	s.lastSchoolQ = query
	return interfaces.PrimaryAndSecondarySchoolListResult{Schools: []models.PrimaryAndSecondarySchool{{ID: "sample-school", Name: "Sample School", StateID: "lagos", CountryCode: "NG", EducationLevels: []string{"primary"}}}, Page: query.Page, PageSize: query.PageSize, Total: 166604, TotalPages: 3333}, s.listErr
}
func (s *extendedEducationHandlerStub) GetPrimaryAndSecondarySchool(context.Context, string) (models.PrimaryAndSecondarySchool, error) {
	return models.PrimaryAndSecondarySchool{ID: "sample-school", Name: "Sample School", StateID: "lagos", CountryCode: "NG", EducationLevels: []string{"primary"}}, s.getErr
}

func TestEducationHandlerExtendedInstitutionContracts(t *testing.T) {
	tests := []struct {
		name     string
		listPath string
		getPath  string
		list     func(*EducationHandler, http.ResponseWriter, *http.Request)
		get      func(*EducationHandler, http.ResponseWriter, *http.Request)
	}{
		{"polytechnics", "/v1/education/polytechnics", "/v1/education/polytechnics/sample-polytechnic", (*EducationHandler).ListPolytechnics, (*EducationHandler).GetPolytechnic},
		{"monotechnics", "/v1/education/monotechnics", "/v1/education/monotechnics/sample-monotechnic", (*EducationHandler).ListMonotechnics, (*EducationHandler).GetMonotechnic},
		{"agriculture", "/v1/education/colleges-of-agriculture", "/v1/education/colleges-of-agriculture/sample-college-of-agriculture", (*EducationHandler).ListCollegesOfAgriculture, (*EducationHandler).GetCollegeOfAgriculture},
		{"health", "/v1/education/colleges-of-health-sciences-and-technology", "/v1/education/colleges-of-health-sciences-and-technology/sample-health-college", (*EducationHandler).ListCollegesOfHealthSciencesAndTechnology, (*EducationHandler).GetCollegeOfHealthSciencesAndTechnology},
		{"nursing", "/v1/education/colleges-of-nursing-and-midwifery", "/v1/education/colleges-of-nursing-and-midwifery/sample-nursing-college", (*EducationHandler).ListCollegesOfNursingAndMidwifery, (*EducationHandler).GetCollegeOfNursingAndMidwifery},
		{"vei", "/v1/education/vocational-enterprise-institutions", "/v1/education/vocational-enterprise-institutions/sample-vei", (*EducationHandler).ListVocationalEnterpriseInstitutions, (*EducationHandler).GetVocationalEnterpriseInstitution},
		{"technical", "/v1/education/technical-colleges", "/v1/education/technical-colleges/sample-technical-college", (*EducationHandler).ListTechnicalColleges, (*EducationHandler).GetTechnicalCollege},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h, err := NewEducationHandler(&extendedEducationHandlerStub{})
			if err != nil {
				t.Fatal(err)
			}
			listResponse := httptest.NewRecorder()
			tc.list(h, listResponse, httptest.NewRequest(http.MethodGet, tc.listPath, nil))
			if listResponse.Code != http.StatusOK || !strings.Contains(listResponse.Body.String(), `"data":[`) {
				t.Fatalf("list response: %d %s", listResponse.Code, listResponse.Body.String())
			}
			getRequest := httptest.NewRequest(http.MethodGet, tc.getPath, nil)
			getRequest.SetPathValue("institution_id", " sample-polytechnic ")
			if tc.name != "polytechnics" {
				getRequest.SetPathValue("institution_id", "sample-"+map[string]string{"monotechnics": "monotechnic", "agriculture": "college-of-agriculture", "health": "health-college", "nursing": "nursing-college", "vei": "vei", "technical": "technical-college"}[tc.name])
			}
			getResponse := httptest.NewRecorder()
			tc.get(h, getResponse, getRequest)
			if getResponse.Code != http.StatusOK {
				t.Fatalf("get response: %d %s", getResponse.Code, getResponse.Body.String())
			}
		})
	}
}

func TestEducationHandlerSchoolPaginationAndValidation(t *testing.T) {
	stub := &extendedEducationHandlerStub{}
	h, err := NewEducationHandler(stub)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/education/primary-and-secondary-schools?page=2&page_size=100&state_id="+url.QueryEscape(" lagos ")+"&education_level=primary&ownership_type=public&search="+url.QueryEscape("  Community School  "), nil)
	h.ListPrimaryAndSecondarySchools(rr, req)
	if rr.Code != http.StatusOK || stub.lastSchoolQ.Page != 2 || stub.lastSchoolQ.PageSize != 100 || stub.lastSchoolQ.StateID != "lagos" || stub.lastSchoolQ.Search != "Community School" {
		t.Fatalf("unexpected school response/query: %d %#v", rr.Code, stub.lastSchoolQ)
	}
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil || body["data"] == nil || body["meta"] == nil {
		t.Fatalf("unexpected school payload: %s", rr.Body.String())
	}

	for _, query := range []string{"page=0", "page_size=101", "page=1.5", "page=999999999999999999999999", "education_level=tertiary", "ownership_type=government", "search=" + strings.Repeat("x", 101)} {
		rr = httptest.NewRecorder()
		h.ListPrimaryAndSecondarySchools(rr, httptest.NewRequest(http.MethodGet, "/v1/education/primary-and-secondary-schools?"+query, nil))
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("query %q status = %d", query, rr.Code)
		}
	}
}

func TestEducationHandlerExtendedErrorsAreSanitized(t *testing.T) {
	stub := &extendedEducationHandlerStub{listErr: errors.New("/private/tmp/secret.json: decoder details")}
	h, err := NewEducationHandler(stub)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	h.ListTechnicalColleges(rr, httptest.NewRequest(http.MethodGet, "/v1/education/technical-colleges", nil))
	if rr.Code != http.StatusInternalServerError || strings.Contains(rr.Body.String(), "secret.json") {
		t.Fatalf("unexpected error response: %d %s", rr.Code, rr.Body.String())
	}
}
