local M = {}

local function indent(text) return #text:match "^%s*" end
local function is_data(text) return text:find "^%s*[|:]" ~= nil end
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
  return table.concat(parts, "/")
end

local function exec()
  local ok, err = pcall(function()
    local line = vim.fn.getline "."

    -- Data line: | means collapse, : means re-execute with body
    if is_data(line) then
      local depth = indent(line)
      local i = vim.fn.line "." - 1
      while i >= 1 do
        local l = vim.fn.getline(i)
        if not is_data(l) and indent(l) < depth and l:find "%S" then
          vim.fn.cursor(i, 1)
          break
        end
        i = i - 1
      end

      local lnum = vim.fn.line "."
      local depth = indent(vim.fn.getline(lnum))
      local first_child = lnum + 1
      local body = {}
      local last_child = nil
      local i = first_child
      while i <= vim.fn.line "$" do
        local l = vim.fn.getline(i)
        if l == "" then break end
        local d = indent(l)
        if d <= depth then break end
        if d == depth + 2 then
          if l:find "^%s*: " then table.insert(body, vim.trim(l)) end
          last_child = i
        end
        i = i + 1
      end

      if line:find "^%s*|" then
        -- | line: collapse (delete children, toggle to +)
        if last_child then vim.cmd(string.format("%d,%ddelete _", first_child, last_child)) end
        vim.fn.setline(lnum, (vim.fn.getline(lnum):gsub("^%- ", "+ ", 1)))
        return
      end

      -- : line: re-execute with body (phase 2)
      local path = build_path()
      if path == "" then return end

      local cmd = "viye " .. vim.fn.shellescape(path)
      if #body > 0 then
        cmd = cmd .. " --"
        for _, b in ipairs(body) do
          cmd = cmd .. " " .. vim.fn.shellescape(b)
        end
      end

      if last_child then vim.cmd(string.format("%d,%ddelete _", first_child, last_child)) end

      local out = vim.fn.system(cmd)
      if vim.v.shell_error ~= 0 then
        vim.notify("viye: " .. (out:gsub("\n$", "")), vim.log.levels.ERROR)
        return
      end

      if out ~= "" then
        local pad = string.rep(" ", depth + 2)
        local lines = vim.split(out, "\n", { trimempty = true })
        for j, l in ipairs(lines) do lines[j] = pad .. l end
        vim.fn.append(lnum, lines)
      end
      return
    end

    -- Command line: toggle expand/collapse
    local lnum = vim.fn.line "."
    local depth = indent(vim.fn.getline(lnum))

    local first_child = lnum + 1
    local last_child = nil
    local i = first_child
    while i <= vim.fn.line "$" do
      local l = vim.fn.getline(i)
      if l == "" then break end
      local d = indent(l)
      if d <= depth then break end
      if d == depth + 2 then last_child = i end
      i = i + 1
    end

    if last_child then
      -- Expanded: collapse
      vim.cmd(string.format("%d,%ddelete _", first_child, last_child))
      vim.fn.setline(lnum, (vim.fn.getline(lnum):gsub("^%- ", "+ ", 1)))
      return
    end

    -- Collapsed: run CLI and insert output
    local path = build_path()
    if path == "" then return end

    local cmd = "viye " .. vim.fn.shellescape(path)
    local out = vim.fn.system(cmd)
    if vim.v.shell_error ~= 0 then
      vim.notify("viye: " .. (out:gsub("\n$", "")), vim.log.levels.ERROR)
      return
    end

    if out ~= "" then
      local pad = string.rep(" ", depth + 2)
      local lines = vim.split(out, "\n", { trimempty = true })
      for j, l in ipairs(lines) do lines[j] = pad .. l end
      vim.fn.append(lnum, lines)
      vim.fn.setline(lnum, (vim.fn.getline(lnum):gsub("^%+ ", "- ", 1)))
    else
      vim.fn.setline(lnum, (vim.fn.getline(lnum):gsub("^%- ", "+ ", 1)))
    end
  end)
  if not ok then
    vim.notify("viye: " .. tostring(err), vim.log.levels.ERROR)
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

if false then -- tests
  local ok, fail = 0, 0
  local function check(got, want, msg)
    if got == want then
      ok = ok + 1
    else
      fail = fail + 1
      print("FAIL " .. msg .. ": got " .. vim.inspect(got) .. " want " .. vim.inspect(want))
    end
  end

  check(indent "", 0, "indent empty")
  check(indent "foo", 0, "indent no indent")
  check(indent "  foo", 2, "indent 2 spaces")
  check(indent "    foo", 4, "indent 4 spaces")

  check(is_data ": hello", true, "is_data colon")
  check(is_data "  : hello", true, "is_data colon indent")
  check(is_data "| hello", true, "is_data pipe")
  check(is_data "  |", true, "is_data pipe alone")
  check(is_data "+ demo", false, "is_data not +")
  check(is_data "- demo", false, "is_data not -")
  check(is_data "hello", false, "is_data plain")

  check(strip_bullet "+ demo", "demo", "strip +")
  check(strip_bullet "- demo", "demo", "strip -")
  check(strip_bullet "demo", "demo", "strip none")
  check(strip_bullet "+ hello world", "hello world", "strip + multiword")

  if fail == 0 then
    print("viye.lua: " .. ok .. "/" .. (ok + fail) .. " tests passed")
  else
    print("viye.lua: " .. fail .. " failures out of " .. (ok + fail) .. " tests")
  end
end

return M
