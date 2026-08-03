# favs

Bookmark your favorite shell commands and recall them into the readline buffer for editing — without executing them.

Similar in spirit to [navi](https://github.com/denisidoro/navi)'s snippets feature.

## Install

```bash
git clone https://github.com/wstevenson1/favs.git
cd favs
go build -o ~/.local/bin/favs .
```

Then set up the bash shell widget:

```bash
favs init >> ~/.bashrc && source ~/.bashrc
```

This binds `Ctrl+F` to the picker. Press it at any prompt to browse your saved commands.

## Usage

```bash
# Add a command
favs add "kubectl get pods -n {namespace}" --tags k8s --desc "List pods in a namespace"

# List all saved commands
favs list

# Filter by tag
favs list --tag ssh

# Remove by ID
favs rm 3

# Pick interactively (or press Ctrl+F)
favs
```

## Picker

```
favs
```

Type to filter (no need to press `/` first), use the arrow keys (or `j`/`k`) to move the highlight, and press Enter to select. The chosen command is placed in your readline buffer ready to edit — `{placeholder}` values and all. `Esc` or `Ctrl+C` quits without selecting anything.

## Storage

Commands are stored in `~/.config/favs/commands.toml`. Edit it directly if you like.

## Key binding

The default binding is `Ctrl+F`. To change it, edit the snippet in your `.bashrc`:

```bash
bind -x '"\C-f": _favs_widget'  # change \C-f to your preferred key
```
