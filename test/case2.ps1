Set-PSDebug -Strict

function Equal-Bytes($a,$b){
    if ($a.Length -ne $b.Length) {
        return $false
    }
    for ($i = 0 ; $i -lt $a.Length ; $i++ ){
        if ( $a[$i] -ne $b[$i] ){
            return $false
        }
    }
    return $true
}

function Try-Test($source,$expect) {
    $sourcePath = [System.IO.Path]::GetTempFileName() + ".csv"
    $resultPath = [System.IO.Path]::GetTempFileName() + ".csv"

    [System.IO.File]::WriteAllBytes($sourcePath,$source)
    ..\csvi -auto "j|l|r|foo|w|$resultPath|q|y" $sourcePath
    $result = [System.io.File]::ReadAllBytes($resultPath)

    if ( -not (Equal-Bytes $result $expect) ){
        Write-Error "[NG]"
        Write-Host "<<<<< expect"
        Write-Host $expect
        Write-Host "-----"
        Write-Host $result
        Write-Host ">>>>> result"
        exit(1)
    }
    Write-Host "[OK]"
    Remove-Item $sourcePath
    Remove-Item $resultPath
}

$source = @"
あ,い,う,え,お
か,き,く,け,こ
さ,し,す,せ,そ
た,ち,つ,て,と
な,に,ぬ,ね,の
は,ひ,ふ,へ,の
ま,み,む,め,も
や,ゆ,よ
ら,り,る,れ,ろ
わ,を,ん
"@

$expect = @"
あ,い,う,え,お
か,foo,く,け,こ
さ,し,す,せ,そ
た,ち,つ,て,と
な,に,ぬ,ね,の
は,ひ,ふ,へ,の
ま,み,む,め,も
や,ゆ,よ
ら,り,る,れ,ろ
わ,を,ん
"@

Write-Host "### Test-1: UTF16LE ###"
Try-Test ([System.Text.Encoding]::Unicode.GetBytes($source)) `
         ([System.Text.Encoding]::Unicode.GetBytes($expect))
Write-Host "### Test-2: UTF16BE ###"
Try-Test ([System.Text.Encoding]::BigEndianUnicode.GetBytes($source)) `
         ([System.Text.Encoding]::BigEndianUnicode.GetBytes($expect))
Write-Host "### Test-3: ANSI ###"
Try-Test ([System.Text.Encoding]::Default.GetBytes($source)) `
         ([System.Text.Encoding]::Default.GetBytes($expect))
