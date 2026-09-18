cd crypto
clear
# go generate ./...
CVPKG=$(go list ./... | grep -v /mocks | tr '\n' ',')
go test -coverpkg="${CVPKG%,}" -coverprofile=tests/coverage/coverage.out $(go list ./... | grep -vE '/(tests|docs)')
go tool cover -html=./tests/coverage/coverage.out -o=./tests/coverage/index.html
cd ..