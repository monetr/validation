# Changelog

## [1.2.0](https://github.com/monetr/validation/compare/v1.1.0...v1.2.0) (2026-06-13)


### Features

* Adding an AllOf helper for more complex unions ([1f2aa96](https://github.com/monetr/validation/commit/1f2aa960659106ac25ac59d5d4f7a509cfc39ca5))
* Adding type constraints like isString or isBool ([c3fd3de](https://github.com/monetr/validation/commit/c3fd3de7c3ad7975ba760f9db840843f12d3cfbe))


### Build Automation

* **deps:** update actions/checkout action to v6.0.3 ([#16](https://github.com/monetr/validation/issues/16)) ([0c1a65b](https://github.com/monetr/validation/commit/0c1a65b06ffeb20e8dbe8ac624b3ace01772ad4b))
* **deps:** update github/codeql-action action to v4.36.1 ([#15](https://github.com/monetr/validation/issues/15)) ([97faa8c](https://github.com/monetr/validation/commit/97faa8c76c941db3c39125f68031b5a28f603894))
* **deps:** update github/codeql-action action to v4.36.2 ([#18](https://github.com/monetr/validation/issues/18)) ([cb75153](https://github.com/monetr/validation/commit/cb751539708003c37fbc23a1f221b00bbebe6b54))

## [1.1.0](https://github.com/monetr/validation/compare/v1.0.6...v1.1.0) (2026-05-30)


### Features

* Adding `IsPrintableUnicode` ([1aafdf4](https://github.com/monetr/validation/commit/1aafdf4ce3b29b06642ac0dd3d3ab1ec1d23e032))
* Adding more helpers, Eq and NotEq ([ae9a45d](https://github.com/monetr/validation/commit/ae9a45de2bd944a0013021c741ed65d25a08f592))
* Adding union/oneof validator ([#14](https://github.com/monetr/validation/issues/14)) ([e0327ea](https://github.com/monetr/validation/commit/e0327ea92a870cec8120ae0a9b7b166dcdaaf890))
* More helpers, between, equal field, not equal field etc ([8d864ed](https://github.com/monetr/validation/commit/8d864ed681aaf5fa65f3efa57c20fb6172b21469))


### Bug Fixes

* Fixed E164 bug ([8ece6da](https://github.com/monetr/validation/commit/8ece6da56d8f8ac56f709ad42d541d9e0b1b9872))
* Fixed missing currency ([4d84cf7](https://github.com/monetr/validation/commit/4d84cf7997922b7bf7d5a44e7d80037997289cb3))
* Fixed nil pointer dereference issue ([bbbff68](https://github.com/monetr/validation/commit/bbbff68600d12ebd297d385573668d6d2a05d473))
* Fixing error template panic ([8201e78](https://github.com/monetr/validation/commit/8201e7828560ad3b358574b5fda74f421f3cf8ab))
* Improving how Indirect works ([8645c5e](https://github.com/monetr/validation/commit/8645c5e1e5d6e467da41efe5e068224d63b2d3b3))


### Miscellaneous

* Cleaning up README ([731c077](https://github.com/monetr/validation/commit/731c077a6f1aa35a251598ceb3088060351d7882))


### Build Automation

* **deps:** update github/codeql-action action to v4.36.0 ([#13](https://github.com/monetr/validation/issues/13)) ([c6a7c91](https://github.com/monetr/validation/commit/c6a7c917c132948c1fd2a5d56ca834928a831442))

## [1.0.6](https://github.com/monetr/validation/compare/v1.0.5...v1.0.6) (2026-05-18)


### Bug Fixes

* Adding proper generic type constraints ([31d34c0](https://github.com/monetr/validation/commit/31d34c07e0a6e339210107fff0cdbd1774563ef0))
* Fixing codeql findings ([1079f2b](https://github.com/monetr/validation/commit/1079f2bae81f4a9bcfdbb8d753ec6ce3506f2318))
* Fixing release please config ([7eeac31](https://github.com/monetr/validation/commit/7eeac31d381af4e6eb7959a9c02a65d8bd1aad4a))


### Build Automation

* Adding release please and renovate ([6a2cb13](https://github.com/monetr/validation/commit/6a2cb1382a62191b23faa231690a16f65a2a30c5))
* **deps:** pin dependencies ([#6](https://github.com/monetr/validation/issues/6)) ([d52dfcf](https://github.com/monetr/validation/commit/d52dfcfa424757248388c37d62bb73adf4305849))
* **deps:** update actions/checkout action to v2.7.0 ([#7](https://github.com/monetr/validation/issues/7)) ([281e4b8](https://github.com/monetr/validation/commit/281e4b8469580f099bae77909b9beb384bce6fa0))
* **deps:** update actions/checkout action to v6 ([#10](https://github.com/monetr/validation/issues/10)) ([7ec6345](https://github.com/monetr/validation/commit/7ec634582e3893e36bcd52186c4282ff65f76748))
* **deps:** update actions/setup-go action to v5.6.0 ([#9](https://github.com/monetr/validation/issues/9)) ([2085b26](https://github.com/monetr/validation/commit/2085b26abf20981ca8f1a412e93b234377120c91))
* **deps:** update actions/setup-go action to v6 ([#11](https://github.com/monetr/validation/issues/11)) ([6755e43](https://github.com/monetr/validation/commit/6755e43332e26907f840ecd4d47aa5d2d66e4c94))
* Removing travis ci and adding codeql ([d33acf0](https://github.com/monetr/validation/commit/d33acf0851dcf232b67000c65323e3d1abfbb635))
