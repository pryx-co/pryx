module pryx-core

go 1.24.0

require (
	github.com/go-chi/chi/v5 v5.0.12
	github.com/google/uuid v1.6.0
	github.com/mattn/go-sqlite3 v1.14.22
	github.com/playwright-community/playwright-go v0.5200.1
	github.com/zalando/go-keyring v0.2.4
	gopkg.in/yaml.v3 v3.0.1
	nhooyr.io/websocket v1.8.17
)

require (
	github.com/alessio/shellescape v1.4.1 // indirect
	github.com/danieljoos/wincred v1.2.0 // indirect
	github.com/deckarep/golang-set/v2 v2.7.0 // indirect
	github.com/go-jose/go-jose/v3 v3.0.4 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
)
# Coverage configuration for Go tests

[profile.test.coverage]
# Use coverage collection instrumentation
# Generates coverage reports for tests
# Run with: go test -coverprofile=test.coverage

[cover]
# Packages to include in coverage
# Defaults to all packages
# Can be overridden with -coverpkg=package1,package2

[report]
# Coverage reporting configuration
# Generate reports with: go tool cover -html=coverage.out
# Supports: html, txt, json, coverprofile
