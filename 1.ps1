param(
  [string]$IP   = "154.9.25.27",
  [string]$USER = "root",
  [string]$PASS = "R5ANdwgndeO6MUag"
)

$BuildDir = "Build"
if (-not (Test-Path $BuildDir)) {
    New-Item -Path $BuildDir -ItemType Directory | Out-Null
}

Write-Host "Start cross-compilation..."
go env -w GOOS=linux GOARCH=amd64 CGO_ENABLED=0
$OutputFile = Join-Path $BuildDir "ddbot"
go build -o $OutputFile ./cmd

if ($LASTEXITCODE -ne 0) {
  Write-Host "Compilation failed"
  exit 1
}

Write-Host "Uploading to $IP ..."
sshpass -p $PASS scp -o StrictHostKeyChecking=no $OutputFile "$($USER)@$($IP):/home/"
Write-Host "Completed!"
sshpass -p $PASS ssh -o StrictHostKeyChecking=no $USER@$IP "chmod +x /home/ddbot"
sshpass -p $PASS ssh -o StrictHostKeyChecking=no $USER@$IP "cd /home && mv ddbot DDBOT"
Write-Host "Rebooting..."
sshpass -p $PASS ssh -o StrictHostKeyChecking=no $USER@$IP "reboot"
sshpass -p $PASS ssh -o StrictHostKeyChecking=no $USER@$IP "exit"

Write-Host "Done, please check the server."
