// Package session provides polecat session lifecycle management.
package session

import (
	"fmt"
	"strings"
)

// Role represents the type of Gas Town agent.
type Role string

const (
	RoleMayor    Role = "mayor"
	RoleDeacon   Role = "deacon"
	RoleWitness  Role = "witness"
	RoleRefinery Role = "refinery"
	RoleCrew     Role = "crew"
	RolePolecat  Role = "polecat"
)

// AgentIdentity represents a parsed Gas Town agent identity.
type AgentIdentity struct {
	Role   Role   // mayor, deacon, witness, refinery, crew, polecat
	Rig    string // rig name (empty for mayor/deacon)
	Name   string // crew/polecat name (empty for mayor/deacon/witness/refinery)
	Prefix string // session prefix (e.g., "gt", "bd", "hq"); empty for mayor/deacon
}

// ParseAddress parses a mail-style address into an AgentIdentity.
func ParseAddress(address string) (*AgentIdentity, error) {
	address = strings.TrimSpace(address)
	if address == "" {
		return nil, fmt.Errorf("empty address")
	}

	if address == "mayor" || address == "mayor/" {
		return &AgentIdentity{Role: RoleMayor}, nil
	}
	if address == "deacon" || address == "deacon/" {
		return &AgentIdentity{Role: RoleDeacon}, nil
	}
	if address == "overseer" {
		return nil, fmt.Errorf("overseer has no session")
	}

	address = strings.TrimSuffix(address, "/")
	parts := strings.Split(address, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid address %q", address)
	}

	rig := parts[0]
	prefix := PrefixForRig(rig)
	switch len(parts) {
	case 2:
		name := parts[1]
		switch name {
		case "witness":
			return &AgentIdentity{Role: RoleWitness, Rig: rig, Prefix: prefix}, nil
		case "refinery":
			return &AgentIdentity{Role: RoleRefinery, Rig: rig, Prefix: prefix}, nil
		case "crew", "polecats":
			return nil, fmt.Errorf("invalid address %q", address)
		default:
			return &AgentIdentity{Role: RolePolecat, Rig: rig, Name: name, Prefix: prefix}, nil
		}
	case 3:
		role := parts[1]
		name := parts[2]
		switch role {
		case "crew":
			return &AgentIdentity{Role: RoleCrew, Rig: rig, Name: name, Prefix: prefix}, nil
		case "polecats":
			return &AgentIdentity{Role: RolePolecat, Rig: rig, Name: name, Prefix: prefix}, nil
		default:
			return nil, fmt.Errorf("invalid address %q", address)
		}
	default:
		return nil, fmt.Errorf("invalid address %q", address)
	}
}

