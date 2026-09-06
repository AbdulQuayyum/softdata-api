# Microfinance Bank Enrichment Acquisition Audit

The complete 790-record machine-readable audit is maintained at
`datasets/metadata/finance/microfinance_banks_enrichment.json`. It preserves
the deterministic dataset position, search completion state, website
classification, evidence URLs, logo source, hashes and unresolved findings.

Internal checkpoints use 25-record batches. Positions 1-790 are complete and
the next resume position is 791. Positions 207-400 completed the deeper
reprocessing pass, and positions 401-790 completed the final forward pass.
Current coverage is:

```text
verified websites: 74
unresolved websites in completed batches: 716
accepted logos: 330
unresolved logos in completed batches: 460
```

The 2026-09-06 supplied-source reconciliation is recorded in
`SOURCE_RECONCILIATION.md`. It reviewed Nigerian Bank Logos, the Woven Finance
CDN, the Pariola Nigerian Bank Logos repository and NigerianBanks.xyz. No new
MFB website or logo was promoted from those four live sources: catalogue/API
retrieval was not reproducible
for identity, Woven numeric filenames were not mapped without an authoritative
MFB code register, and NigerianBanks.xyz was unavailable from the research
runtime. Permission was not used as a rejection criterion. `NPF MFB` remains
an active roster record; no NPF source mark was accepted as Nigeria Police
Mortgage Bank branding.

The supplied `mfb-logo-import` bundle provides 310 manifest matches from the
approved `Nigerian-Bank-Logos/ng-bank-logos` archive commit
`a56f7857b1ccca69784cd86a8080531fd3c7e6a8`. Seven matched IDs already had
stronger existing first-party logos and were preserved; 303 new PNG assets
were imported unchanged. The bundle also identifies 448 source PNG files,
138 unused source files and 480 unmatched roster records. Its manifest is the
authoritative mapping for this import.

The corrected `mfb-unused-source-review` bundle supplied all 138 PNG bytes and
the source catalogue. All 138 entries were reviewed. Fourteen matched active
roster identities; the existing first-party `b-c-kash-microfinance-bank` asset
was preserved and thirteen new PNGs were imported. The review classifies the
remaining 124 entries as rejected no-roster matches or ambiguous manual
reviews in `datasets/metadata/finance/microfinance_banks_unused_source_reconciliation.json`.

All roster records have a completed research status. The active roster, IDs,
names, ordering and CBN/NDIC reconciliation
are unchanged. Website findings are recorded only for exact first-party or
explicit first-party parent profiles; app listings, directories and similarly
named institutions are rejected. A deterministic 25-record quality sample of
positions 207-400 found additional identity-backed websites for EdFin, GTI,
GoldMan and the KKU-to-Turbo identity. All positions 207-400 were then
reprocessed with the deeper search method; no additional logo passed
original-asset verification. The forward pass added Koboweb, Koins and Kolisa
websites but no new logo passed original-asset verification. The final
40-record negative-result sample is recorded in the enrichment manifest; it did
not trigger another negative-result pass. Prior logo assets remain unchanged.

Institution-owned marks remain trademarks of their respective owners. The
project does not require individual permission or an explicit redistribution
licence before technical acceptance, and does not claim that accepted marks
are CC BY 4.0 or owned by SoftData.
