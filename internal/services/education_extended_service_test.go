package services

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

func TestEducationServiceNewInstitutionMethods(t *testing.T) {
	t.Parallel()

	stub := &educationRepositoryStub{
		polytechnics: []models.Polytechnic{{
			ID: "abdu-gusau-polytechnic", Name: "Abdu Gusau Polytechnic", OwnershipType: "state", StateID: "zamfara", CountryCode: "NG",
		}},
		polytechnicGet: map[string]models.Polytechnic{
			"abdu-gusau-polytechnic": {ID: "abdu-gusau-polytechnic", Name: "Abdu Gusau Polytechnic", OwnershipType: "state", StateID: "zamfara", CountryCode: "NG"},
		},
		monotechnics: []models.Monotechnic{{
			ID: "abuja-school-of-pension-and-retirement-planning-karu", Name: "Abuja School of Pension and Retirement Planning, Karu", OwnershipType: "private", StateID: "fct", CountryCode: "NG",
		}},
		monotechnicGet: map[string]models.Monotechnic{
			"abuja-school-of-pension-and-retirement-planning-karu": {ID: "abuja-school-of-pension-and-retirement-planning-karu", Name: "Abuja School of Pension and Retirement Planning, Karu", OwnershipType: "private", StateID: "fct", CountryCode: "NG"},
		},
		collegesOfAgriculture: []models.CollegeOfAgriculture{{
			ID: "audu-bako-college-of-agriculture-danbatta", Name: "Audu Bako College of Agriculture, Danbatta", OwnershipType: "state", StateID: "kano", CountryCode: "NG",
		}},
		collegesOfAgricultureGet: map[string]models.CollegeOfAgriculture{
			"audu-bako-college-of-agriculture-danbatta": {ID: "audu-bako-college-of-agriculture-danbatta", Name: "Audu Bako College of Agriculture, Danbatta", OwnershipType: "state", StateID: "kano", CountryCode: "NG"},
		},
		collegesOfHealthSciencesAndTechnology: []models.CollegeOfHealthSciencesAndTechnology{{
			ID: "abia-state-college-of-health-technology-aba", Name: "Abia State College of Health Technology, Aba", OwnershipType: "state", StateID: "abia", CountryCode: "NG",
		}},
		collegesOfHealthSciencesAndTechnologyGet: map[string]models.CollegeOfHealthSciencesAndTechnology{
			"abia-state-college-of-health-technology-aba": {ID: "abia-state-college-of-health-technology-aba", Name: "Abia State College of Health Technology, Aba", OwnershipType: "state", StateID: "abia", CountryCode: "NG"},
		},
		collegesOfNursingAndMidwifery: []models.CollegeOfNursingAndMidwifery{{
			ID: "abia-state-college-of-nursing-sciences-amachara", Name: "Abia State College of Nursing Sciences, Amachara", OwnershipType: "state", StateID: "abia", CountryCode: "NG",
		}},
		collegesOfNursingAndMidwiferyGet: map[string]models.CollegeOfNursingAndMidwifery{
			"abia-state-college-of-nursing-sciences-amachara": {ID: "abia-state-college-of-nursing-sciences-amachara", Name: "Abia State College of Nursing Sciences, Amachara", OwnershipType: "state", StateID: "abia", CountryCode: "NG"},
		},
		technicalColleges: []models.TechnicalCollege{{
			ID: "agbor-technical-college-agbor", Name: "Agbor Technical College, Agbor", OwnershipType: "state", StateID: "delta", CountryCode: "NG",
		}},
		technicalCollegesGet: map[string]models.TechnicalCollege{
			"agbor-technical-college-agbor": {ID: "agbor-technical-college-agbor", Name: "Agbor Technical College, Agbor", OwnershipType: "state", StateID: "delta", CountryCode: "NG"},
		},
		vocationalEnterpriseInstitutions: []models.VocationalEnterpriseInstitution{{
			ID: "adhama-innovation-enterprise-institute", Name: "Adhama Innovation Enterprise Institute", OwnershipType: "private", StateID: "kano", CountryCode: "NG",
		}},
		vocationalEnterpriseInstitutionsGet: map[string]models.VocationalEnterpriseInstitution{
			"adhama-innovation-enterprise-institute": {ID: "adhama-innovation-enterprise-institute", Name: "Adhama Innovation Enterprise Institute", OwnershipType: "private", StateID: "kano", CountryCode: "NG"},
		},
	}

	svc, err := NewEducationService(stub)
	if err != nil {
		t.Fatalf("NewEducationService() error = %v", err)
	}

	if rows, err := svc.ListPolytechnics(context.Background()); err != nil || len(rows) != 1 || rows[0].ID != "abdu-gusau-polytechnic" {
		t.Fatalf("ListPolytechnics() => %#v, %v", rows, err)
	}
	if got, err := svc.GetPolytechnic(context.Background(), "  abdu-gusau-polytechnic  "); err != nil || got.ID != "abdu-gusau-polytechnic" {
		t.Fatalf("GetPolytechnic() => %#v, %v", got, err)
	}
	if rows, err := svc.ListMonotechnics(context.Background()); err != nil || len(rows) != 1 || rows[0].ID != "abuja-school-of-pension-and-retirement-planning-karu" {
		t.Fatalf("ListMonotechnics() => %#v, %v", rows, err)
	}
	if got, err := svc.GetMonotechnic(context.Background(), "abuja-school-of-pension-and-retirement-planning-karu"); err != nil || got.ID != "abuja-school-of-pension-and-retirement-planning-karu" {
		t.Fatalf("GetMonotechnic() => %#v, %v", got, err)
	}
	if rows, err := svc.ListCollegesOfAgriculture(context.Background()); err != nil || len(rows) != 1 || rows[0].ID != "audu-bako-college-of-agriculture-danbatta" {
		t.Fatalf("ListCollegesOfAgriculture() => %#v, %v", rows, err)
	}
	if got, err := svc.GetCollegeOfAgriculture(context.Background(), "audu-bako-college-of-agriculture-danbatta"); err != nil || got.ID != "audu-bako-college-of-agriculture-danbatta" {
		t.Fatalf("GetCollegeOfAgriculture() => %#v, %v", got, err)
	}
	if rows, err := svc.ListCollegesOfHealthSciencesAndTechnology(context.Background()); err != nil || len(rows) != 1 || rows[0].ID != "abia-state-college-of-health-technology-aba" {
		t.Fatalf("ListCollegesOfHealthSciencesAndTechnology() => %#v, %v", rows, err)
	}
	if got, err := svc.GetCollegeOfHealthSciencesAndTechnology(context.Background(), "abia-state-college-of-health-technology-aba"); err != nil || got.ID != "abia-state-college-of-health-technology-aba" {
		t.Fatalf("GetCollegeOfHealthSciencesAndTechnology() => %#v, %v", got, err)
	}
	if rows, err := svc.ListCollegesOfNursingAndMidwifery(context.Background()); err != nil || len(rows) != 1 || rows[0].ID != "abia-state-college-of-nursing-sciences-amachara" {
		t.Fatalf("ListCollegesOfNursingAndMidwifery() => %#v, %v", rows, err)
	}
	if got, err := svc.GetCollegeOfNursingAndMidwifery(context.Background(), "  abia-state-college-of-nursing-sciences-amachara  "); err != nil || got.ID != "abia-state-college-of-nursing-sciences-amachara" {
		t.Fatalf("GetCollegeOfNursingAndMidwifery() => %#v, %v", got, err)
	}
	if rows, err := svc.ListTechnicalColleges(context.Background()); err != nil || len(rows) != 1 || rows[0].ID != "agbor-technical-college-agbor" {
		t.Fatalf("ListTechnicalColleges() => %#v, %v", rows, err)
	}
	if got, err := svc.GetTechnicalCollege(context.Background(), "agbor-technical-college-agbor"); err != nil || got.ID != "agbor-technical-college-agbor" {
		t.Fatalf("GetTechnicalCollege() => %#v, %v", got, err)
	}
	if rows, err := svc.ListVocationalEnterpriseInstitutions(context.Background()); err != nil || len(rows) != 1 || rows[0].ID != "adhama-innovation-enterprise-institute" {
		t.Fatalf("ListVocationalEnterpriseInstitutions() => %#v, %v", rows, err)
	}
	if got, err := svc.GetVocationalEnterpriseInstitution(context.Background(), "adhama-innovation-enterprise-institute"); err != nil || got.ID != "adhama-innovation-enterprise-institute" {
		t.Fatalf("GetVocationalEnterpriseInstitution() => %#v, %v", got, err)
	}

	rows, err := svc.ListPolytechnics(context.Background())
	if err != nil {
		t.Fatalf("ListPolytechnics() second call error = %v", err)
	}
	rows[0].Name = "Changed"
	rowsAgain, err := svc.ListPolytechnics(context.Background())
	if err != nil {
		t.Fatalf("ListPolytechnics() third call error = %v", err)
	}
	if rowsAgain[0].Name == "Changed" {
		t.Fatal("ListPolytechnics() exposed shared slice state")
	}

	if _, err := svc.GetPolytechnic(context.Background(), "Invalid ID"); !errors.Is(err, ErrInvalidEducationInstitutionID) {
		t.Fatalf("GetPolytechnic invalid id error = %v, want ErrInvalidEducationInstitutionID", err)
	}
	if _, err := svc.GetCollegeOfNursingAndMidwifery(context.Background(), "Invalid ID"); !errors.Is(err, ErrInvalidEducationInstitutionID) {
		t.Fatalf("GetCollegeOfNursingAndMidwifery invalid id error = %v, want ErrInvalidEducationInstitutionID", err)
	}
	if _, err := svc.GetTechnicalCollege(context.Background(), "Invalid ID"); !errors.Is(err, ErrInvalidEducationInstitutionID) {
		t.Fatalf("GetTechnicalCollege invalid id error = %v, want ErrInvalidEducationInstitutionID", err)
	}
}

