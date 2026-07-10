// rules.go
package pack

import (
    "path/filepath"
    "strings"
)

type RuleSet struct {
    Includes []string
    Excludes []string
}

func ParseRules(rules string) RuleSet {
    parts := strings.Split(rules, "|")
    var includes, excludes []string
    for _, p := range parts {
        p = strings.TrimSpace(p)
        if p == "" {
            continue
        }
        if strings.HasPrefix(p, "!") {
            excludes = append(excludes, strings.TrimPrefix(p, "!"))
        } else {
            includes = append(includes, p)
        }
    }
    return RuleSet{Includes: includes, Excludes: excludes}
}

func (rs RuleSet) Matches(relPath string) bool {
    for _, excl := range rs.Excludes {
        if matchPattern(excl, relPath) {
            return false
        }
    }
    if len(rs.Includes) == 0 {
        return true
    }
    for _, incl := range rs.Includes {
        if matchPattern(incl, relPath) {
            return true
        }
    }
    return false
}

func matchPattern(pattern, path string) bool {
    if strings.HasSuffix(pattern, "/") {
        return strings.HasPrefix(path, pattern)
    }
    matched, err := filepath.Match(pattern, path)
    if err != nil {
        return strings.HasPrefix(path, pattern)
    }
    return matched
}