Set-PSDebug -Strict
Set-Location (Split-Path $MyInvocation.MyCommand.path)

if ( -not (Test-Path 27OSAKA.CSV) ){
    curl.exe -O https://www.post.japanpost.jp/zipcode/dl/kogaki/zip/27osaka.zip
    unzip.exe 27osaka.zip
    Remove-Item 27osaka.zip
}
$start = Get-Date
..\csvi.exe -auto 'l|l|l|l|l|l|l|l|l|l|l|l|l|l|q|y' 27OSAKA.CSV
$end = Get-Date
Write-Host ($end-$start).TotalSeconds
