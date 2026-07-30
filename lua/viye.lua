local M = {}

local function indent(text) return #text:match("^%s*") end
local function is_data(line) return line:find("^%s*[|:]") ~= nil end

local function children_of(cmd_lnum)
  local depth = indent(vim.fn.getline(cmd_lnum))
  local first = cmd_lnum + 1
  local last = nil
  local i = first
  while i <= vim.fn.line("$") do
    local l = vim.fn.getline(i)
    if l == "" then break end
    if indent(l) <= depth then break end
    last = i
    i = i + 1
  end
  return first, last
end

local function parent_of(lnum)
  local d = indent(vim.fn.getline(lnum))
  while lnum >= 1 do
    local l = vim.fn.getline(lnum)
    if not is_data(l) and l:find("%S") and indent(l) < d then return lnum end
    lnum = lnum - 1
  end
  return nil
end

local function ancestors(cmd_lnum)
  local parts = {}
  local depth = math.huge
  local i = cmd_lnum
  while i >= 1 do
    local l = vim.fn.getline(i)
    if not is_data(l) and l:find("%S") then
      local d = indent(l)
      if d < depth then
        local text = l:match("^%s*[+-] (.*)")
        if not text then text = vim.trim(l) end
        table.insert(parts, 1, text)
        depth = d
        if d == 0 then break end
      end
    end
    i = i - 1
  end
  return parts
end

local function body_of(cmd_lnum)
  local depth = indent(vim.fn.getline(cmd_lnum))
  local body = {}
  local i = cmd_lnum + 1
  while i <= vim.fn.line("$") do
    local l = vim.fn.getline(i)
    if l == "" then break end
    local d = indent(l)
    if d <= depth then break end
    if d == depth + 2 then
      local val = l:match("^%s*: (.*)")
      if val then
        table.insert(body, ": " .. val)
      end
    end
    i = i + 1
  end
  return body
end

local function run(cmd_lnum)
  local parts = ancestors(cmd_lnum)
  if #parts == 0 then return nil end

  local cmd = "viye"
  for _, p in ipairs(parts) do
    cmd = cmd .. " " .. vim.fn.shellescape(p)
  end
  local body = body_of(cmd_lnum)
  if #body > 0 then
    cmd = cmd .. " --"
    for _, b in ipairs(body) do
      cmd = cmd .. " " .. vim.fn.shellescape(b)
    end
  end

  vim.notify("ex'ing: " .. cmd, vim.log.levels.INFO)
  local out = vim.fn.system(cmd)
  if vim.v.shell_error ~= 0 then
    vim.notify("viye: " .. (vim.trim(out or "")), vim.log.levels.ERROR)
    return nil
  end
  return out
end

local function delete_range(first, last)
  if last then vim.cmd(string.format("%d,%ddelete _", first, last)) end
end

local function insert_output(cmd_lnum, out)
  local pad = string.rep(" ", indent(vim.fn.getline(cmd_lnum)) + 2)
  local lines = vim.split(out, "\n", { trimempty = true })
  for i, l in ipairs(lines) do
    lines[i] = pad .. l
  end
  vim.fn.append(cmd_lnum, lines)
end

local function toggle_bullet(cmd_lnum, to)
  vim.fn.setline(cmd_lnum, (vim.fn.getline(cmd_lnum):gsub("^[+-] ", to .. " ", 1)))
end

local function exec()
  local lnum = vim.fn.line(".")
  local line = vim.fn.getline(lnum)

  -- find the enclosing command line
  local cmd_lnum
  if is_data(line) then
    cmd_lnum = parent_of(lnum)
  else
    cmd_lnum = lnum
  end
  if not cmd_lnum then return end

  local first, last = children_of(cmd_lnum)
  local on_cmd = cmd_lnum == lnum
  local has_bullet = line:find("^%s*[+-] ")

  if on_cmd and last and has_bullet then
    -- expanded togglable command -> collapse
    delete_range(first, last)
    toggle_bullet(cmd_lnum, "+")
    return
  end

  -- | output line: collapse without re-running
  if line:find("^%s*|") then
    if last then
      delete_range(first, last)
      toggle_bullet(cmd_lnum, "+")
    end
    return
  end

  -- collapsed command, or Enter on a child: run and (re-)expand
  local out = run(cmd_lnum)
  if not out then return end

  delete_range(first, last)
  if out ~= "" then
    insert_output(cmd_lnum, out)
  end
  toggle_bullet(cmd_lnum, "-")
end

function M.setup()
  vim.filetype.add({ extension = { viye = "viye" } })
  local group = vim.api.nvim_create_augroup("viye", { clear = true })
  vim.api.nvim_create_autocmd("FileType", {
    pattern = "viye",
    group = group,
    callback = function()
      vim.keymap.set("n", "<CR>", exec, { buffer = true, silent = true })
    end,
  })
  vim.api.nvim_create_autocmd("BufWinEnter", {
    group = group,
    callback = function()
      if vim.fn.expand("%") == "" and vim.bo.buftype == "" then
        vim.bo.filetype = "viye"
      end
    end,
  })
end

return M
