if exists("b:current_syntax")
  finish
endif

syn match viyeTool   /^[^|:=+ \t-]\S*/            " top-level command name
syn match viyeTool   /^\s*[+-] \zs\S\+\ze\/\@<!/  " bullet child
syn match viyeURL    /\v<https:\/\/\S+/           " urls
syn match viyeOutput /^\s*|.*$/                   " | output
syn match viyeBody   /^\s*:.*$/                   " : body

hi def link viyeTool Function
hi def link viyeURL Underlined
hi def link viyeOutput Comment
hi def link viyeBody String

let b:current_syntax = "viye"
