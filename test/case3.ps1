# Test: concatinating multiple-csvfiles

Set-PSDebug -Strict
Set-Location (Split-Path $MyInvocation.MyCommand.path)

$source1path = [System.IO.Path]::GetTempFileName() + ".csv"
Write-Host ("first-source-file: {0}" -f $source1path)
Write-Output "first" | Out-File $source1path

$source2path = [System.IO.Path]::GetTempFileName() + ".csv"
Write-Host ("second-source-file: {0}" -f $source2path)
Write-Output "second" | Out-File $source2path

$source3path = [System.IO.Path]::GetTempFileName() + ".csv"
Write-Host ("third-source-file: {0}" -f $source3path)
Write-Output "third" | Out-File $source3path

$resultPath = [System.IO.Path]::GetTempFileName() + ".csv"
Write-Host ("output-file: {0}" -f $resultPath)

csvi -auto "w|$resultPath|q|y" $source1path $source2path $source3path

$expect = "first`r`nsecond`r`nthird`r`n"
$result = ( Get-Content $resultPath -Raw)

if ( $result -ne $expect ){
    Write-Host ("Expect '{0}', but '{1}'" -f ($expect,$result))
    exit 1
}
Remove-Item $source1path
Remove-Item $source2path
Remove-Item $source3path
Remove-Item $resultPath
Write-Host "-> OK"
exit 0
