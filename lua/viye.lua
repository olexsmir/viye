local M = {}

local function indent(text) return #text:match "^%s*" end
local function is_data(text) return text:find "^%s*[|:] " ~= nil end
local function strip_bullet(text) return (text:gsub("^[+-] ", "", 1)) end

local function build_path()
  local lnum = vim.fn.line "."
  local line = vim.fn.getline(lnum)
  local depth = indent(line)
  local parts = { strip_bullet(vim.trim(line)) }

  local i = lnum - 1
  while i >= 1 do
    local l = vim.fn.getline(i)
    local d = indent(l)
    if d == 0 and l:find "%S" and depth > 0 then
      table.insert(parts, 1, strip_bullet(vim.trim(l)))
      break
    end
    if d < depth and l:find "%S" then
      table.insert(parts, 1, strip_bullet(vim.trim(l)))
      depth = d
    end
    i = i - 1
  end

  for idx = 1, #parts - 1 do
    parts[idx] = parts[idx]:gsub("/+$", "")
  end

  return table.concat(parts, "/")
end

local function has_children()
  local lnum = vim.fn.line "."
  local depth = indent(vim.fn.getline(lnum))
  local next_line = vim.fn.getline(lnum + 1)
  return next_line ~= "" and indent(next_line) > depth
end

local function collapse()
  local lnum = vim.fn.line "."
  local depth = indent(vim.fn.getline(lnum))
  local s = lnum + 1
  local e = s
  while e <= vim.fn.line "$" do
    local l = vim.fn.getline(e)
    if l == "" or indent(l) <= depth then break end
    e = e + 1
  end
  if e > s then vim.cmd(string.format("%d,%ddelete _", s, e - 1)) end
end

local function toggle_bullet()
  local lnum = vim.fn.line "."
  local line = vim.fn.getline(lnum)
  local toggled = line:gsub("^%+ ", "- ", 1)
  if toggled ~= line then
    vim.fn.setline(lnum, toggled)
    return
  end
  toggled = line:gsub("^%- ", "+ ", 1)
  if toggled ~= line then vim.fn.setline(lnum, toggled) end
end

local function insert_output(output)
  local lnum = vim.fn.line "."
  local pad = string.rep(" ", indent(vim.fn.getline(lnum)) + 2)
  local lines = vim.split(output, "\n", { trimempty = true })
  for i, line in ipairs(lines) do lines[i] = pad .. line end
  vim.fn.append(lnum, lines)
end

local function exec()
  local line = vim.fn.getline "."
  if is_data(line) then
    local lnum = vim.fn.line "."
    local depth = indent(line)
    local i = lnum - 1
    while i >= 1 do
      local l = vim.fn.getline(i)
      if not is_data(l) and indent(l) < depth and l:find "%S" then
        vim.fn.cursor(i, 1)
        exec()
        return
      end
      i = i - 1
    end
    return
  end

  if has_children() then
    collapse()
    toggle_bullet()
    return
  end

  local path = build_path()
  if path == "" then return end

  local cmd = "viye " .. vim.fn.shellescape(path)
  vim.notify("viye: " .. cmd, vim.log.levels.INFO)

  local out = vim.fn.system(cmd)
  if vim.v.shell_error ~= 0 then
    vim.notify("viye: " .. out:gsub("\n$", ""), vim.log.levels.ERROR)
    return
  end

  if out ~= "" then
    insert_output(out)
    toggle_bullet()
  end
end

function M.setup()
  vim.filetype.add { extension = { viye = "viye" } }
  vim.api.nvim_create_autocmd("FileType", {
    pattern = "viye",
    group = vim.api.nvim_create_augroup("viye_plugin", { clear = true }),
    callback = function() vim.keymap.set("n", "<CR>", exec, { buffer = true, silent = true }) end,
  })
end

return M
