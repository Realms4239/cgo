<#PSScriptInfo>
.VERSION 1.2.3
.GUID 3f2b1c4d-9a8e-4f1d-b2c3-d4e5f6a7b8c9
.AUTHOR Meteolink
.DESCRIPTION Tunnel Host Windows -> VM Ubuntu (VirtualBox/VMware) : forwards NAT, nom DNS, cle SSH, confiance HTTPS, verification. Idempotent : relancable sans risque.
#>
<#
.SYNOPSIS
  Amene une VM Ubuntu au dashboard securise, depuis un PC Windows, sans rien supposer.
.DESCRIPTION
  1. Localise la VM (parametre -VmName, unique trouvee, ou choix interactif).
  2. Lit son mode reseau (NAT/pont/hote-only) sans l'allumer.
  3. NAT VirtualBox : pose les forwards 2222->22 et 9090->9090 (modifyvm eteinte, controlvm a chaud).
     Pont/hote-only : acces direct a l'IP invitee (pas de forward).
  4. Ecrit meteolink.dev dans hosts (admin requis, sinon affiche la ligne exacte).
  5. Cree la cle ed25519 si absente, la pose via mot de passe tape dans ssh (jamais stocke).
  6. Rapporte le certificat et l'installe (magasin machine, repli utilisateur).
  7. Verifie TCP + HTTPS + ouvre le navigateur (sauf -NoBrowser).
.EXAMPLE
  .\host-tunnel.ps1
.EXAMPLE
  .\host-tunnel.ps1 -VmName "ubuntu" -User operateur -WhatIf
#>
[CmdletBinding(SupportsShouldProcess = $true)]
param(
  [string]$VmName = "",
  [string]$User = "",
  [int]$SshPort = 2222,
  [int]$DashPort = 9090,
  [string]$KeyPath = "$HOME\.ssh\id_ed25519",
  [switch]$NoBrowser
)

$ErrorActionPreference = 'Stop'
try { [Console]::OutputEncoding = [System.Text.Encoding]::UTF8 } catch { }
$LogFile = Join-Path $PSScriptRoot ("host-tunnel-" + (Get-Date -Format 'yyyyMMdd-HHmmss') + ".log")
try { Start-Transcript -Path $LogFile -Append | Out-Null } catch { }

function Step($t) { Write-Host ""; Write-Host ">> $t" -ForegroundColor Cyan }
function Ok($t) { Write-Host "  [OK] $t" -ForegroundColor Green }
function Info($t) { Write-Host "  [..] $t" -ForegroundColor Yellow }
function Die($t) { Write-Host "  [X] $t" -ForegroundColor Red; try { Stop-Transcript | Out-Null } catch { }; exit 1 }

function Test-TcpRapide($Host_, $Port, $Ms = 5000) {
  # Test-NetConnection peut pendre 20 s+ ; TcpClient + timeout explicite.
  $c = New-Object Net.Sockets.TcpClient
  try {
    $iar = $c.BeginConnect($Host_, [int]$Port, $null, $null)
    if (-not $iar.AsyncWaitHandle.WaitOne($Ms)) { return $false }
    $c.EndConnect($iar)
    return $true
  } catch { return $false } finally { $c.Close() }
}

function Find-Exe($names, $paths) {
  foreach ($n in $names) {
    $c = Get-Command $n -ErrorAction SilentlyContinue
    if ($c) { return $c.Source }
  }
  foreach ($p in $paths) { if (Test-Path $p) { return $p } }
  return $null
}

$VBoxManage = Find-Exe @('VBoxManage.exe', 'VBoxManage') @('C:\Program Files\Oracle\VirtualBox\VBoxManage.exe')
$VmRun = Find-Exe @('vmrun.exe', 'vmrun') @('C:\Program Files (x86)\VMware\VMware Workstation\vmrun.exe', 'C:\Program Files\VMware\VMware Workstation\vmrun.exe')
$IsAdmin = ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)

Step "0/7 — outils hotes"
if (-not (Get-Command ssh -ErrorAction SilentlyContinue)) { Die "client OpenSSH absent : winget install --id Microsoft.OpenSSH.Client --source winget" }
Ok "ssh present"
if (-not $VBoxManage -and -not $VmRun) { Die "ni VBoxManage ni vmrun trouves — installez VirtualBox ou VMware Workstation" }
$Hypervisor = if ($VBoxManage) { "virtualbox" } else { "vmware" }
Ok "hyperviseur : $Hypervisor"

