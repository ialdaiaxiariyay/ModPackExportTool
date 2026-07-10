## pack-export

### Primary Documentation (English)

pack-export is a command-line tool written in Go that exports a Minecraft game instance into a Modrinth-compatible pack format (.mrpack). It scans the game directory, selectively collects files based on configurable rules, queries Modrinth and CurseForge APIs to map local files to remote download URLs, and generates a `modrinth.index.json` manifest. Unmatched or unqueried files are copied directly into the `overrides` folder. The tool also supports optional Git repository initialisation and final packaging into a `.mrpack` archive.

[中文文档](README_zh.md)

#### Features

- Configurable file selection covering game core, settings, mods, configs, resource packs, shaders, saves, and many mod-specific data folders.
- Automatic detection of Minecraft version, Forge, Fabric, and NeoForge from the version JSON.
- Network queries to Modrinth (SHA1) and CurseForge (MurmurHash2 fingerprint) for remote download URLs.
- Skip-network mode for offline or direct packaging (warning: may violate redistribution licenses).
- Optional Git initialisation in the output directory with remote and branch settings.
- Extra file/directory copying to the archive root.
- Template configuration generation with full comments.

#### Usage

```bash
pack-export -config <config.yaml> [options]
```

| Flag | Description |
|------|-------------|
| `-config` | Path to YAML configuration file (required unless `-save-config` is used). |
| `-output` | Output directory path (defaults to `name_version`). |
| `-name` | Override the pack name from config. |
| `-version` | Override the pack version from config. |
| `-init-git` | Initialise Git repository in the output directory (overrides `git.init` in config). |
| `-package` | Package the exported directory into a `.mrpack` file alongside it. |
| `-save-config` | Generate a commented YAML configuration template and exit. |

Example:

```bash
pack-export -config mypack.yaml -output MyPack_1.0 -package
```

#### Configuration

The YAML configuration file controls the export behaviour. A full template can be generated using `-save-config`. Key sections:

- `game_path`: Absolute or relative path to the Minecraft instance directory (must be an isolated version folder or `.minecraft` root).
- `export`: Pack name, version, and fine-grained boolean options for each file category (`basic`, `mod`, `config`, `resource_packs`, `saves`, etc.).
- `advanced`: `skip_network` (direct copy all files) and `modrinth_only` (skip CurseForge queries).
- `pcl`: Path to PCL executable (if including the launcher).
- `git`: `init`, `remote`, and `branch` settings.
- `rules_overrides` / `extra_files`: Advanced include/exclude rules and extra files to place at archive root.

#### Build

```bash
go build -o pack-export main.go
```

#### Reference

This project references parts of the implementation from [Meloong-Git/PCL](https://github.com/Meloong-Git/PCL).