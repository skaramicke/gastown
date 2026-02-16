package session

import (
	"testing"
)

func TestParseSessionName(t *testing.T) {
	// Register prefixes for test
	ClearRegistry()
	RegisterPrefix("gt", "gastown")
	RegisterPrefix("bd", "beads")
	RegisterPrefix("hop", "hop")
	defer ClearRegistry()

	tests := []struct {
		name       string
		session    string
		wantRole   Role
		wantRig    string
		wantName   string
		wantPrefix string
		wantErr    bool
	}{
		// Town-level roles (hq-mayor, hq-deacon)
		{
			name:     "mayor",
			session:  "hq-mayor",
			wantRole: RoleMayor,
		},
		{
			name:     "deacon",
			session:  "hq-deacon",
			wantRole: RoleDeacon,
		},

		// Witness
		{
			name:       "witness gastown",
			session:    "gt-witness",
			wantRole:   RoleWitness,
			wantRig:    "gastown",
			wantPrefix: "gt",
		},
		{
			name:       "witness beads",
			session:    "bd-witness",
			wantRole:   RoleWitness,
			wantRig:    "beads",
			wantPrefix: "bd",
		},

		// Refinery
		{
			name:       "refinery gastown",
			session:    "gt-refinery",
			wantRole:   RoleRefinery,
			wantRig:    "gastown",
			wantPrefix: "gt",
		},
		{
			name:       "refinery beads",
			session:    "bd-refinery",
			wantRole:   RoleRefinery,
			wantRig:    "beads",
			wantPrefix: "bd",
		},

		// Crew
		{
			name:       "crew simple",
			session:    "gt-crew-max",
			wantRole:   RoleCrew,
			wantRig:    "gastown",
			wantName:   "max",
			wantPrefix: "gt",
		},
		{
			name:       "crew beads",
			session:    "bd-crew-alice",
			wantRole:   RoleCrew,
			wantRig:    "beads",
			wantName:   "alice",
			wantPrefix: "bd",
		},
		{
			name:       "crew hyphenated name",
			session:    "gt-crew-my-worker",
			wantRole:   RoleCrew,
			wantRig:    "gastown",
			wantName:   "my-worker",
			wantPrefix: "gt",
		},

		// Polecat
		{
			name:       "polecat simple",
			session:    "gt-morsov",
			wantRole:   RolePolecat,
			wantRig:    "gastown",
			wantName:   "morsov",
			wantPrefix: "gt",
		},
		{
			name:       "polecat beads",
			session:    "bd-worker1",
			wantRole:   RolePolecat,
			wantRig:    "beads",
			wantName:   "worker1",
			wantPrefix: "bd",
		},
		{
			name:       "polecat hop",
			session:    "hop-ostrom",
			wantRole:   RolePolecat,
			wantRig:    "hop",
			wantName:   "ostrom",
			wantPrefix: "hop",
		},

		// Boot watchdog
		{
			name:     "boot hq",
			session:  "hq-boot",
			wantRole: RoleDeacon,
			wantName: "boot",
		},

		// Unknown prefix (not registered)
		{
			name:       "unknown prefix",
			session:    "xyz-witness",
			wantRole:   RoleWitness,
			wantRig:    "", // not in registry
			wantPrefix: "xyz",
		},

		// Error cases
		{
			name:    "no separator",
			session: "gastown",
			wantErr: true,
		},
		{
			name:    "empty after prefix",
			session: "gt-",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSessionName(tt.session)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSessionName(%q) error = %v, wantErr %v", tt.session, err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if got.Role != tt.wantRole {
				t.Errorf("ParseSessionName(%q).Role = %v, want %v", tt.session, got.Role, tt.wantRole)
			}
			if got.Rig != tt.wantRig {
				t.Errorf("ParseSessionName(%q).Rig = %v, want %v", tt.session, got.Rig, tt.wantRig)
			}
			if got.Name != tt.wantName {
				t.Errorf("ParseSessionName(%q).Name = %v, want %v", tt.session, got.Name, tt.wantName)
			}
			if got.Prefix != tt.wantPrefix {
				t.Errorf("ParseSessionName(%q).Prefix = %v, want %v", tt.session, got.Prefix, tt.wantPrefix)
			}
		})
	}
}