func TestEducationServiceNewInstitutionErrorTranslation(t *testing.T) {
	t.Parallel()

	stub := &educationRepositoryStub{
		collegesOfNursingAndMidwiferyGet: map[string]models.CollegeOfNursingAndMidwifery{},
		technicalCollegesGet:             map[string]models.TechnicalCollege{},
	}
	svc, err := NewEducationService(stub)
	if err != nil {
		t.Fatalf("NewEducationService() error = %v", err)
	}

	if _, err := svc.GetCollegeOfNursingAndMidwifery(context.Background(), "  abia-state-college-of-nursing-sciences-amachara  "); !errors.Is(err, ErrCollegeOfNursingAndMidwiferyNotFound) {
		t.Fatalf("GetCollegeOfNursingAndMidwifery() not-found translation = %v", err)
	}
	if _, err := svc.GetTechnicalCollege(context.Background(), "agbor-technical-college-agbor"); !errors.Is(err, ErrTechnicalCollegeNotFound) {
		t.Fatalf("GetTechnicalCollege() not-found translation = %v", err)
	}

	stub.collegesOfNursingAndMidwiferyErr = context.Canceled
	if _, err := svc.ListCollegesOfNursingAndMidwifery(context.Background()); !errors.Is(err, context.Canceled) {
		t.Fatalf("ListCollegesOfNursingAndMidwifery() canceled translation = %v", err)
	}

	deadlineCtx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	cancel()
	if _, err := svc.GetTechnicalCollege(deadlineCtx, "agbor-technical-college-agbor"); !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("GetTechnicalCollege() deadline translation = %v", err)
	}

	stub.technicalCollegesErr = errors.New("/private/tmp/education/technical_colleges.json: permission denied")
	if _, err := svc.ListTechnicalColleges(context.Background()); err == nil || strings.Contains(err.Error(), "/private/tmp/education/technical_colleges.json") {
		t.Fatalf("ListTechnicalColleges() sanitization = %v", err)
	}
}

