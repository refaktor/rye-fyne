# Minimal rye-fyne installation script for Windows

$REPO = "refaktor/rye-fyne"
$INSTALL_DIR = "$env:LOCALAPPDATA\rye-fyne\bin"

# Get latest version
$response = Invoke-RestMethod -Uri "https://api.github.com/repos/$REPO/releases/latest"
$VERSION = $response.tag_name

# Download and install
Write-Host "Installing rye-fyne $VERSION..."
New-Item -ItemType Directory -Path $INSTALL_DIR -Force | Out-Null

$downloadUrl = "https://github.com/$REPO/releases/download/$VERSION/rye-fyne.exe"
$targetPath = "$INSTALL_DIR\rye-fyne.exe"

Invoke-WebRequest -Uri $downloadUrl -OutFile $targetPath -UseBasicParsing

# Add to PATH if needed
$currentPath = [Environment]::GetEnvironmentVariable("PATH", [EnvironmentVariableTarget]::User)
if ($currentPath -notlike "*$INSTALL_DIR*") {
    [Environment]::SetEnvironmentVariable("PATH", "$currentPath;$INSTALL_DIR", [EnvironmentVariableTarget]::User)
    Write-Host "Added to PATH. Restart your terminal to use rye-fyne."
}

Write-Host "rye-fyne installed successfully!"
Write-Host "Usage: rye-fyne [your-app.rye]"
