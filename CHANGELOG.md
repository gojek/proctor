# Changelog
All notable changes to this project will be documented in this file.

## [WIP: v0.2.0] - 2018-10-11
### Added
  - Configurable kubernetes job retries (on failure)
### Modified
  - Changelog format
  - CI golang version. Reason: [kubernetes client-go issue](https://github.com/kubernetes/client-go/issues/449)
### Removed
  - Redundant typing of `proc` for  listing, discribing and executing procs

## v0.1.0 - 2018-10-08
### Added
  - Add Authentication Headers: Email-Id and Access-Token
  - Add Job success and failure metrics

[WIP: v0.2.0]: https://github.com/gojektech/proctor/compare/v0.1.0...v0.2.0
