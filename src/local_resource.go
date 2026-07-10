// local_resource.go
package pack

import (
    "crypto/sha1"
    "encoding/hex"
    "os"
)

// MurmurHash2 32-bit, seed=1
func murmurHash2(data []byte, seed uint32) uint32 {
    const m uint32 = 0x5bd1e995
    const r = 24
    h := seed ^ uint32(len(data))
    for len(data) >= 4 {
        k := uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
        k *= m
        k ^= k >> r
        k *= m
        h *= m
        h ^= k
        data = data[4:]
    }
    switch len(data) {
    case 3:
        h ^= uint32(data[2]) << 16
        fallthrough
    case 2:
        h ^= uint32(data[1]) << 8
        fallthrough
    case 1:
        h ^= uint32(data[0])
        h *= m
    }
    h ^= h >> 13
    h *= m
    h ^= h >> 15
    return h
}

type LocalResourceFile struct {
    Path           string
    FullPath       string
    ModrinthHash   string // SHA1
    CurseForgeHash uint32 // MurmurHash2 (after filtering)
}

func NewLocalResourceFile(fullPath, relPath string) (*LocalResourceFile, error) {
    data, err := os.ReadFile(fullPath)
    if err != nil {
        return nil, err
    }

    // Modrinth: SHA1 of entire file
    sha1Sum := sha1.Sum(data)
    sha1Str := hex.EncodeToString(sha1Sum[:])

    // CurseForge: MurmurHash2 after removing whitespace chars (0x09, 0x0A, 0x0D, 0x20)
    filtered := make([]byte, 0, len(data))
    for _, b := range data {
        if b != 0x09 && b != 0x0A && b != 0x0D && b != 0x20 {
            filtered = append(filtered, b)
        }
    }
    hashVal := murmurHash2(filtered, 1)

    return &LocalResourceFile{
        Path:           relPath,
        FullPath:       fullPath,
        ModrinthHash:   sha1Str,
        CurseForgeHash: hashVal,
    }, nil
}