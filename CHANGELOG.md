# Changelog

All notable changes to this project will be documented in this file.

## [v1.1.0](https://github.com/somaz94/major-tag-action/compare/v1.0.4...v1.1.0) (2026-08-07)

### Performance Improvements

- ship a prebuilt multi-arch image instead of building per run ([d1f73f0](https://github.com/somaz94/major-tag-action/commit/d1f73f087be74863bf31be709282aa88c1284204))

### Continuous Integration

- pin the self-referencing major-tag step to an exact version ([3561104](https://github.com/somaz94/major-tag-action/commit/356110469d72a1cc011a98f141061dd59c0255b9))
- add a golangci-lint config scoped to defect-finding linters ([4363630](https://github.com/somaz94/major-tag-action/commit/43636300ee28e03fdc8ba7dd66e4762080b8fe25))

### Contributors

- somaz

<br/>

## [v1.0.4](https://github.com/somaz94/major-tag-action/compare/v1.0.3...v1.0.4) (2026-07-21)

### Bug Fixes

- **tagger:** preserve auth error unwrap chain with %w instead of %v ([a0a529b](https://github.com/somaz94/major-tag-action/commit/a0a529b712c29c4ca752640308b0de7a28b36292))

### Code Refactoring

- **tagger:** propagate context through GitRunner and dedup git command wrappers ([a4d9709](https://github.com/somaz94/major-tag-action/commit/a4d9709d046433df20d4fca5696368e7d40c644d))

### Tests

- isolate global git config in tests as a regression guard ([1d881d0](https://github.com/somaz94/major-tag-action/commit/1d881d00b3f85be07c08eff0558e68ad6fcba34a))

### Continuous Integration

- remove DCO workflow ([acdb42f](https://github.com/somaz94/major-tag-action/commit/acdb42ff861d060d0b5ffb3d6493e367cead6ed1))
- adopt semantic-pr, labels, lock-threads, PR size, and auto-assign reusables ([b484a2a](https://github.com/somaz94/major-tag-action/commit/b484a2a585bce092fe99d7173b4e3904a5489776))
- use reusable stale-issues workflow ([6781b1f](https://github.com/somaz94/major-tag-action/commit/6781b1f7cf9816a02c977c63967d9e149e22d4bd))
- use reusable issue-greeting workflow ([7bc649a](https://github.com/somaz94/major-tag-action/commit/7bc649a87669e1856791020f36c8f400021bd772))
- use reusable dependabot-auto-merge workflow ([76f4a0f](https://github.com/somaz94/major-tag-action/commit/76f4a0f3d4a5bb3a4acc2d1ba833437218313966))
- use reusable contributors workflow ([8108933](https://github.com/somaz94/major-tag-action/commit/8108933e98b878b9923ad54c18829fa7919191fe))
- add ok-to-test workflow stub ([97fd4b8](https://github.com/somaz94/major-tag-action/commit/97fd4b8c9463135c56d381efadb552821a534557))
- add PR welcome workflow stub ([43871c9](https://github.com/somaz94/major-tag-action/commit/43871c94f61207b8114bf0c35b1773e90105c2a3))
- add DCO check via shared reusable workflow ([458b605](https://github.com/somaz94/major-tag-action/commit/458b605a857473db59eaa041f16a9f0cc6b85109))
- add concurrency guards to recurring workflows ([661e730](https://github.com/somaz94/major-tag-action/commit/661e730a949af93bc3c180373bd982fc7603b1e2))
- use go-docker-action-ci-action@v1 (replace inline prelude) ([34829b6](https://github.com/somaz94/major-tag-action/commit/34829b6ec87749d74e0939928093f348390094ef))

### Chores

- **deps:** bump actions/setup-go from 6 to 7 (#7) ([#7](https://github.com/somaz94/major-tag-action/pull/7)) ([b74582f](https://github.com/somaz94/major-tag-action/commit/b74582f8c0a3281d545612103cf1900449ada501))
- **deps:** bump actions/checkout from 6 to 7 (#6) ([#6](https://github.com/somaz94/major-tag-action/pull/6)) ([69e8a3a](https://github.com/somaz94/major-tag-action/commit/69e8a3a48591701745588fa2213acb11eeff3ff5))
- **deps:** bump alpine from 3.23 to 3.24 in the docker-minor group (#4) ([#4](https://github.com/somaz94/major-tag-action/pull/4)) ([80fc695](https://github.com/somaz94/major-tag-action/commit/80fc695d1a6cd3d50ec8c60badadb393d0f471b0))
- **deps:** bump softprops/action-gh-release from 2 to 3 ([b991b82](https://github.com/somaz94/major-tag-action/commit/b991b82062cf42ee3b091808f063532d7b8d25e9))
- **deps:** bump actions/github-script from 8 to 9 ([90f6753](https://github.com/somaz94/major-tag-action/commit/90f6753606801dd5987df72fdead72c7720c19e4))
- **deps:** bump dependabot/fetch-metadata from 2 to 3 ([9f6962e](https://github.com/somaz94/major-tag-action/commit/9f6962e13f39a82d362c2edc4f75537ca3c72b4b))

### Contributors

- somaz

<br/>

## [v1.0.3](https://github.com/somaz94/major-tag-action/compare/v1.0.2...v1.0.3) (2026-04-03)

### Code Refactoring

- interface-based dependency injection and code quality improvements ([c067712](https://github.com/somaz94/major-tag-action/commit/c067712ed4e042f0f6b042d4ac496d6b97e5d7fb))

### Documentation

- remove duplicate rules covered by global CLAUDE.md ([9e061e7](https://github.com/somaz94/major-tag-action/commit/9e061e76e75a9b5cdd796c45a40fbdf55385dd01))

### Chores

- remove duplicate rules from CLAUDE.md (moved to global) ([9517eaa](https://github.com/somaz94/major-tag-action/commit/9517eaa7b4adca302c9aab44e50ae4ff2fbb8ff8))
- add git config protection to CLAUDE.md ([347083f](https://github.com/somaz94/major-tag-action/commit/347083f88a8fffe2baf5bfe9617865374d774e4f))

### Contributors

- somaz

<br/>

## [v1.0.2](https://github.com/somaz94/major-tag-action/compare/v1.0.1...v1.0.2) (2026-03-25)

### Bug Fixes

- add SHA validation, error context, and security hardening ([4938cfd](https://github.com/somaz94/major-tag-action/commit/4938cfd52f52a34e38fcde82ff2947406cf68e8b))

### Documentation

- add no-push rule to CLAUDE.md ([fb8e802](https://github.com/somaz94/major-tag-action/commit/fb8e8029f92483128f29b3c939196d13c3faa31a))
- add CLAUDE.md project guide ([0100b5f](https://github.com/somaz94/major-tag-action/commit/0100b5f931ab0db1828b247646a349c2510b4a03))

### Continuous Integration

- skip auto-generated changelog and contributors commits in release notes ([a75fd45](https://github.com/somaz94/major-tag-action/commit/a75fd45a98bdca0e09857fff79f6cc770c157479))
- revert to body_path RELEASE.md in release workflow ([b28f94c](https://github.com/somaz94/major-tag-action/commit/b28f94cf2fa24014f805c45148cc22d01e275a06))
- use generate_release_notes instead of body_path in release workflow ([b154769](https://github.com/somaz94/major-tag-action/commit/b15476991f1ea4a352a3aaa22d7aa77cd2b8d680))
- migrate gitlab-mirror workflow to multi-git-mirror action ([44f3af9](https://github.com/somaz94/major-tag-action/commit/44f3af9fcd6729c4c897e74e0489b1bbf15bd804))
- use somaz94/contributors-action@v1 for contributors generation ([5884f99](https://github.com/somaz94/major-tag-action/commit/5884f99a34ee20bc806895788440217f85160edd))

### Contributors

- somaz

<br/>

## [v1.0.1](https://github.com/somaz94/major-tag-action/compare/v1.0.0...v1.0.1) (2026-03-17)

### Bug Fixes

- remove go.sum from Dockerfile COPY (no external dependencies) ([05d0865](https://github.com/somaz94/major-tag-action/commit/05d08659733942fd5f0497be41456d89eecd9cc8))

### Code Refactoring

- improve code quality and consistency ([1b83e47](https://github.com/somaz94/major-tag-action/commit/1b83e47047f5b053830a199fc2db61a20550f130))

### Tests

- raise coverage to 90% with additional edge case tests ([f70673a](https://github.com/somaz94/major-tag-action/commit/f70673a9bc8c63b5425fb13a33541455a381cdf5))

### Chores

- use published action in release workflow ([fcda881](https://github.com/somaz94/major-tag-action/commit/fcda88147b4715438bd11c65cdae3c79ddb31635))

### Contributors

- somaz

<br/>

## [v1.0.0](https://github.com/somaz94/major-tag-action/releases/tag/v1.0.0) (2026-03-16)

### Features

- implement major tag action in Go ([7c7d361](https://github.com/somaz94/major-tag-action/commit/7c7d361909a218f7ac917292e356e756fe9d3d21))

### Documentation

- add README and usage guide ([b6ded91](https://github.com/somaz94/major-tag-action/commit/b6ded915e033339ac5f44ddf78d7cf7c066bec70))

### Continuous Integration

- add GitHub Actions workflows ([59c5d0e](https://github.com/somaz94/major-tag-action/commit/59c5d0e80554fbfc80c15488fd9d10f7702b5253))

### Chores

- add project configuration files ([c38cc10](https://github.com/somaz94/major-tag-action/commit/c38cc100352ef9f5f73df1f4fc343dec857e60e0))
- remove template boilerplate files ([402caae](https://github.com/somaz94/major-tag-action/commit/402caaeedd5897a6c32defe2e5c6882aee27c1fb))

### Contributors

- somaz

<br/>

