// exporter.go
package pack

import (
    "fmt"
    "io"
    "io/fs"
    "os"
    "path/filepath"
    "strings"
)

func Export(cfg *Config, outputDir string, initGit bool) error {
    if err := os.MkdirAll(outputDir, 0755); err != nil {
        return err
    }

    files, err := CollectFiles(cfg.GamePath, cfg.Export.Options)
    if err != nil {
        return err
    }
    if len(files) == 0 {
        return fmt.Errorf("no files found to export")
    }

    overridesDir := filepath.Join(outputDir, "overrides")
    if err := os.MkdirAll(overridesDir, 0755); err != nil {
        return err
    }

    var resourceFiles []*LocalResourceFile
    var normalFiles []string

    for _, rel := range files {
        if isResourceFile(rel) {
            src := filepath.Join(cfg.GamePath, rel)
            res, err := NewLocalResourceFile(src, rel)
            if err != nil {
                fmt.Printf("Warning: failed to compute hash for %s (%v), will copy directly\n", rel, err)
                normalFiles = append(normalFiles, rel)
            } else {
                resourceFiles = append(resourceFiles, res)
            }
        } else {
            normalFiles = append(normalFiles, rel)
        }
    }

    fmt.Printf("Copying normal files, total %d ...\n", len(normalFiles))
    for i, rel := range normalFiles {
        src := filepath.Join(cfg.GamePath, rel)
        dst := filepath.Join(overridesDir, rel)
        if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
            return err
        }
        if err := copyFile(src, dst); err != nil {
            return fmt.Errorf("failed to copy %s: %w", rel, err)
        }
        if (i+1)%50 == 0 {
            fmt.Printf("Copied %d/%d normal files\n", i+1, len(normalFiles))
        }
    }

    var manifestResources []*LocalResourceFile
    var unmatchedResources []*LocalResourceFile
    downloadMap := make(map[string]string)

    if cfg.Advanced.SkipNetwork {
        fmt.Printf("Skipping network checks, copying %d resource files directly...\n", len(resourceFiles))
        for i, res := range resourceFiles {
            dst := filepath.Join(overridesDir, res.Path)
            if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
                return err
            }
            if err := copyFile(res.FullPath, dst); err != nil {
                return fmt.Errorf("failed to copy %s: %w", res.Path, err)
            }
            if (i+1)%50 == 0 {
                fmt.Printf("Copied %d/%d resource files\n", i+1, len(resourceFiles))
            }
        }
        fmt.Println("Note: pack contains full files, do not publicly distribute (may violate licenses)")
    } else {
        fmt.Printf("Querying network for %d resource files...\n", len(resourceFiles))
        if len(resourceFiles) > 0 {
            sha1List := make([]string, len(resourceFiles))
            for i, r := range resourceFiles {
                sha1List[i] = r.ModrinthHash
            }
            modrinthMap, err := QueryModrinthDownloads(sha1List)
            if err != nil {
                fmt.Printf("Modrinth query failed: %v\n", err)
            } else {
                for h, url := range modrinthMap {
                    downloadMap[h] = url
                }
                fmt.Printf("Modrinth matched %d resources\n", len(modrinthMap))
            }

            if !cfg.Advanced.ModrinthOnly {
                var unmatched []*LocalResourceFile
                for _, r := range resourceFiles {
                    if _, ok := downloadMap[r.ModrinthHash]; !ok {
                        unmatched = append(unmatched, r)
                    }
                }
                if len(unmatched) > 0 {
                    fingerprints := make([]uint32, len(unmatched))
                    for i, r := range unmatched {
                        fingerprints[i] = r.CurseForgeHash
                    }
                    apiKey := os.Getenv("CURSEFORGE_API_KEY")
                    cfMap, err := QueryCurseForgeDownloads(fingerprints, apiKey)
                    if err != nil {
                        fmt.Printf("CurseForge query failed: %v\n", err)
                    } else {
                        for _, r := range unmatched {
                            if url, ok := cfMap[r.CurseForgeHash]; ok {
                                downloadMap[fmt.Sprintf("cf_%d", r.CurseForgeHash)] = url
                            }
                        }
                        fmt.Printf("CurseForge matched %d resources\n", len(cfMap))
                    }
                }
            }

            for _, res := range resourceFiles {
                if _, ok := downloadMap[res.ModrinthHash]; ok {
                    manifestResources = append(manifestResources, res)
                } else if _, ok := downloadMap[fmt.Sprintf("cf_%d", res.CurseForgeHash)]; ok {
                    manifestResources = append(manifestResources, res)
                } else {
                    unmatchedResources = append(unmatchedResources, res)
                }
            }

            if len(unmatchedResources) > 0 {
                fmt.Printf("Copying %d unmatched resource files to overrides\n", len(unmatchedResources))
                for i, res := range unmatchedResources {
                    dst := filepath.Join(overridesDir, res.Path)
                    if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
                        return err
                    }
                    if err := copyFile(res.FullPath, dst); err != nil {
                        return fmt.Errorf("failed to copy unmatched resource %s: %w", res.Path, err)
                    }
                    if (i+1)%50 == 0 {
                        fmt.Printf("Copied %d/%d unmatched resources\n", i+1, len(unmatchedResources))
                    }
                }
            }
        } else {
            fmt.Println("No resource files to query")
        }
    }

    if len(cfg.ExtraFiles) > 0 {
        fmt.Printf("Processing extra files (%d items)...\n", len(cfg.ExtraFiles))
        for _, line := range cfg.ExtraFiles {
            line = strings.TrimSpace(line)
            if line == "" || strings.HasPrefix(line, "#") {
                continue
            }
            isDir := strings.HasSuffix(line, "\\") || strings.HasSuffix(line, "/")
            src := filepath.Clean(line)
            baseName := filepath.Base(src)
            if isDir {
                dst := filepath.Join(outputDir, baseName)
                fmt.Printf("  Copying directory: %s -> %s\n", src, dst)
                if err := copyDir(src, dst); err != nil {
                    fmt.Printf("  Warning: failed to copy directory %s: %v\n", src, err)
                }
            } else {
                dst := filepath.Join(outputDir, baseName)
                fmt.Printf("  Copying file: %s -> %s\n", src, dst)
                if err := copyFile(src, dst); err != nil {
                    fmt.Printf("  Warning: failed to copy file %s: %v\n", src, err)
                }
            }
        }
    }

    if len(manifestResources) > 0 && !cfg.Advanced.SkipNetwork {
        versionInfo, err := DetectVersion(cfg.GamePath)
        if err != nil {
            fmt.Printf("Warning: version detection failed: %v, using default\n", err)
            versionInfo = &VersionInfo{Minecraft: "1.20.1"}
        } else {
            fmt.Printf("Detected Minecraft %s", versionInfo.Minecraft)
            if versionInfo.Forge != "" {
                fmt.Printf(", Forge %s", versionInfo.Forge)
            }
            if versionInfo.Fabric != "" {
                fmt.Printf(", Fabric %s", versionInfo.Fabric)
            }
            if versionInfo.NeoForge != "" {
                fmt.Printf(", NeoForge %s", versionInfo.NeoForge)
            }
            fmt.Println()
        }

        if err := GenerateManifest(outputDir, cfg.Export.Name, cfg.Export.Version, versionInfo, manifestResources, downloadMap); err != nil {
            return fmt.Errorf("failed to generate manifest: %w", err)
        }
        fmt.Printf("Manifest generated, contains %d downloadable resources\n", len(manifestResources))
    } else if cfg.Advanced.SkipNetwork {
        fmt.Println("Network checks skipped, all files packaged without manifest")
    } else {
        fmt.Println("No resources matched to download links, manifest not generated")
    }

    if initGit {
        if err := InitGitRepo(outputDir, cfg.Git.Remote, cfg.Git.Branch); err != nil {
            return fmt.Errorf("Git initialisation failed: %w", err)
        }
    }

    fmt.Printf("Export completed, files updated in %s\n", outputDir)
    return nil
}

func copyFile(src, dst string) error {
    srcInfo, err := os.Stat(src)
    if err != nil {
        return err
    }
    if srcInfo.IsDir() {
        return nil
    }
    dst = filepath.Clean(dst)
    dir := filepath.Dir(dst)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return err
    }
    srcFile, err := os.Open(src)
    if err != nil {
        return err
    }
    defer srcFile.Close()
    dstFile, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer dstFile.Close()
    _, err = io.Copy(dstFile, srcFile)
    if err != nil {
        return err
    }
    _ = os.Chmod(dst, srcInfo.Mode())
    return nil
}

func copyDir(src, dst string) error {
    return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }
        rel, err := filepath.Rel(src, path)
        if err != nil {
            return err
        }
        dstPath := filepath.Join(dst, rel)
        if d.IsDir() {
            return os.MkdirAll(dstPath, 0755)
        }
        return copyFile(path, dstPath)
    })
}

func isResourceFile(rel string) bool {
    ext := strings.ToLower(filepath.Ext(rel))
    return ext == ".jar" || ext == ".zip" || ext == ".rar" || ext == ".disabled" || ext == ".old"
}