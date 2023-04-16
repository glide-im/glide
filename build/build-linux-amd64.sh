cd ../cmd/im_service || exit

export CGO_ENABLED=0
export GOOS=linux
export GOHOSTOS=linux
export GOARCH=amd64

echo 'build...'
go build
echo 'build complete'
cp ../../config/config.toml config.toml
tar -czvf ./im_service_linux_amd64.tar.gz im_service config.toml
rm config.toml
rm im_service
read -p 'complete.'