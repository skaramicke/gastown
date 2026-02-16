// Package session provides polecat session lifecycle management.
package session

import "sync"

// Registry maps session name prefixes (e.g., "gt", "bd", "hop") to rig names
// (e.g., "gastown", "beads", "hop") and vice versa. This is populated at
// startup from rig configuration and used by ParseSessionName to resolve
// prefixes back to rig names.
var (
	registryMu  sync.RWMutex
	prefixToRig = map[string]string{}
	rigToPrefix = map[string]string{}
)

// RegisterPrefix registers a mapping between a beads prefix and a rig name.
// Must be called at startup before session names are parsed.
func RegisterPrefix(prefix, rigName string) {
	registryMu.Lock()
	defer registryMu.Unlock()
	prefixToRig[prefix] = rigName
	rigToPrefix[rigName] = prefix
}

// PrefixForRig returns the session prefix for a rig name, or "" if unknown.
func PrefixForRig(rigName string) string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return rigToPrefix[rigName]
}

// RigForPrefix returns the rig name for a session prefix, or "" if unknown.
func RigForPrefix(prefix string) string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return prefixToRig[prefix]
}

// AllRegisteredPrefixes returns all known session prefixes.
func AllRegisteredPrefixes() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	prefixes := make([]string, 0, len(prefixToRig))
	for p := range prefixToRig {
		prefixes = append(prefixes, p)
	}
	return prefixes
}

// IsKnownPrefix returns true if the prefix is registered.
func IsKnownPrefix(prefix string) bool {
	registryMu.RLock()
	defer registryMu.RUnlock()
	_, ok := prefixToRig[prefix]
	return ok
}

// ClearRegistry removes all registered prefixes. For testing only.
func ClearRegistry() {
	registryMu.Lock()
	defer registryMu.Unlock()
	prefixToRig = map[string]string{}
	rigToPrefix = map[string]string{}
}
