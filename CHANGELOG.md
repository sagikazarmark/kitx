# Change Log


All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).


## [Unreleased]

### Added

- gRPC status converter


## [0.6.0] - 2020-01-07

### Added

- gRPC error response encoder
- HTTP error response encoder

### Changed

- Renamed `ProblemFactory` to `ProblemConverter`


## [0.5.0] - 2020-01-07

### Added

- Better problem error encoding tools

### Deprecated

- `ProblemErrorEncoder` error encoder. Use the alternative problem error encoders.


## [0.4.0] - 2019-11-14

### Added

- Failer middleware

### Deprecated

- Business error middleware. Use Failer middleware instead.


## [0.3.0] - 2019-10-03

### Added

- `http.WithStatusCode` to make a response implement the `StatusCoder` interface


## [0.2.0] - 2019-09-27

### Added

- `http` and `grpc` `ServerOptions` to wrap a slice of options into a single element
- Business error middleware
- Custom endpoint middleware chain


## 0.1.0 - 2019-09-25

- Initial release


[Unreleased]: https://github.com/sagikazarmark/kitx/compare/v0.6.0...HEAD
[0.6.0]: https://github.com/sagikazarmark/kitx/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/sagikazarmark/kitx/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/sagikazarmark/kitx/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/sagikazarmark/kitx/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/sagikazarmark/kitx/compare/v0.1.0...v0.2.0
