package file

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

var expectedTimeZoneCount = 312
var expectedTimeZoneRelationshipCount = 422
var expectedTimeZoneUniqueMappedCountryAreaCount = 246

var expectedTimeZoneForwardHist = map[int]int{
	0:  1,
	1:  277,
	2:  15,
	3:  7,
	4:  2,
	5:  4,
	6:  1,
	8:  1,
	10: 2,
	12: 1,
	20: 1,
}

var expectedTimeZoneReverseHist = map[int]int{
	1:  213,
	2:  16,
	3:  5,
	4:  3,
	7:  1,
	11: 1,
	12: 2,
	13: 1,
	16: 1,
	23: 1,
	27: 1,
	29: 1,
}

var expectedTimeZoneZeroZoneCountryAreaIDs = map[string]struct{}{
	"bv": {},
	"hm": {},
}

var expectedTimeZoneAnchors = map[string][]string{
	"Africa/Abidjan":       {"bf", "ci", "gh", "gm", "gn", "is", "ml", "mr", "sh", "sl", "sn", "tg"},
	"Africa/Johannesburg":  {"ls", "sz", "za"},
	"Africa/Lagos":         {"ao", "bj", "cd", "cf", "cg", "cm", "ga", "gq", "ne", "ng"},
	"Africa/Maputo":        {"bi", "bw", "cd", "mw", "mz", "rw", "zm", "zw"},
	"Africa/Nairobi":       {"dj", "er", "et", "ke", "km", "mg", "so", "tz", "ug", "yt"},
	"America/Panama":       {"ca", "ky", "pa"},
	"America/Phoenix":      {"ca", "us"},
	"America/Puerto_Rico":  {"ag", "ai", "aw", "bl", "bq", "ca", "cw", "dm", "gd", "gp", "kn", "lc", "mf", "ms", "pr", "sx", "tt", "vc", "vg", "vi"},
	"America/Toronto":      {"bs", "ca"},
	"Asia/Bangkok":         {"cx", "kh", "la", "th", "vn"},
	"Asia/Dubai":           {"ae", "om", "re", "sc", "tf"},
	"Asia/Gaza":            {"ps"},
	"Asia/Hebron":          {"ps"},
	"Asia/Hong_Kong":       {"hk"},
	"Asia/Kuching":         {"bn", "my"},
	"Asia/Macau":           {"mo"},
	"Asia/Qatar":           {"bh", "qa"},
	"Asia/Riyadh":          {"aq", "kw", "sa", "ye"},
	"Asia/Singapore":       {"aq", "my", "sg"},
	"Asia/Taipei":          {},
	"Asia/Tokyo":           {"au", "jp"},
	"Asia/Yangon":          {"cc", "mm"},
	"Atlantic/Faroe":       {"fo"},
	"Europe/Belgrade":      {"ba", "hr", "me", "mk", "rs", "si"},
	"Europe/Berlin":        {"de", "dk", "no", "se", "sj"},
	"Europe/Brussels":      {"be", "lu", "nl"},
	"Europe/Helsinki":      {"ax", "fi"},
	"Europe/London":        {"gb", "gg", "im", "je"},
	"Europe/Paris":         {"fr", "mc"},
	"Europe/Prague":        {"cz", "sk"},
	"Europe/Rome":          {"it", "sm", "va"},
	"Europe/Simferopol":    {"ru", "ua"},
	"Europe/Zurich":        {"ch", "de", "li"},
	"Indian/Maldives":      {"mv", "tf"},
	"Pacific/Auckland":     {"aq", "nz"},
	"Pacific/Guadalcanal":  {"fm", "sb"},
	"Pacific/Guam":         {"gu", "mp"},
	"Pacific/Pago_Pago":    {"as", "um"},
	"Pacific/Port_Moresby": {"aq", "fm", "pg"},
	"Pacific/Tarawa":       {"ki", "mh", "tv", "um", "wf"},
}

var expectedTimeZoneMultiCountryZoneIDs = []string{
	"Africa/Abidjan",
	"Africa/Johannesburg",
	"Africa/Lagos",
	"Africa/Maputo",
	"Africa/Nairobi",
	"America/Panama",
	"America/Phoenix",
	"America/Puerto_Rico",
	"America/Toronto",
	"Asia/Bangkok",
	"Asia/Dubai",
	"Asia/Kuching",
	"Asia/Qatar",
	"Asia/Riyadh",
	"Asia/Singapore",
	"Asia/Tokyo",
	"Asia/Yangon",
	"Europe/Belgrade",
	"Europe/Berlin",
	"Europe/Brussels",
	"Europe/Helsinki",
	"Europe/London",
	"Europe/Paris",
	"Europe/Prague",
	"Europe/Rome",
	"Europe/Simferopol",
	"Europe/Zurich",
	"Indian/Maldives",
	"Pacific/Auckland",
	"Pacific/Guadalcanal",
	"Pacific/Guam",
	"Pacific/Pago_Pago",
	"Pacific/Port_Moresby",
	"Pacific/Tarawa",
}

