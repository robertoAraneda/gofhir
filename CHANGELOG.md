# Changelog

## [1.5.0](https://github.com/robertoAraneda/gofhir/compare/v1.4.1...v1.5.0) (2025-12-29)


### Features

* **validator:** add format validation for code, id, oid, uuid primitives ([75b301c](https://github.com/robertoAraneda/gofhir/commit/75b301c922b417b9a42a9cea5d9dddee2c43437c))


### Bug Fixes

* **validator:** detect primitive type mismatches and missing required fields ([8cd7007](https://github.com/robertoAraneda/gofhir/commit/8cd7007f2cb47050ba185c281f5d6a89d850749b))

## [1.4.1](https://github.com/robertoAraneda/gofhir/compare/v1.4.0...v1.4.1) (2025-12-27)


### Bug Fixes

* **fhirpath:** support Resource and DomainResource base type inheritance ([848d840](https://github.com/robertoAraneda/gofhir/commit/848d8409ef29995c5c67d733a209bf52305f10ec))

## [1.4.0](https://github.com/robertoAraneda/gofhir/compare/v1.3.0...v1.4.0) (2025-12-19)


### Features

* **codegen:** add UnmarshalJSON for polymorphic resource fields ([6eb1155](https://github.com/robertoAraneda/gofhir/commit/6eb1155ae48a60f6eab1fbd40b73eb342a366295))

## [1.3.0](https://github.com/robertoAraneda/gofhir/compare/v1.2.0...v1.3.0) (2025-12-19)


### Features

* **validator:** add contained resource validation support ([de6c7eb](https://github.com/robertoAraneda/gofhir/commit/de6c7eb9930a8719696d4b24ae623fb3870c5bd7))

## [1.2.0](https://github.com/robertoAraneda/gofhir/compare/v1.1.0...v1.2.0) (2025-12-17)


### Features

* **fhirpath:** add iif lazy evaluation, %context variable, and backtick identifiers ([9c3350e](https://github.com/robertoAraneda/gofhir/commit/9c3350e525bdefa3064465437ced56b4f5f5e025))
* **fhirpath:** add UCUM normalization for quantity comparisons ([ecb031b](https://github.com/robertoAraneda/gofhir/commit/ecb031bf5b4129f18c6e3ad20a713556004b4800))
* **fhirpath:** add unit parameter support to convertsToQuantity() ([1bd2a9c](https://github.com/robertoAraneda/gofhir/commit/1bd2a9c5a30765a16057e2187e313999a9e55b74))

## [1.1.0](https://github.com/robertoAraneda/gofhir/compare/v1.0.0...v1.1.0) (2025-12-17)


### Features

* **fhirpath:** add polymorphic element resolution and ofType function ([dc92417](https://github.com/robertoAraneda/gofhir/commit/dc92417163fefe0489d7d59e208ec964baac09de))

## 1.0.0 (2025-12-16)


### Features

* add automatic generation of Functional Options and Fluent Builders ([f033782](https://github.com/robertoAraneda/gofhir/commit/f033782e02606a3daf6ec17982ffb04c06d9561c))
* add Bundle validation and FHIRPath %resource variable ([001fdba](https://github.com/robertoAraneda/gofhir/commit/001fdba0ecf7c42bfc7587a8f194ba2015e45290))
* add clinical helpers and backbone elements for all FHIR versions ([e779da3](https://github.com/robertoAraneda/gofhir/commit/e779da3b90f7d7b1f68173b276b790969dd11000))
* add complete FHIRPath 2.0.0 implementation with security features ([b99fb82](https://github.com/robertoAraneda/gofhir/commit/b99fb82b4c960a2f5b6d4dc08e6cd2a75dba98b1))
* add deep validation for extension values ([20c62ed](https://github.com/robertoAraneda/gofhir/commit/20c62ed40e2bd0dc4a33f86bf19d92ed879621d8))
* add FHIR extension validator and examples for references/extensions ([1da51ab](https://github.com/robertoAraneda/gofhir/commit/1da51ab71c620f3fc22fbc911f5bbe74db060e5f))
* add FHIR R4B and R5 type generation ([d735c62](https://github.com/robertoAraneda/gofhir/commit/d735c62c0aba5ab4b5293a6996d0a3a525cafe5c))
* add FHIR reference validator with comprehensive format support ([657da3c](https://github.com/robertoAraneda/gofhir/commit/657da3c6dd6d823449ea0329e235a09ed0a954ab))
* add terminology validation with embedded ValueSets ([ee5eb33](https://github.com/robertoAraneda/gofhir/commit/ee5eb33b2fb05edd5d750d723319f81a6d253035))
* add UCUM normalization for FHIR quantity search parameters ([78a76df](https://github.com/robertoAraneda/gofhir/commit/78a76df4d6f183c74d4832dfaa6a849f9eb5b111))
* **codegen:** add custom types for required code bindings and primitive extensions ([0b47134](https://github.com/robertoAraneda/gofhir/commit/0b47134b6477f9304c9beed965d11e482c14756c))
* implement global ele-1 constraint validation ([255c0e1](https://github.com/robertoAraneda/gofhir/commit/255c0e15b39c115e262794ce2fcb7504f7853fec))
* initial project setup (Sprint 0) ([f9bf78b](https://github.com/robertoAraneda/gofhir/commit/f9bf78b6edcbbbce427749dd19ac4f0e9b85eca2))
* initial release ([82ec28c](https://github.com/robertoAraneda/gofhir/commit/82ec28c30a38afb26bbf7b2503945573606da517))
* **sprint-1:** implement pkg/common and codegen parser/analyzer ([4cd97fa](https://github.com/robertoAraneda/gofhir/commit/4cd97fa8d593f6ff55416cb3c22e2c20a80848da))
* **sprint-2:** implement code generation from FHIR StructureDefinitions ([458501f](https://github.com/robertoAraneda/gofhir/commit/458501fa5eb615d397f6d6e57565549f987a211a))


### Bug Fixes

* enhance validator for FHIR primitive types ([78a76df](https://github.com/robertoAraneda/gofhir/commit/78a76df4d6f183c74d4832dfaa6a849f9eb5b111))
* prevent pointer-to-interface anti-pattern in generated code ([4b5c04b](https://github.com/robertoAraneda/gofhir/commit/4b5c04bcb0161ee1e5bb20d184abb8cfe047cca4))
* remove unnecessary global factory registry test ([e165df0](https://github.com/robertoAraneda/gofhir/commit/e165df0673c821f5c47b439ca7960d45f566353d))
* resolve all golangci-lint issues for CI compliance ([f7fb591](https://github.com/robertoAraneda/gofhir/commit/f7fb5917125836117eec1130cd173e8553ab7711))
* resolve golangci-lint errors ([a74fe19](https://github.com/robertoAraneda/gofhir/commit/a74fe19eb0048e5293d93842daee921b686ee72a))
* resolve golangci-lint errors for CI compliance ([38d91ec](https://github.com/robertoAraneda/gofhir/commit/38d91ecbac85d489c82d329b9d70f916a0479d56))
* resolve lint issues ([af589d0](https://github.com/robertoAraneda/gofhir/commit/af589d0214d02e8f250dd2b9b2e479f8cd5b8253))
