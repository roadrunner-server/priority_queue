module github.com/roadrunner-server/priority_queue/tests

go 1.23

toolchain go1.23.3

require (
	github.com/google/uuid v1.6.0
	github.com/roadrunner-server/priority_queue v1.0.0
	github.com/stretchr/testify v1.9.0
)

replace github.com/roadrunner-server/priority_queue => ../

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
