// version.go
package pack

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "strings"
)

type VersionInfo struct {
    Minecraft string
    Forge     string
    Fabric    string
    NeoForge  string
}

func DetectVersion(gamePath string) (*VersionInfo, error) {
    absPath, err := filepath.Abs(gamePath)
    if err != nil {
        return nil, fmt.Errorf("failed to get absolute path: %w", err)
    }
    gamePath = absPath

    versionName := filepath.Base(gamePath)
    info := &VersionInfo{}

    jsonFiles := []string{
        filepath.Join(gamePath, versionName+".json"),
        filepath.Join(gamePath, "version.json"),
    }
    var rawData []byte
    var found bool
    for _, path := range jsonFiles {
        if data, err := os.ReadFile(path); err == nil {
            rawData = data
            found = true
            break
        }
    }
    if !found {
        return nil, fmt.Errorf("version JSON not found, please specify minecraft_version in config")
    }

    var v struct {
        Id            string `json:"id"`
        InheritsFrom  string `json:"inheritsFrom"`
        ClientVersion string `json:"clientVersion"`
        Arguments     struct {
            Game []interface{} `json:"game"`
        } `json:"arguments"`
        Libraries []struct {
            Name string `json:"name"`
        } `json:"libraries"`
    }
    if err := json.Unmarshal(rawData, &v); err != nil {
        return nil, fmt.Errorf("failed to parse version JSON: %w", err)
    }

    mcVersion := ""
    if v.ClientVersion != "" && isVersionNumber(v.ClientVersion) {
        mcVersion = v.ClientVersion
    }
    if mcVersion == "" {
        for _, arg := range v.Arguments.Game {
            if str, ok := arg.(string); ok && str == "--fml.mcVersion" {
                idx := 0
                for i, a := range v.Arguments.Game {
                    if s, ok := a.(string); ok && s == "--fml.mcVersion" {
                        idx = i + 1
                        break
                    }
                }
                if idx > 0 && idx < len(v.Arguments.Game) {
                    if val, ok := v.Arguments.Game[idx].(string); ok && isVersionNumber(val) {
                        mcVersion = val
                        break
                    }
                }
            }
        }
    }
    if mcVersion == "" && v.InheritsFrom != "" && isVersionNumber(v.InheritsFrom) {
        mcVersion = v.InheritsFrom
    }
    if mcVersion == "" && isVersionNumber(v.Id) {
        mcVersion = v.Id
    }
    if mcVersion == "" {
        mcVersion = versionName
        if !isVersionNumber(mcVersion) {
            mcVersion = ""
        }
    }

    if mcVersion == "" {
        return nil, fmt.Errorf("could not parse Minecraft version from version JSON, please specify minecraft_version in config")
    }
    info.Minecraft = mcVersion

    // Forge
    for i, arg := range v.Arguments.Game {
        if str, ok := arg.(string); ok && str == "--fml.forgeVersion" {
            if i+1 < len(v.Arguments.Game) {
                if val, ok := v.Arguments.Game[i+1].(string); ok && val != "" {
                    info.Forge = val
                    break
                }
            }
        }
    }

    if info.Forge == "" {
        for _, lib := range v.Libraries {
            if strings.Contains(lib.Name, "net.minecraftforge:forge:") {
                parts := strings.Split(lib.Name, ":")
                if len(parts) >= 3 {
                    info.Forge = parts[2]
                    break
                }
            }
        }
    }

    // Fabric
    for _, lib := range v.Libraries {
        if strings.Contains(lib.Name, "net.fabricmc:fabric-loader:") {
            parts := strings.Split(lib.Name, ":")
            if len(parts) >= 3 {
                info.Fabric = parts[2]
                break
            }
        }
    }
    // NeoForge
    for _, lib := range v.Libraries {
        if strings.Contains(lib.Name, "net.neoforged:neoforge:") {
            parts := strings.Split(lib.Name, ":")
            if len(parts) >= 3 {
                info.NeoForge = parts[2]
                break
            }
        }
    }

    if info.Forge == "" && info.Fabric == "" && info.NeoForge == "" {
        if files, err := filepath.Glob(filepath.Join(gamePath, "forge-*.jar")); err == nil && len(files) > 0 {
            base := filepath.Base(files[0])
            ver := strings.TrimPrefix(base, "forge-")
            ver = strings.TrimSuffix(ver, filepath.Ext(ver))
            if idx := strings.Index(ver, "-"); idx > 0 {
                ver = ver[:idx]
            }
            if ver != "" && strings.Contains(ver, ".") {
                info.Forge = ver
            }
        } else if files, err := filepath.Glob(filepath.Join(gamePath, "neoforge-*.jar")); err == nil && len(files) > 0 {
            base := filepath.Base(files[0])
            ver := strings.TrimPrefix(base, "neoforge-")
            ver = strings.TrimSuffix(ver, filepath.Ext(ver))
            if idx := strings.Index(ver, "-"); idx > 0 {
                ver = ver[:idx]
            }
            if ver != "" && strings.Contains(ver, ".") {
                info.NeoForge = ver
            }
        } else if files, err := filepath.Glob(filepath.Join(gamePath, "fabric-loader-*.jar")); err == nil && len(files) > 0 {
            base := filepath.Base(files[0])
            ver := strings.TrimPrefix(base, "fabric-loader-")
            ver = strings.TrimSuffix(ver, filepath.Ext(ver))
            if ver != "" && strings.Contains(ver, ".") {
                info.Fabric = ver
            }
        }
    }

    return info, nil
}

func isVersionNumber(s string) bool {
    parts := strings.Split(s, ".")
    if len(parts) < 2 {
        return false
    }
    for _, p := range parts {
        if p == "" {
            return false
        }
        for _, c := range p {
            if c < '0' || c > '9' {
                return false
            }
        }
    }
    return true
}