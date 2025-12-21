# Tasks

## TODO

- [ ] Review the spec doc and confirm scope (UX + behavior) before implementing

- [ ] Spec: define docmgr skill create UX (args/flags, ticket scoping, workspace vs ticket skill location, WhatFor/WhenToUse handling)
  # skill - List and show skills                                              
                                                                              
  Manage skills documentation. Skills are documents with DocType=skill that   
  provide structured information about what a skill is for and when to use it.
                                                                              
  For more help, run: docmgr help skill                                       
                                                                              
  Run docmgr help --ui to open the interactive help TUI.                      
                                                                              
  ## Usage:                                                                   
                                                                              
  docmgr skill [command]                                                      
                                                                              
  ## Available Commands:                                                      
                                                                              
  • **list**        List skills                                               
  • **show**        Show detailed information about a skill                   
                                                                              
  ## Flags:                                                                   
                                                                              
        -h, --help    help for skill                                          
       --long-help    Show long help                                          
                                                                              
  Use docmgr skill [command] --help for more information about a command. Use 
  docmgr skill --help --long-help for information about all flags.             UX (args/flags, ticket scoping, workspace vs ticket skill location, WhatFor/WhenToUse handling)
- [ ] Implement docmgr skill create command (cobra wiring + core logic)
  # skill - List and show skills                                              
                                                                              
  Manage skills documentation. Skills are documents with DocType=skill that   
  provide structured information about what a skill is for and when to use it.
                                                                              
  For more help, run: docmgr help skill                                       
                                                                              
  Run docmgr help --ui to open the interactive help TUI.                      
                                                                              
  ## Usage:                                                                   
                                                                              
  docmgr skill [command]                                                      
                                                                              
  ## Available Commands:                                                      
                                                                              
  • **list**        List skills                                               
  • **show**        Show detailed information about a skill                   
                                                                              
  ## Flags:                                                                   
                                                                              
        -h, --help    help for skill                                          
       --long-help    Show long help                                          
                                                                              
  Use docmgr skill [command] --help for more information about a command. Use 
  docmgr skill --help --long-help for information about all flags.             command (wiring + core logic)
- [ ] Add scenario coverage for skill create and active-ticket filtering interactions
- [ ] Update docs: docmgr-how-to-use + using-skills to include skill create
- [ ] Add completion/help examples; ensure output includes copy/pasteable next steps (show/load) (and prints created path)
