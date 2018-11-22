# Changelog
All notable changes to this project will be documented in this file.

## [v0.3.0] - 2018-11-22
### Added
- NewRelic Instrumentation
- Connection timeouts from CLI to ProctorD
- Validation for minimum supported version of CLI
- Command to configure CLI
### Modified
- CLI config validations and error messages
### Removed
- `proctor proc ...` commands

## [v0.2.0] - 2018-10-11
### Added
  - Configurable kubernetes job retries (on failure)
  - Meta fields for proc contributor details
### Modified
  - Changelog format
  - CI golang version. Reason: [kubernetes client-go issue](https://github.com/kubernetes/client-go/issues/449)
  - `PROCTOR_URL` config variable to `PROCTOR_HOST`
### Removed
  - Redundant typing of `proc` for  listing, discribing and executing procs

## v0.1.0 - 2018-10-08
### Added
  - Add Authentication Headers: Email-Id and Access-Token
  - Add Job success and failure metrics

[v0.2.0]: https://github.com/gojektech/proctor/compare/v0.1.0...v0.2.0
[v0.3.0]: https://github.com/gojektech/proctor/compare/v0.2.0...v0.3.0
