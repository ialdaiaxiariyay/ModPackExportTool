// collector.go
package pack

import (
    "io/fs"
    "path/filepath"
)

var optionRules = map[string]string{
    "GameSettings":   "options.txt|configureddefaults/",
    "PersonalInfo":   "hotbar.nbt|command_history.txt",
    "OptiFine":       "optionsof.txt|optionsshaders.txt",
    "Mod":            "mods/|!mods/*.disabled|!mods/*.old|!mods/.connector/|coremods/|lib/",
    "DisabledMods":   "mods/*.disabled|mods/*.old",
    "ImportantData":  "hotai/|bansoukou/|addons/|multiblocked/|modpack-update-checker/|global_packs/|global_resource_packs/|global_data_packs/|optional_data_packs/|moonlight-global-datapacks/|maps/|icon.png|mods-resourcepacks/|matmos/|resource_assorts/|resource_assorts.json|patchouli_books/|datapacks/|kubejs*/|!kubejs*/probe/|!kubejs*/exported/|!kubejs*/jsconfig.json|!kubejs*/README.txt|openloader/|worldshape/|resources/|scripts/|structures/|fontfiles/|oresources/|packmenu/|craftpresence/|pointblanks/|template*/|!template*/playerdata/|!template*/stats/",
    "ModSettings":    "config/|!config/jei/world/|!config/worldedit/|config/worldedit/worldedit.properties|!config/spark/|config/spark/config.json|defaultconfigs/|journeymap/config/|journeymap/server/|TrashSlotSaveState.json|customfov.txt|gg.essential.mod/|essential/|!essential/*/|!essential/*.jar*|!essential/screenshot-checksum-caches.json|!essential/microsoft_accounts.json|paragliderSettings.nbt|local/client_config.json|local/ftbl.json|local/client/sidebar_buttons.json|local/client/ftbutilities.cfg|local/client/ftblib.cfg|local/client/xencraft.cfg|liteloader.properties|default_reference.xml|CustomSkinLoader/CustomSkinLoader.json|!config/tacz/custom",
    "TaczGuns":       "tacz/|config/tacz/custom/",
    "ImmersivePaint": "immersive_paintings/",
    "DrawnMaps":      "journeymap/data/|xaero/|XaeroWaypoints/|XaeroWorldMap/",
    "JEIPersonal":    "config/jei/world/",
    "EMIPersonal":    "emi.json",
    "Patchouli":      "patchouli_data.json",
    "ResourcePacks":  "resourcepacks/|texturepacks/",
    "ShaderPacks":    "shaderpacks/",
    "Screenshots":    "screenshots/",
    "Schematics":     "schematics/",
    "Replay":         "replay_recordings/|replay_videos/",
    "Saves":          "saves/",
    "Licence":        "LICEN*",
    "ServersDat":     "servers.dat",
    "Java":           "runtime/|jre/",
}

func CollectFiles(gamePath string, opts OptionsConfig) ([]string, error) {
    var files []string
    err := filepath.WalkDir(gamePath, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }
        if d.IsDir() {
            return nil
        }
        rel, err := filepath.Rel(gamePath, path)
        if err != nil {
            return err
        }
        relUnix := filepath.ToSlash(rel)
        if shouldInclude(relUnix, opts) {
            files = append(files, relUnix)
        }
        return nil
    })
    return files, err
}

func shouldInclude(rel string, opts OptionsConfig) bool {
    var allIncludes []string
    var allExcludes []string

    if opts.Basic {
        allIncludes = append(allIncludes, "versions/")
        allIncludes = append(allIncludes, "launcher_profiles.json")
        allIncludes = append(allIncludes, "options.txt")
    }

    optionMap := map[string]bool{
        "GameSettings":   opts.GameSettings,
        "PersonalInfo":   opts.PersonalInfo,
        "OptiFine":       opts.OptiFine,
        "Mod":            opts.Mod,
        "DisabledMods":   opts.DisabledMods,
        "ImportantData":  opts.ImportantData,
        "ModSettings":    opts.ModSettings,
        "TaczGuns":       opts.TaczGuns,
        "ImmersivePaint": opts.ImmersivePaint,
        "DrawnMaps":      opts.DrawnMaps,
        "JEIPersonal":    opts.JEIPersonal,
        "EMIPersonal":    opts.EMIPersonal,
        "Patchouli":      opts.Patchouli,
        "ResourcePacks":  opts.ResourcePacks,
        "ShaderPacks":    opts.ShaderPacks,
        "Screenshots":    opts.Screenshots,
        "Schematics":     opts.Schematics,
        "Replay":         opts.Replay,
        "Saves":          opts.Saves,
        "Licence":        opts.Licence,
        "ServersDat":     opts.ServersDat,
        "Java":           opts.Java,
    }

    if !opts.Mod {
        optionMap["DisabledMods"] = false
        optionMap["ImportantData"] = false
        optionMap["ModSettings"] = false
        optionMap["TaczGuns"] = false
        optionMap["ImmersivePaint"] = false
        optionMap["DrawnMaps"] = false
        optionMap["JEIPersonal"] = false
        optionMap["EMIPersonal"] = false
        optionMap["Patchouli"] = false
    }

    for name, enabled := range optionMap {
        if !enabled {
            continue
        }
        rules, ok := optionRules[name]
        if !ok || rules == "" {
            continue
        }
        rs := ParseRules(rules)
        allIncludes = append(allIncludes, rs.Includes...)
        allExcludes = append(allExcludes, rs.Excludes...)
    }

    combined := RuleSet{Includes: allIncludes, Excludes: allExcludes}
    return combined.Matches(rel)
}