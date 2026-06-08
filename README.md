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
  1  [ssh]   ssh aeneas64-ubuntu  aeneas64 Ubuntu VM
  2  [git]   git log --oneline --graph --all  Pretty git log
  3  [sys]   df -h  Disk usage human readable

Select (1-3, q to quit): _
```

Select a number and the command is placed in your readline buffer ready to edit — `{placeholder}` values and all.

## Storage

Commands are stored in `~/.config/favs/commands.json`. Edit it directly if you like.

## Key binding

The default binding is `Ctrl+F`. To change it, edit the snippet in your `.bashrc`:

```bash
bind -x '"\C-f": _favs_widget'  # change \C-f to your preferred key
```