// ListTimeZones returns canonical IANA time zones, optionally filtered by country or area.
func (r *GeographyFileRepository) ListTimeZones(ctx context.Context, filter interfaces.TimeZoneFilter) ([]models.TimeZone, error) {
	timeZones, countries, err := r.loadTimeZoneData(ctx)
	if err != nil {
		return nil, err
	}
	if err := validateTimeZones(timeZones, countries); err != nil {
		return nil, err
	}

	countryAreaID := strings.TrimSpace(filter.CountryAreaID)
	if countryAreaID == "" {
		return cloneTimeZoneList(timeZones), nil
	}
	if !countryOrAreaIDPattern.MatchString(countryAreaID) {
		return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}

	filtered := make([]models.TimeZone, 0)
	for _, timeZone := range timeZones {
		if containsCountryAreaID(timeZone.CountryAreaIDs, countryAreaID) {
			filtered = append(filtered, timeZone)
		}
	}
	return cloneTimeZoneList(filtered), nil
}

// GetTimeZone returns a single time zone using its exact canonical IANA identifier.
func (r *GeographyFileRepository) GetTimeZone(ctx context.Context, timeZoneID string) (models.TimeZone, error) {
	timeZoneID = strings.TrimSpace(timeZoneID)
	if !isCanonicalTimeZoneID(timeZoneID) {
		return models.TimeZone{}, fmt.Errorf("%w", interfaces.ErrTimeZoneNotFound)
	}

	timeZones, countries, err := r.loadTimeZoneData(ctx)
	if err != nil {
		return models.TimeZone{}, err
	}
	if err := validateTimeZones(timeZones, countries); err != nil {
		return models.TimeZone{}, err
	}

	for _, timeZone := range timeZones {
		if timeZone.ID == timeZoneID {
			return cloneTimeZone(timeZone), nil
		}
	}

	return models.TimeZone{}, fmt.Errorf("%w", interfaces.ErrTimeZoneNotFound)
}

