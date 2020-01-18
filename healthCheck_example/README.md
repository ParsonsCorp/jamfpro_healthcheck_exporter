# Jamf Health Check Example

Can use this file to represent the scrape endpoint.

## Process

1. Open two consoles
1. From console A, change directory to jamfpro_healthcheck_exporter
   1. run: `go run jamfpro_healthcheck_exporter.go
1. From console B, change directory to healthCheck_example
   1. run: `go run healthCheck_example.go`
   1. can adjust what health code is showing my passing: -healthcode=[0-6]
1. Navigate to `http://localhost:9613/metrics` in a browser
