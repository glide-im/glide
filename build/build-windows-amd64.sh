cd ../cmd/im_service || exit

export CGO_ENABLED=0
export GOOS=windows
export GOHOSTOS=windows
export GOARCH=amd64

echo 'build...'
go build -o im_service.exe
echo 'build complete'
cp ../../config/config.toml config.toml
tar -czvf ./im_service_windows_amd64.tar.gz im_service.exe config.toml
rm config.toml
rm im_service
read -p 'complete.'