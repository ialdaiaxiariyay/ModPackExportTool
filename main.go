// main.go
package main

import (
    "flag"
    "fmt"
    "os"
    "pack-export/src"
    "strings"
)

func main() {
    configPath := flag.String("config", "", "Configuration file path (YAML)")
    output := flag.String("output", "", "Output directory (default: name_version)")
    name := flag.String("name", "", "Pack name (overrides config)")
    version := flag.String("version", "", "Pack version (overrides config)")
    initGit := flag.Bool("init-git", false, "Initialize Git if output directory is not a repo (overrides config git.init)")
    saveConfig := flag.String("save-config", "", "Save current config to specified file with template comments and exit")
    packageFlag := flag.Bool("package", false, "Package the exported directory as .mrpack after export")
    flag.Parse()

    if *configPath == "" && *saveConfig == "" {
        fmt.Println("Please specify a config file with -config")
        flag.Usage()
        os.Exit(1)
    }

    var cfg *pack.Config
    var err error
    if *configPath != "" {
        cfg, err = pack.LoadConfig(*configPath)
        if err != nil {
            fmt.Printf("Failed to load config: %v\n", err)
            os.Exit(1)
        }
    } else {
        cfg = &pack.Config{
            GamePath: ".",
            Export: pack.ExportConfig{
                Name:    "MyPack",
                Version: "1.0",
                Options: pack.OptionsConfig{
                    Basic:          true,
                    GameSettings:   true,
                    OptiFine:       false,
                    Mod:            true,
                    ImportantData:  true,
                    ModSettings:    true,
                    ResourcePacks:  true,
                    ShaderPacks:    true,
                    Licence:        true,
                    Java:           false,
                    PCL:            false,
                    PCLCustom:      false,
                },
            },
            Advanced: pack.AdvancedConfig{
                SkipNetwork:  false,
                ModrinthOnly: false,
            },
            PCL: pack.PCLConfig{
                Executable: "",
            },
            Git: pack.GitConfig{
                Init:   false,
                Remote: "",
                Branch: "main",
            },
        }
    }

    if *name != "" {
        cfg.Export.Name = *name
    }
    if *version != "" {
        cfg.Export.Version = *version
    }

    if *saveConfig != "" {
        configTemplate := `# Game instance path (must point to a specific isolated version directory)
# Examples:
#   - Official launcher isolated: C:/Users/You/.minecraft/versions/1.18.2/
#   - MultiMC instance: /home/you/MultiMC/instances/MyPack/
#   - Plain .minecraft root (not isolated): C:/Users/You/.minecraft/
game_path: "%s"

export:
  # Pack name
  name: "%s"
  # Pack version
  version: "%s"
  options:
    # Game core (required)
    basic: %v
    # Game settings (keybindings, volume, video, etc.)
    game_settings: %v
    # Personal game data (command history, saved hotbars)
    personal_info: %v
    # OptiFine settings (requires OptiFine)
    optifine_settings: %v
    # Mods
    mod: %v
    # Disabled mods (files with .disabled .old)
    disabled_mods: %v
    # Important pack data (scripts, built-in resourcepacks, datapacks, etc.)
    important_data: %v
    # Mod settings (config folder)
    mod_settings: %v
    # TaCZ gun packs
    tacz_guns: %v
    # Uploaded immersive paintings
    immersive_paintings: %v
    # Drawn maps (map mod records)
    drawn_maps: %v
    # JEI personal data (favorites, etc.)
    jei_personal: %v
    # EMI personal data (favorites, recipe history, etc.)
    emi_personal: %v
    # Patchouli personal data (read status, bookmarks, etc.)
    patchouli_personal: %v
    # Resource packs (texture packs)
    resource_packs: %v
    # Shader packs
    shader_packs: %v
    # Screenshots
    screenshots: %v
    # Exported schematics (schematics folder)
    schematics: %v
    # Replay recordings (Replay Mod files)
    replay_recordings: %v
    # Singleplayer saves (worlds/maps)
    saves: %v
    # Licence file
    licence: %v
    # Multiplayer server list
    servers_dat: %v
    # Java runtime in version folder
    java: %v
    # PCL launcher program (packaging official PCL)
    pcl: %v
    # PCL custom content (hidden features, homepage, etc.)
    pcl_custom: %v

# Advanced options
advanced:
  # Skip network checks and package all files directly (equivalent to DontCheckHostedAssets)
  # Note: If enabled, all files (including mods) will be bundled into the pack,
  # redistribution may violate licenses of some mods, use with caution.
  skip_network: %v
  # Only query Modrinth for download links, skip CurseForge
  modrinth_only: %v

# PCL launcher configuration (if pcl is true)
pcl:
  # PCL executable path, leave empty to auto‑discover PCL.exe in current directory
  executable: "%s"

# Git helper configuration (only initialises repo, does not auto‑commit)
git:
  # Auto‑initialise Git repository if directory is not a repo
  init: %v
  # Remote repository URL (only recorded, not pushed)
  remote: "%s"
  # Branch name
  branch: "%s"

# Override rules and extra file list (for importing advanced config from external files)
# Leave empty or delete if unused
# rules_overrides:
#   - "# Modify rules below to control exported content."
#   - "# Prefix with ! to negate. Supports *, ?, [] wildcards. Later lines override earlier ones."
#   - ""
#   - "mods/"
#   - "!mods/*.disabled"
#   - ...
# extra_files:
#   - "# To add extra files to the root of the archive, write their full paths below."
#   - "# Must be absolute paths. Lines ending with \\ or / denote directories, otherwise files."
rules_overrides: []
extra_files: []
`
        content := fmt.Sprintf(configTemplate,
            cfg.GamePath,
            cfg.Export.Name,
            cfg.Export.Version,
            cfg.Export.Options.Basic,
            cfg.Export.Options.GameSettings,
            cfg.Export.Options.PersonalInfo,
            cfg.Export.Options.OptiFine,
            cfg.Export.Options.Mod,
            cfg.Export.Options.DisabledMods,
            cfg.Export.Options.ImportantData,
            cfg.Export.Options.ModSettings,
            cfg.Export.Options.TaczGuns,
            cfg.Export.Options.ImmersivePaint,
            cfg.Export.Options.DrawnMaps,
            cfg.Export.Options.JEIPersonal,
            cfg.Export.Options.EMIPersonal,
            cfg.Export.Options.Patchouli,
            cfg.Export.Options.ResourcePacks,
            cfg.Export.Options.ShaderPacks,
            cfg.Export.Options.Screenshots,
            cfg.Export.Options.Schematics,
            cfg.Export.Options.Replay,
            cfg.Export.Options.Saves,
            cfg.Export.Options.Licence,
            cfg.Export.Options.ServersDat,
            cfg.Export.Options.Java,
            cfg.Export.Options.PCL,
            cfg.Export.Options.PCLCustom,
            cfg.Advanced.SkipNetwork,
            cfg.Advanced.ModrinthOnly,
            cfg.PCL.Executable,
            cfg.Git.Init,
            cfg.Git.Remote,
            cfg.Git.Branch,
        )
        if err := os.WriteFile(*saveConfig, []byte(content), 0644); err != nil {
            fmt.Printf("Failed to save config: %v\n", err)
            os.Exit(1)
        }
        fmt.Printf("Configuration template saved to %s\n", *saveConfig)
        return
    }

    if cfg.GamePath == "" {
        fmt.Println("game_path is empty in config")
        os.Exit(1)
    }
    if _, err := os.Stat(cfg.GamePath); err != nil {
        fmt.Printf("Game path does not exist: %v\n", err)
        os.Exit(1)
    }

    if *initGit {
        cfg.Git.Init = true
    }

    outputDir := *output
    if outputDir == "" {
        outputDir = cfg.Export.Name + "_" + cfg.Export.Version
    }

    // Ensure output directory does not end with archive extension, as we treat it as a directory
    if strings.HasSuffix(outputDir, ".mrpack") || strings.HasSuffix(outputDir, ".zip") {
        fmt.Println("Warning: -output should not end with .mrpack or .zip, it will be treated as a directory path")
    }

    // Perform export
    if err := pack.Export(cfg, outputDir, cfg.Git.Init); err != nil {
        fmt.Printf("Export failed: %v\n", err)
        os.Exit(1)
    }

    // If packaging is enabled
    if *packageFlag {
        archivePath := outputDir + ".mrpack"
        fmt.Printf("Packaging to %s ...\n", archivePath)
        if err := pack.CreateArchive(outputDir, archivePath); err != nil {
            fmt.Printf("Warning: packaging failed: %v\n", err)
        } else {
            fmt.Printf("Pack packaged to %s\n", archivePath)
        }
    }
}