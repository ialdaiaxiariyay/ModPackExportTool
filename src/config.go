// config.go
package pack

import (
    "os"
    "gopkg.in/yaml.v3"
)

type Config struct {
    GamePath string         `yaml:"game_path"`
    Export   ExportConfig   `yaml:"export"`
    Advanced AdvancedConfig `yaml:"advanced"`
    PCL      PCLConfig      `yaml:"pcl"`
    Git      GitConfig      `yaml:"git"`
    RulesOverrides   []string `yaml:"rules_overrides,omitempty"`
    ExtraFiles       []string `yaml:"extra_files,omitempty"`
}

type ExportConfig struct {
    Name    string        `yaml:"name"`
    Version string        `yaml:"version"`
    Options OptionsConfig `yaml:"options"`
}

type OptionsConfig struct {
    Basic          bool `yaml:"basic"`
    GameSettings   bool `yaml:"game_settings"`
    PersonalInfo   bool `yaml:"personal_info"`
    OptiFine       bool `yaml:"optifine_settings"`
    Mod            bool `yaml:"mod"`
    DisabledMods   bool `yaml:"disabled_mods"`
    ImportantData  bool `yaml:"important_data"`
    ModSettings    bool `yaml:"mod_settings"`
    TaczGuns       bool `yaml:"tacz_guns"`
    ImmersivePaint bool `yaml:"immersive_paintings"`
    DrawnMaps      bool `yaml:"drawn_maps"`
    JEIPersonal    bool `yaml:"jei_personal"`
    EMIPersonal    bool `yaml:"emi_personal"`
    Patchouli      bool `yaml:"patchouli_personal"`
    ResourcePacks  bool `yaml:"resource_packs"`
    ShaderPacks    bool `yaml:"shader_packs"`
    Screenshots    bool `yaml:"screenshots"`
    Schematics     bool `yaml:"schematics"`
    Replay         bool `yaml:"replay_recordings"`
    Saves          bool `yaml:"saves"`
    Licence        bool `yaml:"licence"`
    ServersDat     bool `yaml:"servers_dat"`
    Java           bool `yaml:"java"`
    PCL            bool `yaml:"pcl"`
    PCLCustom      bool `yaml:"pcl_custom"`
}

type AdvancedConfig struct {
    SkipNetwork    bool `yaml:"skip_network"`
    ModrinthOnly   bool `yaml:"modrinth_only"`
}

type PCLConfig struct {
    Executable string `yaml:"executable"`
}

type GitConfig struct {
    Init   bool   `yaml:"init"`
    Remote string `yaml:"remote"`
    Branch string `yaml:"branch"`
}

func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    var cfg Config
    err = yaml.Unmarshal(data, &cfg)
    if err != nil {
        return nil, err
    }
    if cfg.Git.Branch == "" {
        cfg.Git.Branch = "main"
    }
    return &cfg, nil
}