#!/bin/bash

_icw_complete() {
    local cur prev commands
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    commands="tag dumpdepend dd tree depend-ng update-ng release relocate status st commit ci add update depend wipe help"
    case "${prev}" in
        icw)
            COMPREPLY=( $(compgen -W "${commands}" -- ${cur}) )
            return 0
            ;;
        add) 
            COMPREPLY=( $(compgen -d) )
            return 0
            ;;
        *)
        if [[ "${COMP_WORDS[1]}" == "add" && ${#COMP_WORDS[@]} -eq 4 ]]; then
            OPTS="setup digital analog"
            COMPREPLY=( $(compgen -W "${OPTS}" -- ${cur}) )
            return 0
        fi

    esac
}

complete -F _icw_complete icw