#!/bin/bash

_icw_complete() {
    local cur prev commands
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    # Current Go commands (implemented)
    # Future-proof: add new commands here as they are implemented
    commands="update status st tree add version test list ls"

    # Legacy Perl commands (if still supported)
    # commands+=" tag dumpdepend dd depend-ng update-ng release relocate commit ci depend wipe help"

    case "${prev}" in
        icw)
            COMPREPLY=( $(compgen -W "${commands}" -- ${cur}) )
            return 0
            ;;
        add)
            # Complete with directories for first argument
            COMPREPLY=( $(compgen -d -- ${cur}) )
            return 0
            ;;
        *)
        # Second argument for 'add' command: component types
        if [[ "${COMP_WORDS[1]}" == "add" && ${#COMP_WORDS[@]} -eq 4 ]]; then
            OPTS="setup digital analog process"
            COMPREPLY=( $(compgen -W "${OPTS}" -- ${cur}) )
            return 0
        fi
    esac
}

complete -F _icw_complete icw