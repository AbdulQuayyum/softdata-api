package interfaces

import "errors"

var (
	// ErrNotFound reports a missing persistence record.
	ErrNotFound = errors.New("repository: not found")
	// ErrConflict reports a portable uniqueness conflict.
	ErrConflict = errors.New("repository: conflict")
	// ErrInvalidRateLimitInput reports an invalid rate-limit request.
	ErrInvalidRateLimitInput = errors.New("repository: invalid rate limit input")
	// ErrRateLimitUnavailable reports that Redis-backed rate limiting is unavailable.
	ErrRateLimitUnavailable = errors.New("repository: rate limit unavailable")
	// ErrInvalidDatasetPath reports an unsafe or unsupported dataset path.
	ErrInvalidDatasetPath = errors.New("repository: invalid dataset path")
	// ErrDatasetFileNotFound reports that the requested dataset file is missing.
	ErrDatasetFileNotFound = errors.New("repository: dataset file not found")
	// ErrDatasetFileTooLarge reports that the dataset file exceeds the configured limit.
	ErrDatasetFileTooLarge = errors.New("repository: dataset file too large")
	// ErrInvalidDatasetFile reports that the dataset file could not be decoded or validated.
	ErrInvalidDatasetFile = errors.New("repository: invalid dataset file")
	// ErrDatasetFileUnavailable reports that the dataset file cannot be accessed right now.
	ErrDatasetFileUnavailable = errors.New("repository: dataset file unavailable")
	// ErrStateNotFound reports that a requested state is not present in the dataset.
	ErrStateNotFound = errors.New("repository: state not found")
	// ErrGeopoliticalZoneNotFound reports that a requested geopolitical zone is not present in the dataset.
	ErrGeopoliticalZoneNotFound = errors.New("repository: geopolitical zone not found")
	// ErrLocalGovernmentUnitNotFound reports that a requested local-government unit is not present in the dataset.
	ErrLocalGovernmentUnitNotFound = errors.New("repository: local government unit not found")
	// ErrLanguageNotFound reports that a requested language is not present in the dataset.
	ErrLanguageNotFound = errors.New("repository: language not found")
	// ErrInvalidCountryLanguageCountryAreaID reports an invalid country/area filter.
	ErrInvalidCountryLanguageCountryAreaID = errors.New("repository: invalid country language country area id")
	// ErrInvalidCountryLanguageLanguageID reports an invalid language filter.
	ErrInvalidCountryLanguageLanguageID = errors.New("repository: invalid country language language id")
	// ErrInvalidCountryLanguageStatus reports an invalid relationship status filter.
	ErrInvalidCountryLanguageStatus = errors.New("repository: invalid country language status")
	// ErrTimeZoneNotFound reports that a requested time zone is not present in the dataset.
	ErrTimeZoneNotFound = errors.New("repository: time zone not found")
	// ErrCountryOrAreaNotFound reports that a requested country or area is not present in the dataset.
	ErrCountryOrAreaNotFound = errors.New("repository: country or area not found")
	// ErrUniversityNotFound reports that a requested university is not present in the dataset.
	ErrUniversityNotFound = errors.New("repository: university not found")
	// ErrCollegeOfEducationNotFound reports that a requested college of education is not present in the dataset.
	ErrCollegeOfEducationNotFound = errors.New("repository: college of education not found")
	// ErrPolytechnicNotFound reports that a requested polytechnic is not present in the dataset.
	ErrPolytechnicNotFound = errors.New("repository: polytechnic not found")
	// ErrMonotechnicNotFound reports that a requested monotechnic is not present in the dataset.
	ErrMonotechnicNotFound = errors.New("repository: monotechnic not found")
	// ErrCollegeOfAgricultureNotFound reports that a requested college of agriculture is not present in the dataset.
	ErrCollegeOfAgricultureNotFound = errors.New("repository: college of agriculture not found")
	// ErrCollegeOfHealthSciencesAndTechnologyNotFound reports that a requested health-sciences-and-technology college is not present in the dataset.
	ErrCollegeOfHealthSciencesAndTechnologyNotFound = errors.New("repository: college of health sciences and technology not found")
	// ErrCollegeOfNursingAndMidwiferyNotFound reports that a requested college of nursing and midwifery is not present in the dataset.
	ErrCollegeOfNursingAndMidwiferyNotFound = errors.New("repository: college of nursing and midwifery not found")
	// ErrVocationalEnterpriseInstitutionNotFound reports that a requested vocational enterprise institution is not present in the dataset.
	ErrVocationalEnterpriseInstitutionNotFound = errors.New("repository: vocational enterprise institution not found")
	// ErrTechnicalCollegeNotFound reports that a requested technical college is not present in the dataset.
	ErrTechnicalCollegeNotFound = errors.New("repository: technical college not found")
	// ErrPrimaryAndSecondarySchoolNotFound reports that a requested primary or secondary school is not present in the dataset.
	ErrPrimaryAndSecondarySchoolNotFound = errors.New("repository: primary and secondary school not found")
	// ErrHealthFacilityNotFound reports that a requested health facility is not present in the dataset.
	ErrHealthFacilityNotFound = errors.New("repository: health facility not found")
	// ErrMedicalLaboratoryAccreditationNotFound reports that a requested medical laboratory accreditation is not present in the dataset.
	ErrMedicalLaboratoryAccreditationNotFound = errors.New("repository: medical laboratory accreditation not found")
	// ErrNHIAAccreditedHealthMaintenanceOrganisationNotFound reports that a requested NHIA accredited HMO is not present in the dataset.
	ErrNHIAAccreditedHealthMaintenanceOrganisationNotFound = errors.New("repository: nhia accredited health maintenance organisation not found")
	// ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationQuery reports an invalid NHIA accredited HMO query.
	ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationQuery = errors.New("repository: invalid nhia accredited health maintenance organisation query")
	// ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationStatusFilter reports an unsupported NHIA accredited HMO status filter.
	ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationStatusFilter = errors.New("repository: invalid nhia accredited health maintenance organisation status filter")
	// ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationHMOIDFilter reports an invalid NHIA accredited HMO ID filter.
	ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationHMOIDFilter = errors.New("repository: invalid nhia accredited health maintenance organisation hmo id filter")
	// ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationSearch reports an invalid NHIA accredited HMO search filter.
	ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationSearch = errors.New("repository: invalid nhia accredited health maintenance organisation search")
	// ErrInvalidMedicalLaboratoryAccreditationQuery reports an invalid medical-laboratory-accreditation query.
	ErrInvalidMedicalLaboratoryAccreditationQuery = errors.New("repository: invalid medical laboratory accreditation query")
	// ErrInvalidMedicalLaboratoryAccreditationStateFilter reports an unknown medical-laboratory-accreditation state filter.
	ErrInvalidMedicalLaboratoryAccreditationStateFilter = errors.New("repository: invalid medical laboratory accreditation state filter")
	// ErrInvalidMedicalLaboratoryAccreditationStatusFilter reports an unsupported medical-laboratory-accreditation status filter.
	ErrInvalidMedicalLaboratoryAccreditationStatusFilter = errors.New("repository: invalid medical laboratory accreditation status filter")
	// ErrInvalidMedicalLaboratoryAccreditationSearch reports an invalid medical-laboratory-accreditation search filter.
	ErrInvalidMedicalLaboratoryAccreditationSearch = errors.New("repository: invalid medical laboratory accreditation search")
	// ErrInvalidHealthFacilityQuery reports an invalid health-facility query.
	ErrInvalidHealthFacilityQuery = errors.New("repository: invalid health facility query")
	// ErrInvalidHealthFacilityStateFilter reports an unknown health-facility state filter.
	ErrInvalidHealthFacilityStateFilter = errors.New("repository: invalid health facility state filter")
	// ErrInvalidHealthFacilityLGAFilter reports an unknown health-facility LGA filter.
	ErrInvalidHealthFacilityLGAFilter = errors.New("repository: invalid health facility lga filter")
	// ErrInvalidHealthFacilityStateLGA reports an invalid health-facility state/LGA combination.
	ErrInvalidHealthFacilityStateLGA = errors.New("repository: invalid health facility state lga relationship")
	// ErrInvalidHealthFacilityType reports an unsupported facility type filter.
	ErrInvalidHealthFacilityType = errors.New("repository: invalid health facility type")
	// ErrInvalidHealthFacilityLevel reports an unsupported facility level filter.
	ErrInvalidHealthFacilityLevel = errors.New("repository: invalid health facility level")
	// ErrInvalidHealthFacilityOwnership reports an unsupported ownership filter.
	ErrInvalidHealthFacilityOwnership = errors.New("repository: invalid health facility ownership")
	// ErrInvalidHealthFacilitySearch reports an invalid search filter.
	ErrInvalidHealthFacilitySearch = errors.New("repository: invalid health facility search")
	// ErrInvalidPrimaryAndSecondarySchoolQuery reports that a primary/secondary school query is invalid.
	ErrInvalidPrimaryAndSecondarySchoolQuery = errors.New("repository: invalid primary and secondary school query")
	// ErrPaymentServiceProviderNotFound reports that a requested payment service provider is not present in the dataset.
	ErrPaymentServiceProviderNotFound = errors.New("repository: payment service provider not found")
	// ErrInternationalMoneyTransferOperatorNotFound reports that a requested IMTO is not present in the dataset.
	ErrInternationalMoneyTransferOperatorNotFound = errors.New("repository: international money transfer operator not found")
	// ErrCurrencyNotFound reports that a requested currency is not present in the dataset.
	ErrCurrencyNotFound = errors.New("repository: currency not found")
	// ErrCommercialBankNotFound reports that a requested commercial bank is not present in the dataset.
	ErrCommercialBankNotFound                  = errors.New("repository: commercial bank not found")
	ErrMicrofinanceBankNotFound                = errors.New("repository: microfinance bank not found")
	ErrNonInterestFinancialInstitutionNotFound = errors.New("repository: non-interest financial institution not found")
	ErrMerchantBankNotFound                    = errors.New("repository: merchant bank not found")
	ErrPaymentServiceBankNotFound              = errors.New("repository: payment service bank not found")
	ErrFinancialHoldingCompanyNotFound         = errors.New("repository: financial holding company not found")
	ErrDevelopmentFinanceInstitutionNotFound   = errors.New("repository: development finance institution not found")
	ErrPrimaryMortgageInstitutionNotFound      = errors.New("repository: primary mortgage institution not found")
	// ErrInvalidCurrencyCountryAreaID reports that a currency country/area filter is invalid or unknown.
	ErrInvalidCurrencyCountryAreaID = errors.New("repository: invalid currency country area id")
)