func TestEducationServicePrimaryAndSecondarySchoolValidationAndTranslation(t *testing.T) {
	t.Parallel()

	stub := &educationRepositoryStub{
		primaryAndSecondarySchoolListResult: interfaces.PrimaryAndSecondarySchoolListResult{
			Schools: []models.PrimaryAndSecondarySchool{{
				ID:              "na-ube-junior-secondary-school-ajaka-kogi-igalamela-odolu-ajaka",
				Name:            "[(NA) UBE JUNIOR SECONDARY SCHOOL, AJAKA]",
				OwnershipType:   "public",
				StateID:         "kogi",
				LGAID:           "kogi-igalamela-odolu",
				CountryCode:     "NG",
				EducationLevels: []string{"junior-secondary"},
				UBECSchoolCode:  "22080001",
			}},
			Page: 1, PageSize: 50, Total: 1, TotalPages: 1,
		},
		primaryAndSecondarySchoolGetResult: map[string]models.PrimaryAndSecondarySchool{
			"na-ube-junior-secondary-school-ajaka-kogi-igalamela-odolu-ajaka": {
				ID:              "na-ube-junior-secondary-school-ajaka-kogi-igalamela-odolu-ajaka",
				Name:            "(NA) UBE JUNIOR SECONDARY SCHOOL, AJAKA",
				OwnershipType:   "public",
				StateID:         "kogi",
				LGAID:           "kogi-igalamela-odolu",
				CountryCode:     "NG",
				EducationLevels: []string{"junior-secondary"},
				UBECSchoolCode:  "22080001",
			},
		},
	}
	svc, err := NewEducationService(stub)
	if err != nil {
		t.Fatalf("NewEducationService() error = %v", err)
	}

	result, err := svc.ListPrimaryAndSecondarySchools(context.Background(), interfaces.PrimaryAndSecondarySchoolQuery{})
	if err != nil {
		t.Fatalf("ListPrimaryAndSecondarySchools() error = %v", err)
	}
	if len(result.Schools) != 1 || result.Total != 1 || result.Page != 1 || result.PageSize != 50 {
		t.Fatalf("unexpected school result: %#v", result)
	}
	result.Schools[0].EducationLevels[0] = "changed"
	again, err := svc.ListPrimaryAndSecondarySchools(context.Background(), interfaces.PrimaryAndSecondarySchoolQuery{})
	if err != nil {
		t.Fatalf("ListPrimaryAndSecondarySchools() second error = %v", err)
	}
	if again.Schools[0].EducationLevels[0] == "changed" {
		t.Fatal("ListPrimaryAndSecondarySchools() exposed shared education level backing slice")
	}

	school, err := svc.GetPrimaryAndSecondarySchool(context.Background(), "  na-ube-junior-secondary-school-ajaka-kogi-igalamela-odolu-ajaka  ")
	if err != nil {
		t.Fatalf("GetPrimaryAndSecondarySchool() error = %v", err)
	}
	if school.ID != "na-ube-junior-secondary-school-ajaka-kogi-igalamela-odolu-ajaka" {
		t.Fatalf("unexpected school result: %#v", school)
	}

	for _, query := range []interfaces.PrimaryAndSecondarySchoolQuery{
		{Page: -1},
		{Page: 1, PageSize: 101},
		{StateID: "Abia"},
		{LGAID: "bad id"},
		{EducationLevel: "tertiary"},
		{OwnershipType: "public-school"},
		{Search: strings.Repeat("x", maxSchoolSearchLength+1)},
	} {
		if _, err := svc.ListPrimaryAndSecondarySchools(context.Background(), query); err == nil {
			t.Fatalf("expected validation error for %#v", query)
		}
	}

	stub.primaryAndSecondarySchoolListErr = interfaces.ErrInvalidPrimaryAndSecondarySchoolQuery
	if _, err := svc.ListPrimaryAndSecondarySchools(context.Background(), interfaces.PrimaryAndSecondarySchoolQuery{LGAID: "kogi-igalamela-odolu"}); !errors.Is(err, ErrInvalidSchoolLGAFilter) {
		t.Fatalf("unexpected school translation: %v", err)
	}

	stub.primaryAndSecondarySchoolGetErr = interfaces.ErrPrimaryAndSecondarySchoolNotFound
	if _, err := svc.GetPrimaryAndSecondarySchool(context.Background(), "missing-school"); !errors.Is(err, ErrPrimaryAndSecondarySchoolNotFound) {
		t.Fatalf("GetPrimaryAndSecondarySchool() not-found translation = %v", err)
	}

	stub.primaryAndSecondarySchoolGetErr = errors.New("/private/tmp/education/primary_and_secondary_schools.json: permission denied")
	if _, err := svc.GetPrimaryAndSecondarySchool(context.Background(), "na-ube-junior-secondary-school-ajaka-kogi-igalamela-odolu-ajaka"); err == nil || strings.Contains(err.Error(), "/private/tmp/education/primary_and_secondary_schools.json") {
		t.Fatalf("unexpected school get sanitization: %v", err)
	}
}
