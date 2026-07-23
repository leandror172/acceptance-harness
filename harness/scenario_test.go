package harness

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun_ExecutesPhasesInOrder(t *testing.T) {
	var order []string
	s := Scenario{
		Name: "test",
		Given: func(ctx *Context) {
			order = append(order, "given")
		},
		When: func(ctx *Context) {
			order = append(order, "when")
		},
		Then: []func(*Context){
			func(ctx *Context) {
				order = append(order, "then1")
			},
			func(ctx *Context) {
				order = append(order, "then2")
			},
		},
	}
	Run(t, s)
	assert.Equal(t, []string{"given", "when", "then1", "then2"}, order)
}

func TestRun_InitializesContext(t *testing.T) {
	var capturedCtx *Context
	s := Scenario{
		Name: "test",
		Given: func(ctx *Context) {
			capturedCtx = ctx
		},
	}
	Run(t, s)

	assert.NotEmpty(t, capturedCtx.WorkDir)
	assert.DirExists(t, capturedCtx.WorkDir)
	assert.NotNil(t, capturedCtx.Artifacts)
	assert.NotNil(t, capturedCtx.Env)
	assert.Equal(t, t, capturedCtx.T)
}

func TestRun_TolerantOfNilPhases(t *testing.T) {
	var order []string
	s := Scenario{
		Name:  "test",
		Given: nil,
		When:  nil,
		Then: []func(*Context){
			nil,
			func(ctx *Context) {
				order = append(order, "then")
			},
		},
	}
	Run(t, s)
	assert.Equal(t, []string{"then"}, order)
}

func TestRun_RemovesWorkDirAfterTest(t *testing.T) {
	var workDir string
	t.Run("inner", func(t *testing.T) {
		Run(t, Scenario{
			Name: "test",
			Then: []func(*Context){
				func(ctx *Context) {
					workDir = ctx.WorkDir
				},
			},
		})
	})
	assert.NoDirExists(t, workDir)
}

// --- v1.1: engine-owned scenario plumbing ---------------------------------

func TestRun_AssignsFixtureDirBeforeGiven(t *testing.T) {
	var seen string
	Run(t, Scenario{
		Name:    "test",
		Fixture: "/fixtures/example",
		Given: func(ctx *Context) {
			seen = ctx.FixtureDir
		},
	})
	assert.Equal(t, "/fixtures/example", seen)
}

func TestRun_GivenStillOwnsFixtureDirWhenScenarioOmitsIt(t *testing.T) {
	var seen string
	Run(t, Scenario{
		Name: "test",
		Given: func(ctx *Context) {
			ctx.FixtureDir = "/assigned/by/given"
		},
		Then: []func(*Context){
			func(ctx *Context) { seen = ctx.FixtureDir },
		},
	})
	assert.Equal(t, "/assigned/by/given", seen)
}

func TestRun_AssignsBinaryFromUseBinary(t *testing.T) {
	t.Cleanup(func() { UseBinary("") })
	UseBinary("/build/cli")

	var seen string
	Run(t, Scenario{
		Name:  "test",
		Given: func(ctx *Context) { seen = ctx.BinaryPath },
	})
	assert.Equal(t, "/build/cli", seen)
}

func TestRun_ScenarioBinaryOverridesDefault(t *testing.T) {
	t.Cleanup(func() { UseBinary("") })
	UseBinary("/build/cli")

	var seen string
	Run(t, Scenario{
		Name:   "test",
		Binary: "/build/other-cli",
		Given:  func(ctx *Context) { seen = ctx.BinaryPath },
	})
	assert.Equal(t, "/build/other-cli", seen)
}

func TestRun_GivenStillOwnsBinaryPath(t *testing.T) {
	t.Cleanup(func() { UseBinary("") })
	UseBinary("/build/cli")

	var seen string
	Run(t, Scenario{
		Name:  "test",
		Given: func(ctx *Context) { ctx.BinaryPath = "/assigned/by/given" },
		Then: []func(*Context){
			func(ctx *Context) { seen = ctx.BinaryPath },
		},
	})
	assert.Equal(t, "/assigned/by/given", seen)
}

func TestRun_BeforeWhenHooksRunBetweenGivenAndWhen(t *testing.T) {
	var order []string
	Run(t, Scenario{
		Name: "test",
		Given: func(ctx *Context) {
			order = append(order, "given")
			ctx.BeforeWhen(func() { order = append(order, "flush1") })
			ctx.BeforeWhen(func() { order = append(order, "flush2") })
		},
		When: func(ctx *Context) { order = append(order, "when") },
		Then: []func(*Context){
			func(ctx *Context) { order = append(order, "then") },
		},
	})
	assert.Equal(t, []string{"given", "flush1", "flush2", "when", "then"}, order)
}

func TestRun_BeforeWhenRunsWithoutAWhen(t *testing.T) {
	flushed := false
	Run(t, Scenario{
		Name:  "test",
		Given: func(ctx *Context) { ctx.BeforeWhen(func() { flushed = true }) },
	})
	assert.True(t, flushed, "a registered flush must run even when the scenario has no When")
}

func TestRun_BeforeWhenIgnoresNilAndEmpty(t *testing.T) {
	assert.NotPanics(t, func() {
		Run(t, Scenario{
			Name:  "test",
			Given: func(ctx *Context) { ctx.BeforeWhen(nil) },
			When:  func(ctx *Context) {},
		})
	})
}

func TestRun_StateIsUsableAndScenarioScoped(t *testing.T) {
	Run(t, Scenario{
		Name: "first",
		Given: func(ctx *Context) {
			assert.NotNil(t, ctx.State)
			assert.Empty(t, ctx.State)
			ctx.State["key"] = "value"
		},
	})
	Run(t, Scenario{
		Name: "second",
		Given: func(ctx *Context) {
			assert.Empty(t, ctx.State, "state must not leak between scenarios")
		},
	})
}

func TestEvents_AppliesInOrderAndSkipsNil(t *testing.T) {
	var order []string
	composed := Events(
		func(ctx *Context) { order = append(order, "one") },
		nil,
		func(ctx *Context) { order = append(order, "two") },
	)
	Run(t, Scenario{Name: "test", Given: composed})
	assert.Equal(t, []string{"one", "two"}, order)
}

func TestEvents_WithNoEventsIsANoOp(t *testing.T) {
	assert.NotPanics(t, func() {
		Run(t, Scenario{Name: "test", Given: Events()})
	})
}
