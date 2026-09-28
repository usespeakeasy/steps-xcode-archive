# Temporary Enterprise signing patch

This Speak-owned fork is based on upstream Xcode Archive step 6.1.3
(`f2fc3386609eee5c5db9e128a1449f92fe6de54b`). It keeps automatic API-key
signing for Enterprise accounts.

Apple's Enterprise certificate endpoint rejects `DEVELOPMENT` and `DISTRIBUTION`
filters. The patched vendored `go-xcode/v2` client requests only
`IOS_DEVELOPMENT` and `IOS_DISTRIBUTION` for Enterprise accounts. App Store
accounts keep all four certificate types. Other API errors still fail signing.

Run the regression tests with:

```sh
go test -mod=vendor ./step -run 'TestCertificateQueriesMatchAccountType|TestEnterpriseCertificateQueryPreservesAPIErrors'
```

The tests use a fake HTTP client and need no Apple credentials. They also run
with the step's other unit tests in GitHub Actions. Do not run `go mod vendor`
without restoring this patch; that command replaces the patched dependency.

## Upstream and removal

No upstream PR has been submitted. Add its link here when one is opened in
[go-xcode](https://github.com/bitrise-io/go-xcode). The archive step must then
release a version that includes the fixed library.

Speak iOS must pin this fork to a full commit SHA, not a mutable branch. Keep
the normal App Store and TestFlight workflows on the official archive step.
Before removing the fork, run both QA Dev and QA Prod workflows with the fixed
official release and confirm automatic signing and Firebase distribution succeed.
