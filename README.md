# Go Meetup Brisbane, May 2019

## To Run

```go get github.com/brettniven/gomeetupbris```

Then in that dir:

```go test ./... -cover``` (or run tests from an IDE)

## Notes
 * This repo has been copied from a private project of min and package names altered accordingly
 * This has been published to show the test pattern described in the session
 * It can also be a reference point for useful libraries, code quality tools, a microservice template, and a reference point for a data pipeline approach using channels 
 * It will be of no use as an actual service, as to run, it requires a grpc server and appropriate runtime config
 * I've committed the vendor dir as some lib dependencies would otherwise be non-go-gettable as they live in private repos
