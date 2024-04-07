-- This test script requires https://github.com/hymkor/expect
--
--   go install github.com/hymkor/expect@latest
--   expect test.lua

local csvi = "../../csvi.exe"
if #arg >= 1 then
    csvi = arg[1]
end

function ctrl(s)
    local b = string.byte(s)
    local a = string.byte("a")
    return string.char(b-a+1)
end

function try(op,exp)
    local source = [[あ,い,う,え,お
か,き,く,け,こ
さ,し,す,せ,そ
た,ち,つ,て,と
な,に,ぬ,ね,の
は,ひ,ふ,へ,の
ま,み,む,め,も
や,ゆ,よ
ら,り,る,れ,ろ
わ,を,ん]]

    local sourcePath = os.tmpname()..".csv"
    local resultPath = os.tmpname()..".csv"
    print("source:",sourcePath)
    print("result:",resultPath)
    local fd = assert(io.open(sourcePath,"w"))
    fd:write(source)
    fd:close()

    local pid = spawn(csvi,sourcePath)
    expect("[CSV]")

    op()

    expect("[CSV]")
    send("w")
    send(ctrl"u")
    send(ctrl"k")
    sendln(resultPath)
    send("qy")

    wait(pid)
    fd = assert(io.open(resultPath))
    local result = fd:read("*a")
    fd:close()
    os.remove(sourcePath)
    os.remove(resultPath)

    if result ~= exp then
        echo("Error: expect [["..exp.."]] but [[".. result .. "]]")
        echo("<NG>")
        os.exit(1)
    end
    echo("<OK>")
end

-- "<" and "i" --
try(function()
    send("<")
    send("i")
    sendln("new")
end,[[new,あ,い,う,え,お
か,き,く,け,こ
さ,し,す,せ,そ
た,ち,つ,て,と
な,に,ぬ,ね,の
は,ひ,ふ,へ,の
ま,み,む,め,も
や,ゆ,よ
ら,り,る,れ,ろ
わ,を,ん]])

-- "<" "$" and "a" --
try(function()
    send("<")
    send("$")
    send("a")
    sendln("new")
end,[[あ,い,う,え,お,new
か,き,く,け,こ
さ,し,す,せ,そ
た,ち,つ,て,と
な,に,ぬ,ね,の
は,ひ,ふ,へ,の
ま,み,む,め,も
や,ゆ,よ
ら,り,る,れ,ろ
わ,を,ん]])

-- "<" "$" and "x" --
try(function()
    send("<")
    send("$")
    send("x")
end,[[あ,い,う,え
か,き,く,け,こ
さ,し,す,せ,そ
た,ち,つ,て,と
な,に,ぬ,ね,の
は,ひ,ふ,へ,の
ま,み,む,め,も
や,ゆ,よ
ら,り,る,れ,ろ
わ,を,ん]])

-- "<" and "x" --
try(function()
    send("<")
    send("x")
end,[[い,う,え,お
か,き,く,け,こ
さ,し,す,せ,そ
た,ち,つ,て,と
な,に,ぬ,ね,の
は,ひ,ふ,へ,の
ま,み,む,め,も
や,ゆ,よ
ら,り,る,れ,ろ
わ,を,ん]])

--- ">" and "o" ---
try(function()
    send(">")
    send("o")
    sendln("ががが")
end,[[あ,い,う,え,お
か,き,く,け,こ
さ,し,す,せ,そ
た,ち,つ,て,と
な,に,ぬ,ね,の
は,ひ,ふ,へ,の
ま,み,む,め,も
や,ゆ,よ
ら,り,る,れ,ろ
わ,を,ん
ががが]])

--- ">" and "O" ---
try(function()
    send(">")
    send("O")
    sendln("ががが")
end,[[あ,い,う,え,お
か,き,く,け,こ
さ,し,す,せ,そ
た,ち,つ,て,と
な,に,ぬ,ね,の
は,ひ,ふ,へ,の
ま,み,む,め,も
や,ゆ,よ
ら,り,る,れ,ろ
ががが
わ,を,ん]])

--- "<" and "O" ---
try(function()
    send("<")
    send("O")
    sendln("ががが")
end,[[ががが
あ,い,う,え,お
か,き,く,け,こ
さ,し,す,せ,そ
た,ち,つ,て,と
な,に,ぬ,ね,の
は,ひ,ふ,へ,の
ま,み,む,め,も
や,ゆ,よ
ら,り,る,れ,ろ
わ,を,ん]])

--- "<" and "o" ---
try(function()
    send("<")
    send("o")
    sendln("ががが")
end,[[あ,い,う,え,お
ががが
か,き,く,け,こ
さ,し,す,せ,そ
た,ち,つ,て,と
な,に,ぬ,ね,の
は,ひ,ふ,へ,の
ま,み,む,め,も
や,ゆ,よ
ら,り,る,れ,ろ
わ,を,ん]])

--- ">" and "D" ---
try(function()
    send(">")
    send("D")
end,[[あ,い,う,え,お
か,き,く,け,こ
さ,し,す,せ,そ
た,ち,つ,て,と
な,に,ぬ,ね,の
は,ひ,ふ,へ,の
ま,み,む,め,も
や,ゆ,よ
ら,り,る,れ,ろ]])

--- "<" and "D" ---
try(function()
    send("<")
    send("D")
end,[[か,き,く,け,こ
さ,し,す,せ,そ
た,ち,つ,て,と
な,に,ぬ,ね,の
は,ひ,ふ,へ,の
ま,み,む,め,も
や,ゆ,よ
ら,り,る,れ,ろ
わ,を,ん]])

-- "<" and "r" quotations --
try(function()
    send("<")
    send("r")
    send(ctrl"u")
    send(ctrl"k")
    send("ぎ")
    send(ctrl"q")
    send(ctrl"j")
    sendln("ゃあ")
end,[["ぎ
ゃあ",い,う,え,お
か,き,く,け,こ
さ,し,す,せ,そ
た,ち,つ,て,と
な,に,ぬ,ね,の
は,ひ,ふ,へ,の
ま,み,む,め,も
や,ゆ,よ
ら,り,る,れ,ろ
わ,を,ん]])
