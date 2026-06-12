local M = {}

local secnotes_password = nil

local function get_password()
    if secnotes_password == nil or secnotes_password == "" then
        secnotes_password = vim.fn.inputsecret("SecNotes Master Password: ")
    end
    return secnotes_password
end

local function get_note_id(file_path)
    -- Extract filename without extension as note ID
    local filename = vim.fn.fnamemodify(file_path, ":t")
    return vim.fn.fnamemodify(filename, ":r")
end

local function goto_secnote()
    local line = vim.api.nvim_get_current_line()
    local col = vim.api.nvim_win_get_cursor(0)[2]
    
    local i = col + 1
    local start_idx = nil
    local end_idx = nil
    
    for j = i, 1, -1 do
        if string.sub(line, j, j+1) == "[[" then
            start_idx = j + 2
            break
        end
    end
    
    if start_idx then
        for j = i, string.len(line) do
            if string.sub(line, j, j+1) == "]]" then
                end_idx = j - 1
                break
            end
        end
    end
    
    if start_idx and end_idx and start_idx <= end_idx then
        local target = string.sub(line, start_idx, end_idx)
        vim.cmd("e " .. target .. ".secnote")
    else
        vim.notify("No SecNotes link [[...]] under cursor", vim.log.levels.WARN)
    end
end

vim.api.nvim_create_autocmd("BufReadCmd", {
    pattern = "*.secnote",
    callback = function(args)
        local file = args.file
        local note_id = get_note_id(file)
        local password = get_password()
        
        if password == "" then
            vim.notify("SecNotes: Password required to read note", vim.log.levels.ERROR)
            return
        end

        local obj = vim.system({'secnotes', 'read', note_id}, {
            env = { SECNOTES_PASSWORD = password },
            text = true,
        }):wait()

        if obj.code ~= 0 then
            -- Note might not exist yet, which is fine for a new buffer
            if string.find(obj.stderr or "", "no such file") then
                vim.notify("SecNotes: New note " .. note_id, vim.log.levels.INFO)
            else
                vim.notify("SecNotes: Failed to decrypt note. " .. (obj.stderr or ""), vim.log.levels.ERROR)
                return
            end
        else
            -- Load output into the buffer
            local lines = vim.split(obj.stdout or "", "\n")
            vim.api.nvim_buf_set_lines(args.buf, 0, -1, false, lines)
        end
        
        vim.bo[args.buf].modified = false
        vim.bo[args.buf].buftype = "acwrite" -- Allow saving via BufWriteCmd
        
        -- Map 'gf' to jump to links inside this buffer
        vim.keymap.set('n', 'gf', goto_secnote, { buffer = args.buf, silent = true, desc = "Go to SecNotes link" })
    end
})

vim.api.nvim_create_autocmd("BufWriteCmd", {
    pattern = "*.secnote",
    callback = function(args)
        local file = args.file
        local note_id = get_note_id(file)
        local password = get_password()
        
        if password == "" then
            vim.notify("SecNotes: Password required to write note", vim.log.levels.ERROR)
            return
        end

        local lines = vim.api.nvim_buf_get_lines(args.buf, 0, -1, false)
        local content = table.concat(lines, "\n")
        
        local obj = vim.system({'secnotes', 'write', note_id}, {
            env = { SECNOTES_PASSWORD = password },
            stdin = content,
            text = true,
        }):wait()
        
        if obj.code ~= 0 then
            vim.notify("SecNotes: Failed to encrypt note. " .. (obj.stderr or ""), vim.log.levels.ERROR)
            return
        end
        
        vim.bo[args.buf].modified = false
        vim.notify("SecNotes: Note saved locally.", vim.log.levels.INFO)

        -- Trigger background sync
        vim.system({'secnotes', 'sync'}, {
            text = true,
        }, function(sync_obj)
            vim.schedule(function()
                if sync_obj.code == 0 then
                    vim.notify("SecNotes: Background sync complete.", vim.log.levels.INFO)
                else
                    vim.notify("SecNotes: Background sync failed.", vim.log.levels.ERROR)
                end
            end)
        end)
    end
})

return M
