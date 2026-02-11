# aptly Agent Notes

## Source Code Fallback

When source code is unavailable from:
- `aptly account source-code <address> [module_name] [--package <name>] [--raw]`

Direct users (and use yourself) to decompile:
- `aptly decompile address <address>`
- `aptly decompile module <address> <module_name>`

Rule: if source metadata is missing, use `decompile`.
