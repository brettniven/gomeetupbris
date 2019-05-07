# Go Meetup Brisbane, May 2019

[![Go Report Card](https://goreportcard.com/badge/brettniven/gomeetupbris)](https://goreportcard.com/report/brettniven/gomeetupbris)

## To Run

```go get github.com/brettniven/gomeetupbris```

Then in that dir:

```go test ./... -cover``` (or better still - run the tests from an IDE that allows you to see the coverage (such as IntelliJ, VSCode))

## Notes
 * This repo has been copied from a private project of mine and package names altered accordingly
 * This has been published to show the test pattern described in the session which was basically:
   * Focus on testing the service as opposed to the internal algorithms. Services are smaller these days. See the service as your 'unit'.
   * These tests are typically in the form of input, someExternalCall(s), output. Using dir/file based approaches make this very clean
   * The service_test tests most cases, with minimal effort. This alone results in > 85% coverage
   * Other tests can then be created to fill coverage gaps, where the tests are meaningful. In this case, pipeline_test covers error conditions that service_test wasn't granular enough for
 * This repo can also be a reference point for useful libraries, code quality tools, a microservice template, and a reference point for a data pipeline approach using channels 
 * It will be of no use as an actual service, as to run, it requires a grpc server and appropriate runtime config
 * I've committed the vendor dir as some lib dependencies would otherwise be non-go-gettable as they live in private repos
