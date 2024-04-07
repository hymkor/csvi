function ctrl(s)
    local b = string.byte(s)
    local a = string.byte("a")
    return string.char(b-a+1)
end
if #arg < 0 then
    os.exit(1)
end

local pid = spawn("../../csvi.exe",arg[1])
echo("../../csvi.exe "..arg[1])
expect("[CSV]")
send("jlxw")
expect("write to>")
sendln("_result")
expect("[CSV]")
send("q")
expect("[y/n]")
send("y")
wait(pid)
