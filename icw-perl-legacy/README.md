# ICW Perl Legacy

This directory contains the original Perl implementation of ICW for fallback purposes.

## Contents

- **icw.pl** - Original Perl ICW script (version 127)
- **icw_release.pl** - Original Perl ICW release script
- **completions/** - Bash completion scripts

## Usage

If you need to use the old Perl version:

```bash
# Run directly from this directory
./icw.pl --help

# Or create a symlink
ln -s $(pwd)/icw.pl ~/bin/icw-legacy
icw-legacy --help
```

## Notes

- This version is preserved from commit `d1d0adc` before the Go migration
- The Perl version requires:
  - Perl 5.x
  - Subversion client (`/usr/bin/svn`)
  - Perl modules: LWP::UserAgent, Term::ANSIColor, Getopt::Std, FileHandle, Cwd, URI::Escape

## Migration

The current ICW is being rewritten in Go for:
- Better performance
- Type safety
- Single binary distribution
- Modern tooling support
- Support for both Git (software tools) and SVN (design components)

See the main README.md for the Go version documentation.
