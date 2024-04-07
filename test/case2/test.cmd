@setlocal
@set "PROMPT=$D$S$T$G$S"
copy sample UTF-8.csv
nkf32 -W8 -w16L < UTF-8.csv > UTF-16LE.csv
nkf32 -W8 -w16B < UTF-8.csv > UTF-16BE.csv
nkf32 -W8 -s    < UTF-8.csv > Shift_JIS.csv
if exist UTF-8_result.csv del UTF-8_result.csv
expect modify.lua UTF-8.csv

if exist UTF-16LE_result.csv del UTF-16LE_result.csv
expect modify.lua UTF-16LE.csv
nkf32 -W8 -w16L < UTF-8_result.csv > UTF-16LE_expect.csv
fc /b UTF-16LE_result.csv UTF-16LE_expect.csv || exit /b 1
@echo [OK:UTF-16LE]
del UTF-16LE*.csv

if exist UTF-16BE_result.csv del UTF-16BE_result.csv
expect modify.lua UTF-16BE.csv
nkf32 -W8 -w16B < UTF-8_result.csv > UTF-16BE_expect.csv
fc /b UTF-16BE_result.csv UTF-16BE_expect.csv || exit /b 1
@echo [OK:UTF-16BE]
del UTF-16BE*.csv

if exist Shift_JIS_result.csv del Shift_JIS_result.csv
expect modify.lua Shift_JIS.csv
nkf32 -W8 -s < UTF-8_result.csv > Shift_JIS_expect.csv
fc /b Shift_JIS_result.csv Shift_JIS_expect.csv || exit /b 1
@echo [OK:Shift_JIS]
@del *.csv
@endlocal
