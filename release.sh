go build -o egg-linux
GOARCH=arm64 go build -o egg-linux-arm64
GOARCH=riscv64 go build -o egg-linux-riscv64

GOOS=plan9 go build -o egg-plan9

GOOS=darwin go build -o egg-darwin
GOOS=darwin GOARCH=arm64 go build -o egg-darwin-arm64

GOOS=windows go build -o egg-windows.exe
GOOS=windows GOARCH=arm64 go build -o egg-windows-arm64.exe
