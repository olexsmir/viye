if vim.g.loaded_viye_plugin ~= nil then return end
vim.g.loaded_viye_plugin = true

local function indent(t) return #t:match("^%s*") end

local function tokenize(s)
  local out, buf, q = {}, "", nil
  for i = 1, #s do
    local c = s:sub(i, i)
    if q then
      if c == q then q = nil else buf = buf .. c end
    elseif c == "'" or c == '"' then
      q = c
    elseif c:match("%s") then
      if #buf > 0 then
        table.insert(out, buf); buf = ""
      end
    else
      buf = buf .. c
    end
  end
  if #buf > 0 then table.insert(out, buf) end
  return out
end

local function children_of(cmd_lnum)
  local depth = indent(vim.fn.getline(cmd_lnum))
  local first, last = cmd_lnum + 1, nil
  for i = first, vim.fn.line("$") do
    local l = vim.fn.getline(i)
    if l == "" or indent(l) <= depth then break end
    last = i
  end
  return first, last
end

local function parent_of(lnum)
  local d = indent(vim.fn.getline(lnum))
  for i = lnum - 1, 1, -1 do
    local l = vim.fn.getline(i)
    if l:find("%S") and not l:match("^%s*[|:]") and indent(l) < d then return i end
  end
end

local function ancestors(cmd_lnum)
  local parts, lnum = {}, cmd_lnum
  while lnum do
    local l = vim.fn.getline(lnum)
    local line = l:match("^%s*[+-] (.*)") or vim.trim(l)
    local words = tokenize(line)
    for i = #words, 1, -1 do table.insert(parts, 1, words[i]) end
    lnum = parent_of(lnum)
  end
  return parts
end

local function body_of(cmd_lnum)
  local depth = indent(vim.fn.getline(cmd_lnum))
  local first, last = children_of(cmd_lnum)
  local body = {}
  if last then
    for i = first, last do
      local l = vim.fn.getline(i)
      local val = indent(l) == depth + 2 and l:match("^%s*: (.*)")
      if val then
        table.insert(body, ": " .. val)
      end
    end
  end
  return body
end

local function run(cmd_lnum)
  local cmd = ancestors(cmd_lnum)
  if #cmd == 0 then return end
  table.insert(cmd, 1, "viye")
  local body = body_of(cmd_lnum)
  if #body > 0 then vim.list_extend(vim.list_extend(cmd, { "--" }), body) end
  local res = vim.system(cmd, { text = true }):wait(2000)
  if res.code ~= 0 then
    vim.notify(vim.trim(res.stderr), vim.log.levels.ERROR)
    return
  end
  return res.stdout
end

local function insert_output(cmd_lnum, out)
  local pad = string.rep(" ", indent(vim.fn.getline(cmd_lnum)) + 2)
  local lines = vim.split(out, "\n", { trimempty = true })
  for i, l in ipairs(lines) do lines[i] = pad .. l end
  vim.fn.append(cmd_lnum, lines)
end

local function toggle_bullet(cmd_lnum)
  vim.fn.setline(cmd_lnum, (vim.fn.getline(cmd_lnum):gsub("^[+-] ", function(c)
    return c == "+" and "-" or "+"
  end, 1)))
end

local function exec()
  local lnum = vim.fn.line(".")
  local line = vim.fn.getline(lnum)
  local is_out = line:match("^%s*|")
  local cmd_lnum = (line:match("^%s*[|:]") and parent_of(lnum)) or lnum
  if not cmd_lnum or vim.fn.getline(cmd_lnum):match("^==%s") then return end
  local out
  if not is_out then
    out = run(cmd_lnum)
    if not out then return end
  end
  local first, last = children_of(cmd_lnum)
  if last then vim.cmd(string.format("%d,%ddelete _", first, last)) end
  if out and out ~= "" then insert_output(cmd_lnum, out) end
  if not is_out then toggle_bullet(cmd_lnum) end
end

vim.filetype.add { extension = { viye = "viye" } }
local group = vim.api.nvim_create_augroup("viye", { clear = true })
vim.api.nvim_create_autocmd("BufEnter", {
  pattern = "*",
  group = group,
  callback = function()
    if vim.api.nvim_buf_get_name(0) == "" then
      vim.bo.filetype = "viye"
    end
  end,
})
vim.api.nvim_create_autocmd("FileType", {
  pattern = "viye",
  group = group,
  callback = function()
    vim.keymap.set("n", "<CR>", exec, { buffer = true, silent = true })
  end,
})