func (r *GeographyFileRepository) loadTimeZonesOnly(ctx context.Context) ([]models.TimeZone, error) {
	if r == nil || r.jsonRepository == nil {
		return nil, fmt.Errorf("%w", interfaces.ErrDatasetFileUnavailable)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var timeZones []models.TimeZone
	if err := r.jsonRepository.Decode(ctx, r.timeZonesPath, &timeZones); err != nil {
		return nil, translateGeographyLoadError(err)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if timeZones == nil || len(timeZones) == 0 {
		return nil, fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	return timeZones, nil
}

func (r *GeographyFileRepository) loadTimeZoneData(ctx context.Context) ([]models.TimeZone, []models.CountryOrArea, error) {
	timeZones, err := r.loadTimeZonesOnly(ctx)
	if err != nil {
		return nil, nil, err
	}
	countries, err := r.loadCountriesAndAreasOnly(ctx)
	if err != nil {
		return nil, nil, err
	}
	if err := validateCountryOrAreas(countries); err != nil {
		return nil, nil, err
	}
	return timeZones, countries, nil
}

func validateTimeZones(timeZones []models.TimeZone, countries []models.CountryOrArea) error {
	if len(timeZones) != expectedTimeZoneCount {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}

	countryIDs := make(map[string]struct{}, len(countries))
	for _, country := range countries {
		countryIDs[country.ID] = struct{}{}
	}

	seenIDs := make(map[string]struct{}, len(timeZones))
	countryZoneCounts := make(map[string]int, len(countries))
	for _, country := range countries {
		countryZoneCounts[country.ID] = 0
	}

	totalRelationships := 0
	forwardHist := make(map[int]int)
	observedMultiCountryZones := make(map[string][]string)
	zeroMappingZone := ""
	zeroMappingCount := 0

	var prevID string
	for i, timeZone := range timeZones {
		if timeZone.ID == "" || timeZone.CountryAreaIDs == nil {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if !isCanonicalTimeZoneID(timeZone.ID) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if strings.HasPrefix(timeZone.ID, "Etc/") || strings.HasPrefix(timeZone.ID, "posix/") || strings.HasPrefix(timeZone.ID, "right/") || timeZone.ID == "Factory" {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := seenIDs[timeZone.ID]; ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if i > 0 && strings.Compare(prevID, timeZone.ID) > 0 {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if !reflectDeepEqualStrings(timeZone.CountryAreaIDs, sortedUniqueStrings(timeZone.CountryAreaIDs)) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		for _, countryID := range timeZone.CountryAreaIDs {
			if _, ok := countryIDs[countryID]; !ok {
				return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
			}
			countryZoneCounts[countryID]++
		}
		switch len(timeZone.CountryAreaIDs) {
		case 0:
			zeroMappingZone = timeZone.ID
			zeroMappingCount++
		default:
			if len(timeZone.CountryAreaIDs) > 1 {
				observedMultiCountryZones[timeZone.ID] = append([]string(nil), timeZone.CountryAreaIDs...)
			}
		}
		forwardHist[len(timeZone.CountryAreaIDs)]++
		totalRelationships += len(timeZone.CountryAreaIDs)
		seenIDs[timeZone.ID] = struct{}{}
		prevID = timeZone.ID
	}

	if zeroMappingZone != "Asia/Taipei" || zeroMappingCount != 1 {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	for countryAreaID := range expectedTimeZoneZeroZoneCountryAreaIDs {
		if count := countryZoneCounts[countryAreaID]; count != 0 {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
	}
	if totalRelationships != expectedTimeZoneRelationshipCount {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	if len(seenIDs) != expectedTimeZoneCount {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	uniqueMapped := 0
	reverseHist := make(map[int]int)
	for _, count := range countryZoneCounts {
		if count > 0 {
			reverseHist[count]++
			uniqueMapped++
		}
	}
	if uniqueMapped != expectedTimeZoneUniqueMappedCountryAreaCount {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	if !equalIntHistograms(forwardHist, expectedTimeZoneForwardHist) {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	if !equalIntHistograms(reverseHist, expectedTimeZoneReverseHist) {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	if _, ok := countryZoneCounts["bv"]; !ok || countryZoneCounts["bv"] != 0 {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	if _, ok := countryZoneCounts["hm"]; !ok || countryZoneCounts["hm"] != 0 {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	if len(observedMultiCountryZones) != len(expectedTimeZoneMultiCountryZoneIDs) {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	for zoneID, want := range expectedTimeZoneAnchors {
		var got []string
		for _, timeZone := range timeZones {
			if timeZone.ID == zoneID {
				got = timeZone.CountryAreaIDs
				break
			}
		}
		if !reflectDeepEqualStrings(got, want) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
	}

	expectedMulti := make(map[string]struct{}, len(expectedTimeZoneMultiCountryZoneIDs))
	for _, zoneID := range expectedTimeZoneMultiCountryZoneIDs {
		expectedMulti[zoneID] = struct{}{}
	}
	if len(observedMultiCountryZones) != len(expectedMulti) {
		return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
	}
	for zoneID, got := range observedMultiCountryZones {
		want, ok := expectedTimeZoneAnchors[zoneID]
		if !ok || !reflectDeepEqualStrings(got, want) {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
		if _, ok := expectedMulti[zoneID]; !ok {
			return fmt.Errorf("%w", interfaces.ErrInvalidDatasetFile)
		}
	}

	return nil
}

func isCanonicalTimeZoneID(id string) bool {
	if id == "" || strings.HasPrefix(id, "/") || strings.HasSuffix(id, "/") {
		return false
	}
	if strings.ContainsAny(id, " \t\n\r\\%?#") {
		return false
	}
	if !timeZoneIDPattern.MatchString(id) {
		return false
	}
	parts := strings.Split(id, "/")
	if len(parts) < 2 {
		return false
	}
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return false
		}
	}
	return true
}

func containsCountryAreaID(ids []string, countryAreaID string) bool {
	for _, id := range ids {
		if id == countryAreaID {
			return true
		}
	}
	return false
}

func cloneTimeZoneList(timeZones []models.TimeZone) []models.TimeZone {
	if len(timeZones) == 0 {
		return make([]models.TimeZone, 0)
	}
	cloned := make([]models.TimeZone, len(timeZones))
	for i, timeZone := range timeZones {
		cloned[i] = cloneTimeZone(timeZone)
	}
	return cloned
}

func cloneTimeZone(timeZone models.TimeZone) models.TimeZone {
	timeZone.CountryAreaIDs = cloneStringSlice(timeZone.CountryAreaIDs)
	return timeZone
}

func cloneStringSlice(values []string) []string {
	if len(values) == 0 {
		return make([]string, 0)
	}
	cloned := make([]string, len(values))
	copy(cloned, values)
	return cloned
}

func sortedUniqueStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	out := append([]string(nil), values...)
	sort.Strings(out)
	dedup := out[:1]
	for i := 1; i < len(out); i++ {
		if out[i] != out[i-1] {
			dedup = append(dedup, out[i])
		}
	}
	return dedup
}

func reflectDeepEqualStrings(a, b []string) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func equalIntHistograms(got, want map[int]int) bool {
	if len(got) != len(want) {
		return false
	}
	for key, wantValue := range want {
		if got[key] != wantValue {
			return false
		}
	}
	for key := range got {
		if _, ok := want[key]; !ok {
			return false
		}
	}
	return true
}

func findCountryOrAreaByID(countries []models.CountryOrArea, countryAreaID string) (models.CountryOrArea, bool) {
	for _, country := range countries {
		if country.ID == countryAreaID {
			return country, true
		}
	}
	return models.CountryOrArea{}, false
}
