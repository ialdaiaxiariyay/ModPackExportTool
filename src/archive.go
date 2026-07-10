// archive.go
package pack

import (
    "archive/zip"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strings"
)

func CreateArchive(sourceDir, archivePath string) error {

    if err := os.MkdirAll(filepath.Dir(archivePath), 0755); err != nil {
        return err
    }

    zipFile, err := os.Create(archivePath)
    if err != nil {
        return err
    }
    defer zipFile.Close()

    zipWriter := zip.NewWriter(zipFile)
    defer zipWriter.Close()

    baseDir := filepath.Clean(sourceDir)
    err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if path == sourceDir {
            return nil
        }
        relPath, err := filepath.Rel(baseDir, path)
        if err != nil {
            return err
        }

        if strings.HasPrefix(relPath, ".git") {
            if info.IsDir() {
                return filepath.SkipDir
            }
            return nil
        }
        if info.IsDir() {
            return nil
        }
        zipRelPath := filepath.ToSlash(relPath)

        srcFile, err := os.Open(path)
        if err != nil {
            return err
        }
        defer srcFile.Close()

        header, err := zip.FileInfoHeader(info)
        if err != nil {
            return err
        }
        header.Name = zipRelPath
        header.Method = zip.Deflate

        writer, err := zipWriter.CreateHeader(header)
        if err != nil {
            return err
        }
        _, err = io.Copy(writer, srcFile)
        return err
    })

    if err != nil {
        return err
    }
    fmt.Printf("Archive created: %s\n", archivePath)
    return nil
}