# Change Log


All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).


## [Unreleased]

### Changed

- Updated dependencies
- Updated repo


## [0.12.0] - 2020-02-17

### Changed

- Updated dependencies


## [0.11.0] - 2020-02-10

### Added

- `transport/grpc`: Error encoder handler


## [0.10.0] - 2020-01-13

### Added

- Operation name middleware
- Transport error handler tools


## [0.9.0] - 2020-01-12

### Changed

- `endpoint`: Failer middleware now uses a simpler error matcher function type **(Breaking change)**
- `transport/http`: Problem converter returns an `interface{}` from now **(Breaking change)**

### Removed

- `endpoint`: Deprecated Business error middleware **(Breaking change)**
- `transport/http`: Deprecated `ProblemErrorEncoder` **(Breaking change)**
- `transport/http`: Problem converter. See https://github.com/sagikazarmark/appkit **(Breaking change)**
- `transport/grpc`: Status converter. See https://github.com/sagikazarmark/appkit **(Breaking change)**


## [0.8.0] - 2020-01-09

### Changed

- Update dependencies


## [0.7.1] - 2020-01-08

### Fixed

- Do not convert gRPC errors twice


## [0.7.0] - 2020-01-08

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


[Unreleased]: https://github.com/sagikazarmark/kitx/compare/v0.12.0...HEAD
[0.12.0]: https://github.com/sagikazarmark/kitx/compare/v0.11.0...v0.12.0
[0.11.0]: https://github.com/sagikazarmark/kitx/compare/v0.10.0...v0.11.0
[0.10.0]: https://github.com/sagikazarmark/kitx/compare/v0.9.0...v0.10.0
[0.9.0]: https://github.com/sagikazarmark/kitx/compare/v0.8.0...v0.9.0
[0.8.0]: https://github.com/sagikazarmark/kitx/compare/v0.7.1...v0.8.0
[0.7.1]: https://github.com/sagikazarmark/kitx/compare/v0.7.0...v0.7.1
[0.7.0]: https://github.com/sagikazarmark/kitx/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/sagikazarmark/kitx/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/sagikazarmark/kitx/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/sagikazarmark/kitx/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/sagikazarmark/kitx/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/sagikazarmark/kitx/compare/v0.1.0...v0.2.0
