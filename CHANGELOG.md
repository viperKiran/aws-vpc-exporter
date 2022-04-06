# Changelog

All notable changes to this project will be documented in this file. It is following [keepachangelog] format.

## [Unreleased]

## [1.0.0] (2022-01-12)

### Added

* Exports subnets info for the given VPC ID and registers with Prometheus metrics ([#1])
* Installed pre-commit to do go lint ([#1])
* Bitbucket pipeline to do build and push docker image to ECR ([#1])
* Added changelog ([#3])
### Fixed

* Bitbucket pipeline to include ECR image tag ([#8])

### Updated dependencies

* quay.io/prometheus/busybox:latest docker digest to 2548dd9 ([#4])
* golang to v1.17.6 ([#5])
* module github.com/aws/aws-sdk-go-v2 to v1.12.0 ([#7])

[keepachangelog]: https://keepachangelog.com/en/1.0.0/

[Unreleased]: https://bitbucket.org/paytmteam/aws-vpc-exporter/branches/compare/HEAD..v1.0.0
[1.0.0]: https://bitbucket.org/paytmteam/aws-vpc-exporter/commits/tag/v1.0.0

[#1]: https://bitbucket.org/paytmteam/aws-vpc-exporter/pull-requests/1
[#3]: https://bitbucket.org/paytmteam/aws-vpc-exporter/pull-requests/3
[#4]: https://bitbucket.org/paytmteam/aws-vpc-exporter/pull-requests/4
[#5]: https://bitbucket.org/paytmteam/aws-vpc-exporter/pull-requests/5
[#7]: https://bitbucket.org/paytmteam/aws-vpc-exporter/pull-requests/7
[#8]: https://bitbucket.org/paytmteam/aws-vpc-exporter/pull-requests/8
