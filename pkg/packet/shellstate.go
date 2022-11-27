package packet

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/scripthaus-dev/mshell/pkg/binpack"
	"github.com/scripthaus-dev/mshell/pkg/statediff"
)

const ShellStatePackVersion = 0
const ShellStateDiffPackVersion = 0

type ShellState struct {
	Version   string `json:"version"` // [type] [semver]
	Cwd       string `json:"cwd,omitempty"`
	ShellVars []byte `json:"shellvars,omitempty"`
	Aliases   string `json:"aliases,omitempty"`
	Funcs     string `json:"funcs,omitempty"`
	Error     string `json:"error,omitempty"`
}

type ShellStateDiff struct {
	Version     string `json:"version"` // [type] [semver]
	BaseHash    string `json:"basehash"`
	Cwd         string `json:"cwd,omitempty"`
	VarsDiff    []byte `json:"shellvarsdiff,omitempty"` // vardiff
	AliasesDiff []byte `json:"aliasesdiff,omitempty"`   // linediff
	FuncsDiff   []byte `json:"funcsdiff,omitempty"`     // linediff
	Error       string `json:"error,omitempty"`
}

func (state ShellState) IsEmpty() bool {
	return state.Version == "" && state.Cwd == "" && len(state.ShellVars) == 0 && state.Aliases == "" && state.Funcs == "" && state.Error == ""
}

// returns (SHA1, encoded-state)
func (state ShellState) EncodeAndHash() (string, []byte) {
	var buf bytes.Buffer
	binpack.PackInt(&buf, ShellStatePackVersion)
	binpack.PackValue(&buf, []byte(state.Version))
	binpack.PackValue(&buf, []byte(state.Cwd))
	binpack.PackValue(&buf, state.ShellVars)
	binpack.PackValue(&buf, []byte(state.Aliases))
	binpack.PackValue(&buf, []byte(state.Funcs))
	binpack.PackValue(&buf, []byte(state.Error))
	hvalRaw := sha1.Sum(buf.Bytes())
	hval := base64.StdEncoding.EncodeToString(hvalRaw[:])
	return hval, buf.Bytes()
}

func (state ShellState) MarshalJSON() ([]byte, error) {
	_, encodedState := state.EncodeAndHash()
	return json.Marshal(encodedState)
}

func (state *ShellState) UnmarshalJSON(jsonBytes []byte) error {
	var barr []byte
	err := json.Unmarshal(jsonBytes, &barr)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(barr)
	u := binpack.MakeUnpacker(buf)
	version := u.UnpackInt("ShellState pack version")
	if version != ShellStatePackVersion {
		return fmt.Errorf("invalid ShellState pack version: %d", version)
	}
	state.Version = string(u.UnpackValue("ShellState.Version"))
	state.Cwd = string(u.UnpackValue("ShellState.Cwd"))
	state.ShellVars = u.UnpackValue("ShellState.ShellVars")
	state.Aliases = string(u.UnpackValue("ShellState.Aliases"))
	state.Funcs = string(u.UnpackValue("ShellState.Funcs"))
	state.Error = string(u.UnpackValue("ShellState.Error"))
	return u.Error()
}

func (sdiff ShellStateDiff) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	binpack.PackInt(&buf, ShellStateDiffPackVersion)
	binpack.PackValue(&buf, []byte(sdiff.Version))
	binpack.PackValue(&buf, []byte(sdiff.BaseHash))
	binpack.PackValue(&buf, []byte(sdiff.Cwd))
	binpack.PackValue(&buf, sdiff.VarsDiff)
	binpack.PackValue(&buf, sdiff.AliasesDiff)
	binpack.PackValue(&buf, sdiff.FuncsDiff)
	binpack.PackValue(&buf, []byte(sdiff.Error))
	return buf.Bytes(), nil
}

func (sdiff *ShellStateDiff) UnmarshalJSON(jsonBytes []byte) error {
	var barr []byte
	err := json.Unmarshal(jsonBytes, &barr)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(barr)
	u := binpack.MakeUnpacker(buf)
	version := u.UnpackInt("ShellState pack version")
	if version != ShellStateDiffPackVersion {
		return fmt.Errorf("invalid ShellStateDiff pack version: %d", version)
	}
	sdiff.Version = string(u.UnpackValue("ShellStateDiff.Version"))
	sdiff.BaseHash = string(u.UnpackValue("ShellStateDiff.BaseHash"))
	sdiff.Cwd = string(u.UnpackValue("ShellStateDiff.Cwd"))
	sdiff.VarsDiff = u.UnpackValue("ShellStateDiff.VarsDiff")
	sdiff.AliasesDiff = u.UnpackValue("ShellStateDiff.AliasesDiff")
	sdiff.FuncsDiff = u.UnpackValue("ShellStateDiff.FuncsDiff")
	sdiff.Error = string(u.UnpackValue("ShellStateDiff.Error"))
	return u.Error()
}

func (sdiff ShellStateDiff) Dump() {
	fmt.Printf("ShellStateDiff:\n")
	fmt.Printf("  version: %s\n", sdiff.Version)
	fmt.Printf("  base: %s\n", sdiff.BaseHash)
	var mdiff statediff.MapDiffType
	err := mdiff.Decode(sdiff.VarsDiff)
	if err != nil {
		fmt.Printf("  vars: error[%s]\n", err.Error())
	} else {
		mdiff.Dump()
	}
	fmt.Printf("  aliases: %d, funcs: %d\n", len(sdiff.AliasesDiff), len(sdiff.FuncsDiff))
	if sdiff.Error != "" {
		fmt.Printf("  error: %s\n", sdiff.Error)
	}
}
