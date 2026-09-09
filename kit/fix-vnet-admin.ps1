<#PSScriptInfo>
.VERSION 1.2.3
.AUTHOR Meteolink
.DESCRIPTION Repare le reseau hote VMware (VMnet8 retombe en APIPA 169.254.x : invitée saine mais injoignable). A executer ELEVE : clic-droit -> Executer en tant qu'administrateur. Idempotent.
#>
#Requires -RunAsAdministrator
param([string]$GwIp = "192.168.174.1")  # passerelle du subnet NAT (kit vnet la connaît ; défaut = VMware standard)
$ErrorActionPreference = 'Stop'
# GARDE : 10.0.2.x = espace invité VirtualBox (injoignable depuis l'hôte).
# L'assigner à un adaptateur hôte casse le réseau au lieu de le réparer
# (incident v1.2.10 : 10.0.2.1/24 posé sur VMnet8 — rollback :
# Remove-NetIPAddress -IPAddress 10.0.2.1).
if ($GwIp -match '^10\.0\.2\.') {
  Write-Host "  [X] REFUS : $GwIp est dans l'espace invité VirtualBox (10.0.2.x)," -ForegroundColor Red
  Write-Host "      injoignable depuis l'hôte. Pour VirtualBox NAT, utilisez le"
  Write-Host "      port-forward (cgo.exe kit ensure), pas une adresse hôte."
  exit 3
}
if ($GwIp -notmatch '^(10\.|172\.(1[6-9]|2[0-9]|3[01])\.|192\.168\.)[0-9.]+\.[0-9]+$') {
  Write-Host "  [X] REFUS : $GwIp n'est pas une IPv4 privée valide." -ForegroundColor Red
  exit 3
}
$Alias = "VMware Network Adapter VMnet8"
$Gw = $GwIp

Write-Host ">> vnet hôte : $Alias -> $Gw/24"
$ad = Get-NetAdapter -Name $Alias -ErrorAction SilentlyContinue
if (-not $ad) { Write-Host "  [X] $Alias introuvable — VMware installe ?"; exit 3 }
$cur = Get-NetIPAddress -AddressFamily IPv4 -ErrorAction SilentlyContinue |
  Where-Object { $_.InterfaceAlias -eq $Alias -and $_.IPAddress -notlike '169.254.*' }
if ($cur | Where-Object { $_.IPAddress -eq $Gw }) {
  Write-Host "  [OK] $Gw deja en place — rien a faire"
} else {
  Get-NetIPAddress -AddressFamily IPv4 -ErrorAction SilentlyContinue |
    Where-Object { $_.InterfaceAlias -eq $Alias } |
    Remove-NetIPAddress -Confirm:$false -ErrorAction SilentlyContinue
  New-NetIPAddress -InterfaceAlias $Alias -IPAddress $Gw -PrefixLength 24 | Out-Null
  Write-Host "  [OK] $Gw/24 pose sur $Alias"
}
Restart-Service -Name 'VMware NAT Service' -Force -ErrorAction SilentlyContinue
Restart-Service -Name 'VMware DHCP Service' -Force -ErrorAction SilentlyContinue
Write-Host "  [OK] services NAT/DHCP relances"
$ok = Get-NetIPAddress -AddressFamily IPv4 -ErrorAction SilentlyContinue |
  Where-Object { $_.InterfaceAlias -eq $Alias -and $_.IPAddress -eq $Gw }
if ($ok) { Write-Host "TERMINE — relancez : cgo.exe kit ensure" -ForegroundColor Green; exit 0 }
Write-Host "  [X] $Gw toujours absent — redemarrez le poste" -ForegroundColor Red
exit 3
