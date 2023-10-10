ROOT="$(dirname "$(dirname "$(readlink -fm "$0")")")"
cd $(dirname "$0")
go run $ROOT/src/client/main.go 