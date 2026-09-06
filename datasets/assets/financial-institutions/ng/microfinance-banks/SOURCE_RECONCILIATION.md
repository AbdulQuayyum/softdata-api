# Microfinance Logo Source Reconciliation

Audit date: 2026-09-06

This pass evaluated the supplied community/catalogue sources against the
790-record `ng-microfinance-banks` roster. Explicit redistribution permission
is not required by project policy, but identity verification, original-byte
validation and provenance recording remain mandatory. No roster record,
website, or logo asset was changed because this pass did not produce a new
identity-verified MFB asset.

## Current baseline

| Measure | Count |
| --- | ---: |
| Active roster records | 790 |
| Records with `website_url` | 74 |
| Records with `logo_url` | 330 |
| Existing embedded PNG assets | 330 |
| Missing referenced assets | 0 |
| Orphaned assets | 0 |

The public dataset remains limited to the approved CBN/NDIC roster. This
source pass does not claim that all 790 institutions have a website or logo.

## Supplied source results

| Source | Retrieval result | Inventory/reconciliation result | Decision |
| --- | --- | --- | --- |
| [Nigerian Bank Logos](https://nigerianbanklogos.xyz/) | Homepage reachable on 2026-09-06 | Catalogue reports 140 total entries and exposes category/search/API links. The page is a community catalogue; its API/root host returned an HTTP 500 page from this runtime. | No exact active-MFB identity was established. |
| [Nigerian Bank Logos API](https://api.nigerianbanklogos.xyz/) | Root and `/api` returned HTTP 500 HTML from this runtime | No reproducible JSON response, item manifest, or source hash was available. Rights permission was not used as a rejection criterion; unresolved transport and identity evidence were. | Not accepted for this pass. |
| [Woven Finance CDN](https://github.com/wovenfinance/cdn) | Repository and `logos/` listing reachable; `main` head observed as `37d5c3b06f22c828ad152b161bd83c4f6e57369b` | PNG filenames are numeric institution-code candidates, including six-digit names. The source README describes a jsDelivr delivery pattern, but the roster has no approved complete MFB code mapping for these files. | No numeric filename was mapped or imported. |
| [Pariola Nigerian Bank Logos](https://github.com/Pariola-droid/Nigerian-Bank-Logos) | Repository reachable; `src/data/logos.ts` and `public/library/` are present | The manifest uses title/category/route/website/ticker metadata and includes commercial, holding and mortgage marks. Retrieved content did not establish a current exact MFB match with rights and identity evidence. | No asset imported. |
| [NigerianBanks.xyz](https://nigerianbanks.xyz/) | Site could not be fetched by the research runtime | No API, JSON, bundle, sitemap, or asset inventory could be verified. | Unavailable; no use. |

## Approved curated repository retry

The project-approved source
[Nigerian-Bank-Logos/ng-bank-logos](https://github.com/Nigerian-Bank-Logos/ng-bank-logos)
was requested directly from the local environment on 2026-09-06. Both the
required shallow clone and the normal clone were attempted in temporary
directories outside the repository. Both failed before receiving repository
data:

```text
fatal: unable to access 'https://github.com/Nigerian-Bank-Logos/ng-bank-logos.git/':
Could not resolve host: github.com
```

The supplied `mfb-logo-import-bundle.zip` subsequently provided the source
inventory and `manifest.json` for this repository. It reports 310 unique
manifest entries, 448 source PNG files, 138 unused source files and 480
unmatched roster records, pinned to commit
`a56f7857b1ccca69784cd86a8080531fd3c7e6a8`. All 310 manifest PNGs passed
SHA-256, byte-size, PNG-signature and MIME checks. Seven entries overlapped
existing first-party assets and were preserved; 303 assets were imported
unchanged. The later corrected `mfb-unused-source-review-bundle.zip` was
reviewed separately below; no entry from its unmatched roster was used.

Repository license notices were recorded only as repository-level facts. Under
the revised project policy, a repository or individual redistribution licence
is not required for technical acceptance when identity is established. The
[Pariola MIT license](https://raw.githubusercontent.com/Pariola-droid/Nigerian-Bank-Logos/main/LICENSE.md)
and the [Woven repository](https://github.com/wovenfinance/cdn) license do not
by themselves prove that a file represents the current MFB subsidiary.
Institutional trademarks remain owned by their respective institutions.

## Identity checks

The roster contains `NPF MFB` (`npf-mfb`). No source result was accepted as
evidence for that record. In particular, an NPF-branded mark must not be
treated as Nigeria Police Mortgage Bank branding, and no source entry was
used to transfer a parent, commercial-bank, mortgage-bank, fintech, or other
category logo into the MFB roster.

The first bundle supplied 310 exact or normalized manifest matches. The seven
overlapping records retained their existing first-party assets, while 303
curated repository assets were added. The corrected second-stage bundle then
provided 14 additional source-to-roster matches, preserving one existing
first-party asset and adding 13 curated assets. No fuzzy match outside the
review evidence was accepted, and no unmatched roster record was imported.

The corrected review bundle supplied all 138 PNG bytes, `review-manifest.json`,
`source-bank.json` and `unmatched-roster.json`. All entries were reviewed
individually and are enumerated in
`metadata/finance/microfinance_banks_unused_source_reconciliation.json`.
Fourteen source entries matched active roster identities; the existing BC Kash
asset was preserved and thirteen new assets were imported unchanged. The
remaining 124 entries were not imported because their source identity did not
match an active roster record or remained ambiguous.

## Final ambiguous-pair reconciliation

The six former ambiguous entries were reviewed against the visible PNG marks,
the source catalogue names and codes, and CBN/NDIC or first-party records. All
six are rejected as distinct identities and remain unimported:

| Source asset | Proposed roster record | Decision | Evidence |
| --- | --- | --- | --- |
| `Arise Microfinance Bank.png` | `aris-microfinance-bank` | `rejected_distinct_identity` | CBN lists Aris at 29 Zik Avenue, Enugu and Arise at 99 Ogudu Road, Ojota, Lagos separately. |
| `Crust Microfinance Bank.png` | `crest-mfb` | `rejected_distinct_identity` | CBN and NDIC identify Crest at 43 Obafemi Awolowo Way, Oke-Bola, Ibadan; no Crust-to-Crest identity evidence exists. |
| `Microbiz MFB.png` | `microvis-mfb` | `rejected_distinct_identity` | The source code is 090587 and Microbiz's first-party API identifies Microbiz; Microvis's first-party page identifies a separate Kaduna institution. |
| `RSU MFB.png` | `orsu-mfb` | `rejected_distinct_identity` | The source code is 090535; CBN identifies ORSU at Orsuihiteukwa, Orsu, Imo State. |
| `Rank MFB.png` | `rano-mfb` | `rejected_distinct_identity` | NDIC lists Rank in Ibeju Lekki-Epe, Lagos and Rano in Rano Town, Kano separately. |
| `STANFORD MFB.png` | `standard-mfb` | `rejected_distinct_identity` | CBN lists Standard in Jimeta-Yola, Adamawa and Stanford at 142 Abak Road, Uyo, Akwa Ibom separately. |

The complete per-file evidence, source hashes and visible-mark text are in
`datasets/metadata/finance/microfinance_banks_unused_source_reconciliation.json`.

## Acceptance rule

A future promotion must record the exact active roster ID, canonical name,
source page, direct asset URL, retrieval date, source and generated hashes,
actual MIME type and dimensions, and a provenance basis. Woven files also
require an independently verified authoritative MFB code mapping; the
six-digit filename alone is not evidence of identity.

The required attribution statement for every accepted mark is: “Logo and
trademark rights remain with the respective institution. The asset is
reproduced for institutional identification.” The existing first-party assets
and attribution records remain unchanged.
