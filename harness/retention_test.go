package harness

import (
	"flag"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnvBool_Truthiness(t *testing.T) {
	tests := []struct {
		name  string
		value string
		set   bool
		want  bool
	}{
		{"unset", "", false, false},
		{"empty", "", true, false},
		{"zero", "0", true, false},
		{"false lower", "false", true, false},
		{"false upper", "FALSE", true, false},
		{"false mixed", "False", true, false},
		{"one", "1", true, true},
		{"true", "true", true, true},
		{"yes", "yes", true, true},
		{"arbitrary", "anything", true, true},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name := fmt.Sprintf("HARNESS_TEST_ENVBOOL_%d", i)
			if tt.set {
				t.Setenv(name, tt.value)
			}
			require.Equal(t, tt.want, envBool(name))
		})
	}
}

func TestRegisterFlags_SafeToCallTwice(t *testing.T) {
	require.NotPanics(t, func() { RegisterFlags() })
	require.NotPanics(t, func() { RegisterFlags() })
	require.NotNil(t, flag.CommandLine.Lookup("keep-artifacts"))
	require.NotNil(t, flag.CommandLine.Lookup("keep-on-failure"))
}
