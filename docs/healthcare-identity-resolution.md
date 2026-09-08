# Health-facility identity review, 9 September 2026

The five identity collisions previously treated as exact duplicates are resolved for publication by quarantining **both observations of each pair** as `exclude_unresolved_identity`. The available evidence does not establish whether either pair describes one site, separate sites, relocation, or source error. No observation is selected as canonical, and no coordinates or classifications are averaged or silently preferred. This decision does not claim that the facilities do not exist.

The [review manifest](../datasets/metadata/healthcare/health_facilities_identity_reviews.json) is the generator input and the regression fixture for all five pairs. It records source positions, stable IDs, every retrieved GRID3 attribute and geometry, normalized names/geography, previous decisions and public IDs, conflicting fields, distance, evidence URLs, retrieval dates, coordinate hypotheses, final decisions/IDs, and canonical-value decisions. Historical decision labels in that manifest describe the superseded output, not current reconciliation decisions.

| Pair | Source positions | Separation (m) | Principal conflicts | Final decision |
|---|---|---:|---|---|
| Awo Primary Health Center, Edo | 7801 / 27202 | 1,141.633 | Clinic vs health centre; ownership subtype; coordinates; NHFR identity/provenance | Exclude both unresolved observations |
| Ibuji Comprehensive Health Center, Ondo | 8152 / 28718 | 628.820 | Public vs unknown ownership; coordinates; NHFR identity/provenance | Exclude both unresolved observations |
| Nkume Health Post, Enugu | 8185 / 31128 | 474.398 | Clinic vs health post; ownership subtype; coordinates; NHFR identity/provenance | Exclude both unresolved observations |
| Aran Orin Comprehensive Health Center, Kwara | 14772 / 35746 | 3,229.159 | Ownership subtype; coordinates; NHFR identity/provenance | Exclude both unresolved observations |
| Kurba Health Clinic, Gombe | 37122 / 46719 | 446.054 | Unknown vs clinic/primary/public/local-government; coordinates; name and coordinate provenance | Exclude both unresolved observations |

All pairs also have different source IDs/OBJECTIDs and ward formatting. The manifest contains the complete differences rather than only this summary.

## Evidence and limits

