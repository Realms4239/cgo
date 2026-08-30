#Requires -RunAsAdministrator
# meteolink.dev → local + VM (portable, idempotent)
$hosts = "$env:SystemRoot\System32\drivers\etc\hosts"
$entries = @(
  "127.0.0.1`tmeteolink.dev",
  "192.168.174.128`tmeteolink.vm"
)
foreach ($e in $entries) {
  $ip = ($e -split "`t")[0]; $hostName = ($e -split "`t")[1]
  if (Select-String -Path $hosts -Pattern "^\s*$ip\s+.*\b$hostName\b" -Quiet) { Write-Host "$hostName déjà dans hosts" }
  else { Add-Content -Path $hosts -Value $e; Write-Host "$hostName → $ip ajouté" }
}
# HSTS .dev : générer un cert local si mkcert présent
if (Get-Command mkcert -ErrorAction SilentlyContinue) {
  mkcert meteolink.dev "*.meteolink.dev" meteolink.vm localhost 127.0.0.1 192.168.174.128
  Write-Host "cert mkcert généré — lancez : cgo --serve --tls-cert meteolink.dev+4.pem --tls-key meteolink.dev+4-key.pem"
} else {
  Write-Host "mkcert non trouvé — http://meteolink.dev:9090 marchera, mais Chrome forcera https pour .dev (HSTS). Préférez http://localhost:9090 ou http://meteolink.vm:9090, ou installez mkcert https://github.com/FiloSottile/mkcert"
}
Write-Host "Fait. Test : ping meteolink.dev ; curl http://meteolink.dev:9090/api/health"
