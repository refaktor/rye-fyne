# Minimal rye-fyne uninstallation script for Windows

$INSTALL_DIR = "$env:LOCALAPPDATA\rye-fyne\bin"
$targetPath = "$INSTALL_DIR\rye-fyne.exe"

if (Test-Path $targetPath) {
    Write-Host "Removing rye-fyne..."
    Remove-Item $targetPath -Force
    
    # Remove from PATH
    $currentPath = [Environment]::GetEnvironmentVariable("PATH", [EnvironmentVariableTarget]::User)
    if ($currentPath -like "*$INSTALL_DIR*") {
        $newPath = ($currentPath -split ';' | Where-Object { $_ -ne $INSTALL_DIR }) -join ';'
        [Environment]::SetEnvironmentVariable("PATH", $newPath, [EnvironmentVariableTarget]::User)
        Write-Host "Removed from PATH"
    }
    
    # Remove empty directory
    if ((Get-ChildItem $INSTALL_DIR -ErrorAction SilentlyContinue).Count -eq 0) {
        Remove-Item $INSTALL_DIR -Force -ErrorAction SilentlyContinue
    }
    
    Write-Host "rye-fyne uninstalled successfully!"
} else {
    Write-Host "rye-fyne not found in $INSTALL_DIR"
}
