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