// ParseSessionName parses a tmux session name into an AgentIdentity.
//
// Session name formats (new prefix-based naming):
//   - hq-mayor → Role: mayor (town-level, one per machine)
//   - hq-deacon → Role: deacon (town-level, one per machine)
//   - hq-boot → Role: deacon, Name: boot (boot watchdog)
//   - <prefix>-witness → Role: witness (e.g., gt-witness, bd-witness)
//   - <prefix>-refinery → Role: refinery (e.g., gt-refinery, bd-refinery)
//   - <prefix>-crew-<name> → Role: crew (e.g., gt-crew-jack, bd-crew-emma)
//   - <prefix>-<name> → Role: polecat (e.g., gt-furiosa, bd-worker1)
//
// The prefix (before the first hyphen) identifies the rig via the prefix
// registry. If the prefix is not registered, the Rig field will be empty.
func ParseSessionName(session string) (*AgentIdentity, error) {
	// Check for town-level roles (hq- prefix)
	if strings.HasPrefix(session, HQPrefix) {
		suffix := strings.TrimPrefix(session, HQPrefix)
		if suffix == "mayor" {
			return &AgentIdentity{Role: RoleMayor}, nil
		}
		if suffix == "deacon" {
			return &AgentIdentity{Role: RoleDeacon}, nil
		}
		if suffix == "boot" {
			return &AgentIdentity{Role: RoleDeacon, Name: "boot"}, nil
		}
		if suffix == "overseer" {
			return &AgentIdentity{Role: RoleMayor, Name: "overseer"}, nil
		}
		return nil, fmt.Errorf("invalid session name %q: unknown hq- role", session)
	}

	// Rig-level sessions: <prefix>-<rest>
	// The prefix is everything before the first hyphen.
	idx := strings.Index(session, "-")
	if idx <= 0 || idx == len(session)-1 {
		return nil, fmt.Errorf("invalid session name %q: no prefix-role separator", session)
	}

	prefix := session[:idx]
	rest := session[idx+1:] // everything after "prefix-"

	// Look up rig name from prefix registry
	rigName := RigForPrefix(prefix)

	// Parse the role/name from the rest
	// "witness" → witness
	// "refinery" → refinery
	// "crew-<name>" → crew
	// "<name>" → polecat (fallback)

	if rest == "witness" {
		return &AgentIdentity{Role: RoleWitness, Rig: rigName, Prefix: prefix}, nil
	}
	if rest == "refinery" {
		return &AgentIdentity{Role: RoleRefinery, Rig: rigName, Prefix: prefix}, nil
	}
	if rest == "boot" {
		return &AgentIdentity{Role: RoleDeacon, Name: "boot", Prefix: prefix}, nil
	}

	// Check for crew marker
	if strings.HasPrefix(rest, "crew-") {
		name := rest[5:] // len("crew-") == 5
		if name == "" {
			return nil, fmt.Errorf("invalid session name %q: empty crew name", session)
		}
		return &AgentIdentity{Role: RoleCrew, Rig: rigName, Name: name, Prefix: prefix}, nil
	}

	// Default: polecat
	if rest == "" {
		return nil, fmt.Errorf("invalid session name %q: empty name after prefix", session)
	}
	return &AgentIdentity{Role: RolePolecat, Rig: rigName, Name: rest, Prefix: prefix}, nil
}

// SessionName returns the tmux session name for this identity.
func (a *AgentIdentity) SessionName() string {
	switch a.Role {
	case RoleMayor:
		return MayorSessionName()
	case RoleDeacon:
		if a.Name == "boot" {
			return BootSessionName()
		}
		return DeaconSessionName()
	case RoleWitness:
		return WitnessSessionName(a.rigPrefix())
	case RoleRefinery:
		return RefinerySessionName(a.rigPrefix())
	case RoleCrew:
		return CrewSessionName(a.rigPrefix(), a.Name)
	case RolePolecat:
		return PolecatSessionName(a.rigPrefix(), a.Name)
	default:
		return ""
	}
}

// rigPrefix returns the session prefix for this identity.
// Uses the Prefix field if set, otherwise looks up from the registry.
func (a *AgentIdentity) rigPrefix() string {
	if a.Prefix != "" {
		return a.Prefix
	}
	if a.Rig != "" {
		if p := PrefixForRig(a.Rig); p != "" {
			return p
		}
	}
	return ""
}

// Address returns the mail-style address for this identity.
// Examples:
//   - mayor → "mayor"
//   - deacon → "deacon"
//   - witness → "gastown/witness"
//   - refinery → "gastown/refinery"
//   - crew → "gastown/crew/max"
//   - polecat → "gastown/polecats/Toast"
func (a *AgentIdentity) Address() string {
	switch a.Role {
	case RoleMayor:
		return "mayor"
	case RoleDeacon:
		return "deacon"
	case RoleWitness:
		return fmt.Sprintf("%s/witness", a.Rig)
	case RoleRefinery:
		return fmt.Sprintf("%s/refinery", a.Rig)
	case RoleCrew:
		return fmt.Sprintf("%s/crew/%s", a.Rig, a.Name)
	case RolePolecat:
		return fmt.Sprintf("%s/polecats/%s", a.Rig, a.Name)
	default:
		return ""
	}
}

// GTRole returns the GT_ROLE environment variable format.
// This is the same as Address() for most roles.
func (a *AgentIdentity) GTRole() string {
	return a.Address()
}