Step "1/7 — localiser la VM"
$VmPath = ""
$VmRegName = ""
if ($Hypervisor -eq "virtualbox") {
  $list = & $VBoxManage list vms 2>&1 | Out-String
  $found = @()
  foreach ($ln in ($list -split "`n")) {
    if ($ln -match '"([^"]+)"\s+\{([0-9a-f-]+)\}') { $found += [pscustomobject]@{ Name = $Matches[1]; Uuid = $Matches[2] } }
  }
  if ($VmName -ne "") {
    $pick = $found | Where-Object { $_.Name -eq $VmName } | Select-Object -First 1
    if (-not $pick) { Die "VM '$VmName' non enregistree dans VirtualBox (Machine > Ajouter... d'abord)" }
    $VmRegName = $pick.Name
  } elseif ($found.Count -eq 1) {
    $VmRegName = $found[0].Name
    Ok "VM unique : $VmRegName"
  } elseif ($found.Count -eq 0) {
    Die "aucune VM enregistree — ajoutez-la dans VirtualBox (Machine > Ajouter...)"
  } else {
    Write-Host "  Plusieurs VMs :"
    for ($i = 0; $i -lt $found.Count; $i++) { Write-Host ("   [{0}] {1}" -f $i, $found[$i].Name) }
    $ch = Read-Host "  Numero"
    if ($ch -notmatch '^\d+$' -or [int]$ch -ge $found.Count) { Die "choix invalide" }
    $VmRegName = $found[[int]$ch].Name
  }
  $info = & $VBoxManage showvminfo $VmRegName --machinereadable 2>&1 | Out-String
  if ($info -match 'nic1="([^"]+)"') { $NicMode = $Matches[1] } else { $NicMode = "inconnu" }
  if ($info -match 'CfgFile="([^"]+)"') { $VmPath = $Matches[1] }
} else {
  # VMware : chemin .vmx (fichier direct, pas de registre obligatoire)
  if ($VmName -ne "" -and (Test-Path $VmName)) { $VmPath = (Resolve-Path $VmName).Path }
  else {
    $roots = @()
    foreach ($d in @('D:\', 'C:\')) { if (Test-Path $d) { $roots += $d } }
    $homeVmware = Join-Path $HOME "vmware"
    if (Test-Path $homeVmware) { $roots += $homeVmware }
    $cands = @()
    foreach ($r in $roots) { $cands += @(Get-ChildItem $r -Filter *.vmx -ErrorAction SilentlyContinue | Where-Object { -not $_.PSIsContainer }) }
    foreach ($d in @('D:\VMs', 'C:\VMs') | Where-Object { Test-Path $_ }) {
      $cands += @(Get-ChildItem $d -Filter *.vmx -Recurse -Depth 2 -ErrorAction SilentlyContinue | Where-Object { -not $_.PSIsContainer })
    }
    if ($cands.Count -eq 1) { $VmPath = $cands[0].FullName }
    elseif ($cands.Count -eq 0) { Die "aucun .vmx trouve (racines D:/ C:/, ~/vmware, D:/VMs, C:/VMs) — passez -VmName CHEMIN" }
    else {
      Write-Host "  Plusieurs .vmx :"
      for ($i = 0; $i -lt $cands.Count; $i++) { Write-Host ("   [{0}] {1}" -f $i, $cands[$i].FullName) }
      $ch = Read-Host "  Numero"
      if ($ch -notmatch '^\d+$' -or [int]$ch -ge $cands.Count) { Die "choix invalide" }
      $VmPath = $cands[[int]$ch].FullName
    }
  }
  $NicMode = "inconnu"
  $txt = Get-Content $VmPath -Raw -ErrorAction SilentlyContinue
  if ($txt -match '(?m)^ethernet0\.connectionType\s*=\s*"([^"]+)"') { $NicMode = $Matches[1] }
  $VmRegName = [System.IO.Path]::GetFileNameWithoutExtension($VmPath)
}
Ok "VM : $VmRegName (reseau : $NicMode)"

Step "2/7 — acces reseau (forwards si NAT)"
$SshTarget = ""
$SshPortEff = 22
if ($Hypervisor -eq "virtualbox" -and ($NicMode -eq "nat" -or $NicMode -eq "inconnu")) {
  # état : eteinte => modifyvm, allumée => controlvm a chaud (syntaxe nue)
  $st = & $VBoxManage showvminfo $VmRegName --machinereadable 2>&1 | Out-String
  $running = $st -match 'VMState="running"'
  if ($running) {
    if ($PSCmdlet.ShouldProcess($VmRegName, "regles NAT a chaud")) {
      & $VBoxManage controlvm $VmRegName natpf1 delete cgo-ssh 2>&1 | Out-Null
      & $VBoxManage controlvm $VmRegName natpf1 "cgo-ssh,tcp,,$SshPort,,22" 2>&1 | Out-Null
      if ($LASTEXITCODE -ne 0) { Die "regle SSH a chaud refusee (VM verrouillee ?)" }
      & $VBoxManage controlvm $VmRegName natpf1 delete cgo-dashboard 2>&1 | Out-Null
      & $VBoxManage controlvm $VmRegName natpf1 "cgo-dashboard,tcp,,$DashPort,,9090" 2>&1 | Out-Null
    }
    Ok "forwards poses a chaud (VM allumee)"
  } else {
    if ($PSCmdlet.ShouldProcess($VmRegName, "regles NAT (VM eteinte)")) {
      & $VBoxManage modifyvm $VmRegName --natpf1 delete cgo-ssh 2>&1 | Out-Null
      & $VBoxManage modifyvm $VmRegName --natpf1 "cgo-ssh,tcp,,$SshPort,,22"
      if ($LASTEXITCODE -ne 0) { Die "regle SSH refusee : port $SshPort occupe ?" }
      & $VBoxManage modifyvm $VmRegName --natpf1 delete cgo-dashboard 2>&1 | Out-Null
      & $VBoxManage modifyvm $VmRegName --natpf1 "cgo-dashboard,tcp,,$DashPort,,9090"
    }
    Ok "forwards poses (VM eteinte) — demarrez-la"
  }
  $SshTarget = "127.0.0.1"; $SshPortEff = $SshPort
} else {
  # pont / hote-only / VMware : IP directe
  $GuestIP = ""
  if ($Hypervisor -eq "virtualbox") {
    $g = & $VBoxManage guestproperty get $VmRegName "/VirtualBox/GuestInfo/Net/0/V4/IP" 2>&1 | Out-String
    if ($g -match "Value:\s*([0-9.]+)") { $GuestIP = $Matches[1] }
  } else {
    $g = & $VmRun -T ws getGuestIPAddress $VmPath 2>&1 | Out-String
    if ($g -match "([0-9]+\.[0-9]+\.[0-9]+\.[0-9]+)") { $GuestIP = $Matches[1] }
  }
  if ($GuestIP -eq "") {
    Write-Host "  IP non resolue (VM eteinte ou additions absentes)."
    $GuestIP = Read-Host "  IP invitee (console VM : ip -br addr)"
    if ($GuestIP -notmatch "^[0-9.]+$") { Die "IP invalide" }
  }
  Ok "acces direct : $GuestIP"
  $SshTarget = $GuestIP; $SshPortEff = 22
}

Step "3/7 — nom stable (hosts)"
$HostsPath = "$env:SystemRoot\System32\drivers\etc\hosts"
$line = "$SshTarget meteolink.dev  # cgo — tableau de bord (host-tunnel)"
$raw = Get-Content $HostsPath -Raw -ErrorAction SilentlyContinue
if ($raw -match "(?m)^[0-9.]+\s+meteolink\.dev(\s|$)") {
  $raw = ($raw -split "`n" | Where-Object { $_ -notmatch "meteolink\.dev" }) -join "`n"
}
if ($PSCmdlet.ShouldProcess($HostsPath, "ajouter $line")) {
  try { ($raw.TrimEnd() + "`r`n" + $line + "`r`n") | Out-File $HostsPath -Encoding ascii -NoNewline; Ok "hosts : meteolink.dev -> $SshTarget" }
  catch { Write-Host "  [..] hosts non modifiable (pas admin) — ajoutez a la main :" -ForegroundColor Yellow; Write-Host "  $line"; }
}

Step "4/7 — cle SSH"
if ($User -eq "") { $User = Read-Host "  Utilisateur Ubuntu de la VM (vide = abandon)" ; if ($User -eq "") { Die "utilisateur vide" } }
if (-not (Test-Path $KeyPath)) {
  if ($PSCmdlet.ShouldProcess($KeyPath, "generer cle ed25519")) {
    ssh-keygen -t ed25519 -N '""' -f $KeyPath -q
    Ok "cle creee : $KeyPath"
  }
} else { Ok "cle presente : $KeyPath" }
& ssh -p $SshPortEff -i $KeyPath -o BatchMode=yes -o ConnectTimeout=4 "$User@$SshTarget" true 2>$null
if ($LASTEXITCODE -eq 0) {
  Ok "cle deja acceptee par $User@$SshTarget — rien a faire"
} else {
  Write-Host "  Mot de passe Ubuntu demande UNE fois dans ssh (jamais stocke) :"
  Get-Content ($KeyPath + ".pub") | & ssh -p $SshPortEff -i $KeyPath -o StrictHostKeyChecking=accept-new "$User@$SshTarget" "mkdir -p ~/.ssh && cat >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys && chmod 700 ~/.ssh"
}
if ($LASTEXITCODE -ne 0) { Die "pose de cle refusee — mot de passe faux ? sshd absent ? (console VM : sudo apt install -y openssh-server)" }
& ssh -p $SshPortEff -i $KeyPath -o BatchMode=yes "$User@$SshTarget" true 2>$null
if ($LASTEXITCODE -eq 0) { Ok "cle acceptee par $User@$SshTarget" } else { Die "cle refusee apres pose — reessayez" }

Step "5/7 — certificat + confiance HTTPS"
$certTmp = Join-Path $env:TEMP "meteolink-dev-cert.pem"
& scp -P $SshPortEff -i $KeyPath -o BatchMode=yes "$User@${SshTarget}:.config/cgo/cert.pem" $certTmp 2>$null
if ($LASTEXITCODE -ne 0 -or -not (Test-Path $certTmp)) { Die "certificat non rapatrie — dashboard demarre ? (kit deploy / guest-setup.sh etape 4)" }
if ($IsAdmin) {
  $cOut = & certutil -addstore root $certTmp 2>&1 | Out-String
  if ($cOut -match "déjà|already") { Ok "certificat déjà présent dans le magasin machine" }
  else { Ok "certificat dans le magasin machine" }
} else {
  $cOut = & certutil -user -addstore root $certTmp 2>&1 | Out-String
  if ($LASTEXITCODE -eq 0) { Ok "certificat dans le magasin utilisateur (ce compte)" }
  elseif ($cOut -match "déjà|already") { Ok "certificat déjà présent (ce compte)" }
  else { Write-Host "  [..] admin requis pour la confiance totale — en attendant : accepter une fois dans le navigateur" -ForegroundColor Yellow }
}

Step "6/7 — verification de bout en bout"
if (-not (Test-TcpRapide $SshTarget $DashPort 5000)) {
  Die "port $DashPort ferme sur $SshTarget (5 s) — dashboard eteint ? forward manquant ?"
}
Ok "TCP :$DashPort ouvert"
try {
  $h = Invoke-WebRequest "https://meteolink.dev:$DashPort/api/health" -UseBasicParsing -TimeoutSec 8 | ConvertFrom-Json
  if (-not $h.ok) { Die "health inattendu" }
  Ok ("dashboard sain (version " + $h.version + ", mode " + $h.mode + ") — chaine de confiance OK")
} catch { Die "HTTPS echoue : $_ (confiance ? hosts ?)" }

Step "7/7 — ouverture"
if (-not $NoBrowser) { Start-Process "https://meteolink.dev:$DashPort" }
Write-Host ""
Write-Host "RÉCAP — tout est en place :" -ForegroundColor Green
Write-Host ("  VM       : {0} ({1})" -f $VmRegName, $Hypervisor)
Write-Host ("  SSH      : {0}@{1}:{2}  (clé {3})" -f $User, $SshTarget, $SshPortEff, $KeyPath)
Write-Host ("  Dashboard: https://meteolink.dev:{0}" -f $DashPort)
Write-Host ("  Journal  : voir le fichier .log à côté du script")
Write-Host "TERMINE — https://meteolink.dev:$DashPort" -ForegroundColor Green
try { Stop-Transcript | Out-Null } catch { }