func TestAgentIdentity_SessionName(t *testing.T) {
	ClearRegistry()
	RegisterPrefix("gt", "gastown")
	RegisterPrefix("bd", "beads")
	defer ClearRegistry()

	tests := []struct {
		name     string
		identity AgentIdentity
		want     string
	}{
		{
			name:     "mayor",
			identity: AgentIdentity{Role: RoleMayor},
			want:     "hq-mayor",
		},
		{
			name:     "deacon",
			identity: AgentIdentity{Role: RoleDeacon},
			want:     "hq-deacon",
		},
		{
			name:     "witness gastown",
			identity: AgentIdentity{Role: RoleWitness, Rig: "gastown", Prefix: "gt"},
			want:     "gt-witness",
		},
		{
			name:     "refinery beads",
			identity: AgentIdentity{Role: RoleRefinery, Rig: "beads", Prefix: "bd"},
			want:     "bd-refinery",
		},
		{
			name:     "crew",
			identity: AgentIdentity{Role: RoleCrew, Rig: "gastown", Name: "max", Prefix: "gt"},
			want:     "gt-crew-max",
		},
		{
			name:     "polecat",
			identity: AgentIdentity{Role: RolePolecat, Rig: "gastown", Name: "morsov", Prefix: "gt"},
			want:     "gt-morsov",
		},
		{
			name:     "witness prefix lookup from rig",
			identity: AgentIdentity{Role: RoleWitness, Rig: "gastown"},
			want:     "gt-witness",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.identity.SessionName(); got != tt.want {
				t.Errorf("AgentIdentity.SessionName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAgentIdentity_Address(t *testing.T) {
	tests := []struct {
		name     string
		identity AgentIdentity
		want     string
	}{
		{
			name:     "mayor",
			identity: AgentIdentity{Role: RoleMayor},
			want:     "mayor",
		},
		{
			name:     "deacon",
			identity: AgentIdentity{Role: RoleDeacon},
			want:     "deacon",
		},
		{
			name:     "witness",
			identity: AgentIdentity{Role: RoleWitness, Rig: "gastown"},
			want:     "gastown/witness",
		},
		{
			name:     "refinery",
			identity: AgentIdentity{Role: RoleRefinery, Rig: "my-project"},
			want:     "my-project/refinery",
		},
		{
			name:     "crew",
			identity: AgentIdentity{Role: RoleCrew, Rig: "gastown", Name: "max"},
			want:     "gastown/crew/max",
		},
		{
			name:     "polecat",
			identity: AgentIdentity{Role: RolePolecat, Rig: "gastown", Name: "Toast"},
			want:     "gastown/polecats/Toast",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.identity.Address(); got != tt.want {
				t.Errorf("AgentIdentity.Address() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseSessionName_RoundTrip(t *testing.T) {
	ClearRegistry()
	RegisterPrefix("gt", "gastown")
	RegisterPrefix("bd", "beads")
	defer ClearRegistry()

	// Test that parsing then reconstructing gives the same result
	sessions := []string{
		"hq-mayor",
		"hq-deacon",
		"gt-witness",
		"bd-refinery",
		"gt-crew-max",
		"gt-morsov",
		"bd-worker1",
	}

	for _, sess := range sessions {
		t.Run(sess, func(t *testing.T) {
			identity, err := ParseSessionName(sess)
			if err != nil {
				t.Fatalf("ParseSessionName(%q) error = %v", sess, err)
			}
			if got := identity.SessionName(); got != sess {
				t.Errorf("Round-trip failed: ParseSessionName(%q).SessionName() = %q", sess, got)
			}
		})
	}
}

func TestParseAddress(t *testing.T) {
	ClearRegistry()
	RegisterPrefix("gt", "gastown")
	defer ClearRegistry()

	tests := []struct {
		name    string
		address string
		want    AgentIdentity
		wantErr bool
	}{
		{
			name:    "mayor",
			address: "mayor/",
			want:    AgentIdentity{Role: RoleMayor},
		},
		{
			name:    "deacon",
			address: "deacon",
			want:    AgentIdentity{Role: RoleDeacon},
		},
		{
			name:    "witness",
			address: "gastown/witness",
			want:    AgentIdentity{Role: RoleWitness, Rig: "gastown", Prefix: "gt"},
		},
		{
			name:    "refinery",
			address: "rig-a/refinery",
			want:    AgentIdentity{Role: RoleRefinery, Rig: "rig-a"},
		},
		{
			name:    "crew",
			address: "gastown/crew/max",
			want:    AgentIdentity{Role: RoleCrew, Rig: "gastown", Name: "max", Prefix: "gt"},
		},
		{
			name:    "polecat explicit",
			address: "gastown/polecats/nux",
			want:    AgentIdentity{Role: RolePolecat, Rig: "gastown", Name: "nux", Prefix: "gt"},
		},
		{
			name:    "polecat canonical",
			address: "gastown/nux",
			want:    AgentIdentity{Role: RolePolecat, Rig: "gastown", Name: "nux", Prefix: "gt"},
		},
		{
			name:    "invalid",
			address: "gastown/crew",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAddress(tt.address)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseAddress(%q) error = %v", tt.address, err)
			}
			if *got != tt.want {
				t.Fatalf("ParseAddress(%q) = %#v, want %#v", tt.address, *got, tt.want)
			}
		})
	}
}

func TestRegistryOperations(t *testing.T) {
	ClearRegistry()
	defer ClearRegistry()

	// Register some prefixes
	RegisterPrefix("gt", "gastown")
	RegisterPrefix("bd", "beads")

	// Test lookups
	if got := PrefixForRig("gastown"); got != "gt" {
		t.Errorf("PrefixForRig(gastown) = %q, want %q", got, "gt")
	}
	if got := RigForPrefix("bd"); got != "beads" {
		t.Errorf("RigForPrefix(bd) = %q, want %q", got, "beads")
	}
	if got := PrefixForRig("unknown"); got != "" {
		t.Errorf("PrefixForRig(unknown) = %q, want empty", got)
	}
	if !IsKnownPrefix("gt") {
		t.Error("IsKnownPrefix(gt) = false, want true")
	}
	if IsKnownPrefix("zz") {
		t.Error("IsKnownPrefix(zz) = true, want false")
	}

	prefixes := AllRegisteredPrefixes()
	if len(prefixes) != 2 {
		t.Errorf("AllRegisteredPrefixes() returned %d, want 2", len(prefixes))
	}
}
