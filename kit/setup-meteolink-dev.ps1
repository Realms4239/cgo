#Requires -RunAsAdministrator
# meteolink.dev / meteolink.vm -> la VM qui heberge le dashboard.
# Idempotent, IP auto-decouverte (health-check subnet VMware),
# remplace les anciennes entrees au lieu d'empiler.
$hosts = "$env:SystemRoot\System32\drivers\etc\hosts"

# 1) decouvrir l'IP du dashboard : health-check sur les IP probables
$vmIP = $null
foreach ($ip in @("192.168.174.131","192.168.174.130","192.168.174.129","192.168.174.128","192.168.174.132","192.168.174.133","192.168.174.134","192.168.174.135","192.168.174.136","192.168.174.137","192.168.174.138","192.168.174.139","192.168.174.140","192.168.174.141","192.168.174.142","192.168.174.143","192.168.174.144","192.168.174.145","192.168.174.146","192.168.174.147","192.168.174.148","192.168.174.149","192.168.174.150")) {
  try {
    $r = Invoke-WebRequest -Uri "http://${ip}:9090/api/health" -TimeoutSec 2 -UseBasicParsing -ErrorAction Stop
    if ($r.StatusCode -eq 200) { $vmIP = $ip; break }
  } catch { }
}
if (-not $vmIP) {
  $vmIP = Read-Host "Dashboard introuvable sur le subnet - entrez l'IP de la VM (ex. 192.168.174.131)"
}

# 2) reecrire les entrees meteolink (remplace, pas d'empilement)
$lines = Get-Content $hosts | Where-Object { $_ -notmatch "meteolink\.(dev|vm)" }
$lines += "${vmIP}" + [char]9 + "meteolink.dev"
$lines += "${vmIP}" + [char]9 + "meteolink.vm"
Set-Content -Path $hosts -Value $lines -Force
Write-Host "meteolink.dev + meteolink.vm -> ${vmIP}"

# 3) HSTS .dev : avertir, mkcert optionnel
Write-Host "NOTE : .dev est HSTS (Chrome force https). Pour http simple : utilisez meteolink.vm"
Write-Host "Test : curl http://meteolink.vm:9090/api/health"
