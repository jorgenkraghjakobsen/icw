#!/bin/bash

_icw_complete() {
    local cur prev commands
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    # Current Go commands (implemented)
    commands="update status st tree hdl add version test list ls migrate auth completion help"

    # Note: 'migrate' command requires MAW backend (only works on g9 server)

    # Global flags (available for all commands)
    local global_flags="-h --help -v --version"

    # Command-specific flags
    local list_flags="-t --type -b --branches -g --tags -a --all -r --repo"
    local migrate_flags="--create-repo --from --to --add-user --dry-run"

    # Get the main command (first word after icw)
    local command=""
    for ((i=1; i < COMP_CWORD; i++)); do
        if [[ "${COMP_WORDS[i]}" != -* ]]; then
            command="${COMP_WORDS[i]}"
            break
        fi
    done

    # Complete main command
    if [[ $COMP_CWORD -eq 1 ]]; then
        COMPREPLY=( $(compgen -W "${commands}" -- ${cur}) )
        return 0
    fi

    # Flag value completion based on previous word
    case "${prev}" in
        -t|--type)
            # Component types for list command
            COMPREPLY=( $(compgen -W "analog digital setup process" -- ${cur}) )
            return 0
            ;;
        -r|--repo|--from|--to|--create-repo)
            # Repository names - could be enhanced to list actual repos
            # For now, just return to let user type freely
            return 0
            ;;
        --add-user)
            # Usernames - could be enhanced to list actual users
            # For now, just return to let user type freely
            return 0
            ;;
    esac

    # Command-specific completion
    case "${command}" in
        list|ls)
            if [[ ${cur} == -* ]]; then
                COMPREPLY=( $(compgen -W "${list_flags} ${global_flags}" -- ${cur}) )
            else
                # Complete with component paths or patterns
                # For now, let default filename completion handle this
                COMPREPLY=()
            fi
            return 0
            ;;
        migrate)
            if [[ ${cur} == -* ]]; then
                COMPREPLY=( $(compgen -W "${migrate_flags} ${global_flags}" -- ${cur}) )
            fi
            return 0
            ;;
        add)
            # First argument: directory completion
            if [[ $COMP_CWORD -eq 2 ]]; then
                COMPREPLY=( $(compgen -d -- ${cur}) )
                return 0
            fi
            # Second argument: component types
            if [[ $COMP_CWORD -eq 3 ]]; then
                COMPREPLY=( $(compgen -W "setup digital analog process" -- ${cur}) )
                return 0
            fi
            return 0
            ;;
        update|status|st|tree|hdl|test|version)
            # These commands only have global flags
            if [[ ${cur} == -* ]]; then
                COMPREPLY=( $(compgen -W "${global_flags}" -- ${cur}) )
            fi
            return 0
            ;;
        completion)
            # Completion subcommand for shell completion generation
            if [[ ${cur} == -* ]]; then
                COMPREPLY=( $(compgen -W "${global_flags}" -- ${cur}) )
            else
                COMPREPLY=( $(compgen -W "bash zsh fish powershell" -- ${cur}) )
            fi
            return 0
            ;;
        auth)
            # Auth subcommands
            if [[ ${cur} == -* ]]; then
                COMPREPLY=( $(compgen -W "${global_flags}" -- ${cur}) )
            else
                COMPREPLY=( $(compgen -W "login logout status test" -- ${cur}) )
            fi
            return 0
            ;;
        help)
            # Help can take any command as argument
            COMPREPLY=( $(compgen -W "${commands}" -- ${cur}) )
            return 0
            ;;
    esac

    # Default: if starts with -, show global flags
    if [[ ${cur} == -* ]]; then
        COMPREPLY=( $(compgen -W "${global_flags}" -- ${cur}) )
        return 0
    fi
}

complete -F _icw_complete icw