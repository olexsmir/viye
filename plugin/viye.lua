local function indent(t) return #t:match("^%s*") end
local function line_type(line)
  if not line:find("%S") then return "empty" end
  return line:find("^%s*[|:]") and "data" or "node"
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
    if line_type(l) == "node" and indent(l) < d then return i end
  end
  return nil
end

local function ancestors(cmd_lnum)
  local parts = {}
  local lnum = cmd_lnum
  while lnum do
    local l = vim.fn.getline(lnum)
    table.insert(parts, 1, l:match("^%s*[+-] (.*)") or vim.trim(l))
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
      if val then table.insert(body, ": " .. val) end
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
  for i, l in ipairs(lines) do
    lines[i] = pad .. l
  end
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
  local cmd_lnum = line_type(line) == "data" and parent_of(lnum) or lnum
  if not cmd_lnum then return end
  local first, last = children_of(cmd_lnum)
  local has_children = last ~= nil
  local on_cmd = cmd_lnum == lnum
  if has_children and (on_cmd or line_type(line) == "data") then
    if last then vim.cmd(string.format("%d,%ddelete _", first, last)) end
    toggle_bullet(cmd_lnum)
    return
  end
  local out = run(cmd_lnum)
  if not out then return end
  if last then vim.cmd(string.format("%d,%ddelete _", first, last)) end
  if out ~= "" then insert_output(cmd_lnum, out) end
  toggle_bullet(cmd_lnum)
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
