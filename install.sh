version=$(git describe --tags --always --dirty)
commit=$(git rev-parse --short HEAD)

CGO_ENABLED=1 go build -ldflags="-X 'main.Version=${version}' -X 'main.Commit=${commit}'" -o dist/gopad go.gopad.dev/gopad
sudo install -Dm755 dist/gopad -t /usr/bin
rm -r ./dist
