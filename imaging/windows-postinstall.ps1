param(
    [Parameter(Mandatory=$true)][string]$PlanPath,
    [string]$DomainJoinUser = $env:REFORGE_DOMAIN_USER,
    [string]$DomainJoinPassword = $env:REFORGE_DOMAIN_PASSWORD
)

$ErrorActionPreference = "Stop"
$plan = Get-Content -Raw -Path $PlanPath | ConvertFrom-Json

Write-Host "ReForge post-install for $($plan.host.hostname)"

if ($plan.department -and $plan.department.computer_name_pattern) {
    $serial = (Get-CimInstance Win32_BIOS).SerialNumber
    $name = $plan.department.computer_name_pattern
    $name = $name.Replace("{DEPT}", $plan.department.code)
    $name = $name.Replace("{SERIAL}", $serial)
    $name = $name.Substring(0, [Math]::Min(15, $name.Length))
    if ($env:COMPUTERNAME -ne $name) {
        Rename-Computer -NewName $name -Force
    }
}

foreach ($pkg in ($plan.software | Sort-Object install_order)) {
    if ([string]::IsNullOrWhiteSpace($pkg.silent_install)) {
        Write-Warning "No silent command configured for $($pkg.name)"
        continue
    }
    Write-Host "Installing $($pkg.name)"
    $p = Start-Process -FilePath "cmd.exe" -ArgumentList "/c", $pkg.silent_install -Wait -PassThru -NoNewWindow
    if ($p.ExitCode -notin 0,1641,3010) {
        throw "$($pkg.name) installer returned exit code $($p.ExitCode)"
    }
}

if ($plan.department -and $plan.department.ad_ou -and $DomainJoinUser -and $DomainJoinPassword) {
    $secure = ConvertTo-SecureString $DomainJoinPassword -AsPlainText -Force
    $cred = [pscredential]::new($DomainJoinUser, $secure)
    $domain = ($plan.department.ad_ou -split ',DC=' | Select-Object -Skip 1) -join '.'
    if ($domain) {
        Add-Computer -DomainName $domain -OUPath $plan.department.ad_ou -Credential $cred -Force
    }
}

foreach ($script in $plan.department.post_install_scripts) {
    if (Test-Path $script) {
        & powershell.exe -NoProfile -ExecutionPolicy Bypass -File $script
    }
}

Write-Host "ReForge post-install stage completed."
