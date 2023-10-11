ROOT="$(dirname "$(dirname "$(readlink -fm "$0")")")"
cd $(dirname "$0")
go run $ROOT/src/server/main.go 