// manifest.go
package pack

import (
    "crypto/sha512"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "sort"
    "strings"
)

type Dependencies struct {
    Minecraft string `json:"minecraft"`
    Forge     string `json:"forge,omitempty"`
    Fabric    string `json:"fabric-loader,omitempty"`
    NeoForge  string `json:"neoforge,omitempty"`
}

type Manifest struct {
    Game         string         `json:"game"`
    FormatVersion int           `json:"formatVersion"`
    VersionId    string         `json:"versionId"`
    Name         string         `json:"name"`
    Summary      string         `json:"summary"`
    Files        []ManifestFile `json:"files"`
    Dependencies Dependencies   `json:"dependencies"`
}

type ManifestFile struct {
    Path      string            `json:"path"`
    Hashes    map[string]string `json:"hashes"`
    Downloads []string          `json:"downloads"`
    FileSize  int64             `json:"fileSize"`
}

func GenerateManifest(outputDir, name, version string, versionInfo *VersionInfo, resources []*LocalResourceFile, downloadMap map[string]string) error {
    deps := Dependencies{
        Minecraft: versionInfo.Minecraft,
    }
    if versionInfo.Forge != "" {
        deps.Forge = versionInfo.Forge
    }
    if versionInfo.Fabric != "" {
        deps.Fabric = versionInfo.Fabric
    }
    if versionInfo.NeoForge != "" {
        deps.NeoForge = versionInfo.NeoForge
    }

    manifest := Manifest{
        Game:          "minecraft",
        FormatVersion: 1,
        VersionId:     version,
        Name:          name,
        Summary:       fmt.Sprintf("Exported pack %s %s", name, version),
        Dependencies:  deps,
    }

    var files []ManifestFile
    for _, res := range resources {
        var downloads []string
        if url, ok := downloadMap[res.ModrinthHash]; ok {
            downloads = append(downloads, url)
        }
        if url, ok := downloadMap[fmt.Sprintf("cf_%d", res.CurseForgeHash)]; ok {
            downloads = append(downloads, url)
        }
        if len(downloads) == 0 {
            continue
        }

        info, err := os.Stat(res.FullPath)
        if err != nil {
            return err
        }

        data, err := os.ReadFile(res.FullPath)
        if err != nil {
            return err
        }
        sha512Sum := sha512.Sum512(data)
        sha512Str := hex.EncodeToString(sha512Sum[:])

        sort.Slice(downloads, func(i, j int) bool {
            iIsModrinth := strings.Contains(downloads[i], "modrinth.com")
            jIsModrinth := strings.Contains(downloads[j], "modrinth.com")
            if iIsModrinth && !jIsModrinth {
                return false
            }
            if !iIsModrinth && jIsModrinth {
                return true
            }
            return downloads[i] < downloads[j]
        })

        files = append(files, ManifestFile{
            Path:      res.Path,
            Hashes:    map[string]string{"sha1": res.ModrinthHash, "sha512": sha512Str},
            Downloads: downloads,
            FileSize:  info.Size(),
        })
    }

    manifest.Files = files

    data, err := json.MarshalIndent(manifest, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile(filepath.Join(outputDir, "modrinth.index.json"), data, 0644)
}