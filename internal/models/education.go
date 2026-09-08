package models

// University represents one current NUC-listed Nigerian university.
type University struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	OwnershipType string `json:"ownership_type"`
	StateID       string `json:"state_id"`
	CountryCode   string `json:"country_code"`
}

// CollegeOfEducation represents one current NCCE-listed Nigerian college of education.
type CollegeOfEducation struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	OwnershipType string `json:"ownership_type"`
	StateID       string `json:"state_id"`
	CountryCode   string `json:"country_code"`
}

// Polytechnic represents one current NBTE-listed Nigerian polytechnic.
type Polytechnic struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	OwnershipType string `json:"ownership_type"`
	StateID       string `json:"state_id"`
	CountryCode   string `json:"country_code"`
}

// Monotechnic represents one current NBTE-listed specialised institution.
type Monotechnic struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	OwnershipType string `json:"ownership_type"`
	StateID       string `json:"state_id"`
	CountryCode   string `json:"country_code"`
}

// CollegeOfAgriculture represents one current NBTE-listed Nigerian college of agriculture.
type CollegeOfAgriculture struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	OwnershipType string `json:"ownership_type"`
	StateID       string `json:"state_id"`
	CountryCode   string `json:"country_code"`
}

// CollegeOfHealthSciencesAndTechnology represents one current NBTE-listed Nigerian health-technology college.
type CollegeOfHealthSciencesAndTechnology struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	OwnershipType string `json:"ownership_type"`
	StateID       string `json:"state_id"`
	CountryCode   string `json:"country_code"`
}

// CollegeOfNursingAndMidwifery represents one current NMCN-approved Nigerian nursing institution.
type CollegeOfNursingAndMidwifery struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	OwnershipType string `json:"ownership_type"`
	StateID       string `json:"state_id"`
	CountryCode   string `json:"country_code"`
}

// VocationalEnterpriseInstitution represents one current Digital NBTE-listed VEI.
type VocationalEnterpriseInstitution struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	OwnershipType string `json:"ownership_type"`
	StateID       string `json:"state_id"`
	CountryCode   string `json:"country_code"`
}
