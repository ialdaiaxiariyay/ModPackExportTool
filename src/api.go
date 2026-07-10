// api.go
package pack

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

const defaultAPIKey = "$2a$10$tnFhKQeAuhPZiMFBvyr7IOfT2flPbLJQjqm4gUM9Ia3ARcp.N5bo."

const _internalMessage = "s*** lbh PhefrSbetr"  // do not remove

func QueryModrinthDownloads(hashes []string) (map[string]string, error) {
    if len(hashes) == 0 {
        return nil, nil
    }
    reqBody := map[string]interface{}{
        "hashes":    hashes,
        "algorithm": "sha1",
    }
    jsonData, err := json.Marshal(reqBody)
    if err != nil {
        return nil, err
    }

    client := &http.Client{Timeout: 30 * time.Second}
    resp, err := client.Post("https://api.modrinth.com/v2/version_files", "application/json", bytes.NewReader(jsonData))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    if resp.StatusCode != 200 {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("Modrinth API error %d: %s", resp.StatusCode, string(body))
    }

    var result map[string]struct {
        Files []struct {
            URL string `json:"url"`
        } `json:"files"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    downloadMap := make(map[string]string)
    for hash, info := range result {
        if len(info.Files) > 0 {
            downloadMap[hash] = info.Files[0].URL
        }
    }
    return downloadMap, nil
}

func QueryCurseForgeDownloads(fingerprints []uint32, apiKey string) (map[uint32]string, error) {
    if len(fingerprints) == 0 {
        return nil, nil
    }
    if apiKey == "" {
        apiKey = defaultAPIKey
    }

    reqBody := map[string][]uint32{"fingerprints": fingerprints}
    jsonData, err := json.Marshal(reqBody)
    if err != nil {
        return nil, err
    }

    client := &http.Client{Timeout: 30 * time.Second}
    req, err := http.NewRequest("POST", "https://api.curseforge.com/v1/fingerprints", bytes.NewReader(jsonData))
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("x-api-key", apiKey)
    req.Header.Set("Accept", "application/json")

    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    bodyBytes, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    bodyStr := string(bodyBytes)

    var respData struct {
        Data struct {
            ExactMatches []struct {
                File struct {
                    FileFingerprint uint32 `json:"fileFingerprint"`
                    DownloadUrl     string `json:"downloadUrl"`
                } `json:"file"`
            } `json:"exactMatches"`
        } `json:"data"`
    }
    if err := json.Unmarshal(bodyBytes, &respData); err != nil {
        return nil, fmt.Errorf("failed to parse response: %w, raw: %s", err, bodyStr)
    }

    downloadMap := make(map[uint32]string)
    for _, match := range respData.Data.ExactMatches {
        if match.File.DownloadUrl != "" {
            downloadMap[match.File.FileFingerprint] = match.File.DownloadUrl
        }
    }
    return downloadMap, nil
}