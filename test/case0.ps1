Set-PSDebug -Strict

$resultPath = [System.IO.Path]::GetTempFileName() + ".csv"
..\csvi -auto "i|ahaha|w|$resultPath|q|y" >nul
$result = Get-Content $resultPath -Raw
if ( $result -ne "ahaha" ){
    Write-Error "Error"
    Write-Host ("Expect 'ahaha', but '{0}'" -f $result)
    exit 1
}
Remove-Item $resultPath
Write-Host "<OK>"

Write-Output "ihihi" | ..\csvi -auto "w|-|q|y" > $resultPath
$result = Get-Content $resultPath -Raw
if ( $result -ne "ihihi`r`n" ){
    Write-Error "Error"
    Write-Host ("Expect 'ihihi', but '{0}'" -f $result)
    exit 1
}
Write-Host "<OK>"

exit 0