The [public GRID3 layer](https://services3.arcgis.com/BU6Aadhn6tbBEdyk/ArcGIS/rest/services/GRID3_NGA_health_facilities_v2_0/FeatureServer/0) returned all ten feature records. Their preserved attributes match the original reconciliation. Full record attributes add NHFR UIDs, name/coordinate provenance, geography-disagreement flags, and update dates. No feature cross-references its partner. Most pairs combine an NHFR_2024 name observation with a GRID3_EHEALTH observation; Kurba combines GRID3_2024 with NHFR_2024/GRID3_EHEALTH. All carry the same layer update date, 2024-11-11. Source position and attribute completeness do not establish which observation is older or more accurate.

The [NHFR public Facility Finder](https://hfr.fmohconnect.gov.ng/facilitieslist) and its public search implementation were inspected. Read-only name searches for Awo, Ibuji, Nkume, Aran, and Kurba through its search endpoint timed out after 35 seconds. The separately documented [external NHFR API](https://www.hfr.fmohconnect.gov.ng/developers) requires an API key; an unauthenticated probe returned 401. No credentials or access bypass were used. NHFR was not substituted as the dataset's row-level source.

The [Kwara health insurance network](https://kwaracare.com.ng/our-hospital-network/) identifies PHC Aran Orin under SPHCDA, and a [University of Ilorin publication](https://www.unilorin.edu.ng/wp-content/uploads/2024/10/TRANSFORMING-UNIVERSITY-EDUCATION-IN-THE-21ST-CENTURY-1.pdf) describes work with the community health centre and COBES infrastructure. Neither links the two GRID3 points or explains their 3.23 km separation. The [Ondo government report](https://www.ondobudget.org/materials/SDPIC%20Complete%20Report.pdf), printed page 104, identifies a comprehensive health centre at Ibuji without a coordinate crosswalk. The [Enugu government geography codes](https://enugustate.gov.ng/wp-content/uploads/2025/03/Enugu-State-Local-Government-Harmonised-Budget-Classification-Codes-and-Chart-of-Accounts-V2.pdf) corroborate the Adaba-Nkume ward but do not resolve the sites.

NPHCDA and relevant state/federal sources were reviewed or attempted. No primary site-level correction or crosswalk was obtained for Awo, Nkume, or Kurba. Failed URLs, public-search requests and results are recorded per pair. A retrieved third-party Kurba directory only corroborates the coded name; its activity/licensing claims were not adopted. Ordinary search snippets were not used as final evidence. No first-party facility site establishing the disputed points was found; map imagery was not used as an identity authority.

Distances use the inverse haversine great-circle formula with a mean Earth radius of 6,371,008.8 metres, using WGS84 input degrees without rounding before calculation. This is a spherical geodesic approximation, not a survey-grade ellipsoidal distance. All ten points passed polygon containment against the pinned [geoBoundaries Nigeria ADM0 boundary sourced from GRID3](https://github.com/wmgeolab/geoBoundaries/raw/9469f09/releaseData/gbOpen/NGA/ADM0/geoBoundaries-NGA-ADM0.geojson); boundary ID, attribution, license, hash and method are recorded in the manifest. This check covers the ten disputed points, not a new boundary audit of every public record.

GPS error, entrance/centroid differences, relocation, separate campuses, and source error were considered. No accuracy radius, surveyed footprint, coordinate correction or relocation chronology supports selecting one hypothesis. In particular, the kilometre-separated Aran Orin pair received separate-site review rather than an automatic merge. Country membership and distance alone did not decide identity.

## Arithmetic and public identifiers

Before: `51,022 − 363 geography exclusions − 5 incorrectly labeled exact merges = 50,654`.

After: `51,022 − 0 exact merges − 0 same-facility merges − 363 geography exclusions − 10 unresolved identity exclusions = 50,649`.

All 51,022 source positions still have one decision. All 50,649 public records have retain evidence; there are no merge targets or separately retained members of these five pairs. The five formerly published IDs are withdrawn and now return 404. Other IDs and data values are unchanged. Coverage remains 37 states/FCT, 768 canonical LGAs, and 50,649 coordinate pairs. At page size 50 there are 1,013 pages, with 49 records on the final page.

Public IDs are lowercase ASCII slugs bounded at **255 characters** in the model constant, dataset schema, repository checks, service, HTTP validator, and OpenAPI. The maximum observed length remains 143 and 22 published IDs exceed 128 characters. The longest ID is:

```text
dr-lawrence-henshaw-memorial-specialist-hospital-and-research-centre-cross-river-cross-river-calabar-south-661d4da7-42af-4003-83cc-ef8913116d94
```

The server does not configure a custom `MaxHeaderBytes`; Go's default request-header limit is much larger than a 255-character facility slug. That transport limit is not a substitute for the shared application ID limit. Unsupported characters and 256-character IDs are rejected with JSON 400; syntactically valid unknown long IDs return JSON 404. No current activity, website, or logo property is published.

## Reproducible generator and validation

Run from the repository root:

```sh
python3 tools/generate_health_facilities.py --from-reconciliation . --repo . --retrieved-at 2026-09-08
python3 -m unittest discover -s tools -p test_generate_health_facilities.py -v
```

For a separate output directory, provide its `datasets/geography/states.json` and `lgas.json`, then use `--repo OUTPUT`. The generator writes the dataset, all reconciliation partitions, index, metadata, and schema bounds. It validates input partition hashes/sizes and contiguous source positions before replay. The review manifest is authored evidence input, not a manual edit to generated output.

Exact equality compares every supplied normalized attribute except the ArcGIS storage OBJECTID; stable source identity, coordinates, classifications, ownership, geography and provenance remain in the signature. Name equality alone cannot merge rows. Collisions with conflicting attributes or conflicting observations of one stable source ID require a pinned explicit review; unreviewed conflicts fail generation. Reviewed separate-campus fixtures use meaningful campus disambiguators, with the stable source ID available only when geography cannot distinguish observations. There are no arbitrary numeric suffixes.

Offline replay reconstructs the normalization input from preserved reconciliation attributes and uses the full pinned GRID3 attributes for the ten reviewed observations. Metadata records the input projection hash and review-manifest hash. The original response hash remains provenance: replay does **not** claim to reconstruct the original ArcGIS transport payload or all unpreserved source attributes for unrelated rows. Source-position order preserves the original combined response; it is not an OBJECTID sort or evidence of observation chronology. No full source API dump is added.

Regression tests cover all five pairs, conflicting coordinates/ownership/type/geography, exact equality, stale reviews, separate-campus IDs, and two byte-identical complete regenerations. Go tests cover every published ID, longest ID and length boundaries, exact public model/schema fields, and external/embedded dataset startup. Reconciliation files are neither embedded nor loaded at runtime.
