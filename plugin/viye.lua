local function indent(t) return #t:match("^%s*") end
local function is_data(line) return line:find("^%s*[|:]") ~= nil end
local function is_node(line) return line:find("%S") and not is_data(line) end

local function node_text(lnum)
  local l = vim.fn.getline(lnum)
  return l:match("^%s*[+-] (.*)") or vim.trim(l)
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
    if is_node(l) and indent(l) < d then return i end
  end
  return nil
end

local function ancestors(cmd_lnum)
  local parts = {}
  local lnum = cmd_lnum
  while lnum do
    table.insert(parts, 1, node_text(lnum))
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
  local parts = ancestors(cmd_lnum)
  if #parts == 0 then return nil end
  local args = vim.tbl_map(vim.fn.shellescape, parts)
  local body = body_of(cmd_lnum)
  if #body > 0 then
    table.insert(args, "--")
    vim.list_extend(args, vim.tbl_map(vim.fn.shellescape, body))
  end
  local cmd = "viye " .. table.concat(args, " ")
  vim.notify("viye: " .. cmd, vim.log.levels.DEBUG)
  local out = vim.fn.system(cmd)
  if vim.v.shell_error ~= 0 then
    vim.notify("viye: " .. vim.trim(out or ""), vim.log.levels.ERROR)
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
  local cmd_lnum = is_data(line) and parent_of(lnum) or lnum
  if not cmd_lnum then return end
  local first, last = children_of(cmd_lnum)
  local on_cmd = cmd_lnum == lnum
  local has_bullet = line:find("^%s*[+-] ") ~= nil
  local collapsing = last and ((on_cmd and has_bullet) or line:find("^%s*|") ~= nil)
  if collapsing then
    delete_range(first, last)
    toggle_bullet(cmd_lnum, "+")
    return
  end
  local out = run(cmd_lnum)
  if not out then return end
  delete_range(first, last)
  if out ~= "" then
    insert_output(cmd_lnum, out)
  end
  toggle_bullet(cmd_lnum, "-")
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
  end
})

vim.api.nvim_create_autocmd("FileType", {
  pattern = "viye",
  group = group,
  callback = function()
    vim.keymap.set("n", "<CR>", exec, { buffer = true, silent = true })
  end
})
