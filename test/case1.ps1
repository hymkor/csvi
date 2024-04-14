Set-PSDebug -Strict

function Try-Test($op,$exp){
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

    $sourcePath = [System.IO.Path]::GetTempFileName() + ".csv"
    $resultPath = [System.IO.Path]::GetTempFileName() + ".csv"
    $fullop = "$op|w|$resultPath|q|y"
    Write-Host "source:" $sourcePath
    Write-Host "result:" $resultPath
    Write-Host "operation:" $fullop
    Write-Output $source | Out-File $sourcePath -Encoding utf8NoBOM -NoNewline

    ..\csvi -auto $fullop $sourcePath
    $result = Get-Content $resultPath -Encoding utf8NoBOM -Raw
    Remove-Item $sourcePath
    Remove-Item $resultPath

    if ( $result -ne $exp ){
        Write-Error "Error:"
        Write-Host "Expect:---"
        Write-Host $exp
        Write-Host "--- but,---"
        Write-Host $result
        Write-Host "---"
        exit(1)
    }
    Write-Output "<OK>"
}

Try-Test "<|i|new" @"
new,あ,い,う,え,お
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

Try-Test "<|$|a|new" @"
あ,い,う,え,お,new
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

Try-Test "<|$|x" @"
あ,い,う,え
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

Try-Test "<|x" @"
い,う,え,お
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

Try-Test ">|o|ががが" @"
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
ががが
"@

Try-Test ">|O|ががが" @"
あ,い,う,え,お
か,き,く,け,こ
さ,し,す,せ,そ
た,ち,つ,て,と
な,に,ぬ,ね,の
は,ひ,ふ,へ,の
ま,み,む,め,も
や,ゆ,よ
ら,り,る,れ,ろ
ががが
わ,を,ん
"@

Try-Test "<|O|ががが" @"
ががが
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

Try-Test "<|o|ががが" @"
あ,い,う,え,お
ががが
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

Try-Test ">|D" @"
あ,い,う,え,お
か,き,く,け,こ
さ,し,す,せ,そ
た,ち,つ,て,と
な,に,ぬ,ね,の
は,ひ,ふ,へ,の
ま,み,む,め,も
や,ゆ,よ
ら,り,る,れ,ろ
"@

Try-Test "<|D" @"
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

Try-Test "<|r|ぎ`r`nゃあ" @"
"ぎ
ゃあ",い,う,え,お
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
