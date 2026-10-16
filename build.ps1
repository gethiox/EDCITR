$env:GOEXPERIMENT = "jsonv2"

Write-Host "Starting build script"

$binaryPath = "builds\EDCITR.exe"

if (Test-Path $binaryPath -ErrorAction SilentlyContinue) {
    Write-Host -foreground Yellow "Removing previous build"
    Remove-Item $binaryPath
}

Write-Host "Building..."

go-winres make --file-version source-dev --product-version source-dev --in assets/winres/winres.json --out cmd/EDCITR/rsrc

if ($LASTEXITCODE -ne 0) {
    Write-Host -foreground Red "go-winres make failed"
}

go build -trimpath -ldflags "-w -s" -o $binaryPath ./cmd/EDCITR/

if ($LASTEXITCODE -ne 0) {
    Write-Host -foreground Red "Build failed"
} else {
    Write-Host -foreground Green "Build done"
}
