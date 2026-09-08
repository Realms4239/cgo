<#PSScriptInfo>
.VERSION 1.2.3
.AUTHOR Meteolink
.DESCRIPTION Repare le reseau hote VMware (VMnet8 retombe en APIPA 169.254.x : invitée saine mais injoignable). A executer ELEVE : clic-droit -> Executer en tant qu'administrateur. Idempotent.
#>
#Requires -RunAsAdministrator
$ErrorActionPreference = 'Stop'
$Alias = "VMware Network Adapter VMnet8"
$Gw = "192.168.174.1"

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
