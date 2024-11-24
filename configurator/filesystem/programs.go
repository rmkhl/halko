package filesystem

import (
	"encoding/json"

	"github.com/rmkhl/halko/types"
)

type (
	program struct {
		*types.Program
	}

	programs struct{}
)

func (p *programs) ByName(name string) (*types.Program, error) {
	prog, err := byName(name, new(program))
	if err != nil {
		return nil, transformError(err)
	}
	return runtimeCast[types.Program](prog)
}

func (p *programs) All() ([]*types.Program, error) {
	progs, err := all(new(program))
	if err != nil {
		return nil, transformError(err)
	}
	return runtimeCastSlice[types.Program](progs)
}

func (p *programs) CreateOrUpdate(name string, pp *types.Program) (*types.Program, error) {
	ppp, err := save(name, &program{pp})
	if err != nil {
		return nil, transformError(err)
	}
	cast, err := runtimeCast[program](ppp)
	if err != nil {
		return nil, transformError(err)
	}
	return cast.Program, nil
}

func (p *program) name() string {
	return string(p.ProgramName)
}

func (p *program) setName(name string) {
	p.ProgramName = name
}

func (p *program) unmarshalJSON(data []byte) (any, error) {
	var prog types.Program

	if err := json.Unmarshal(data, &prog); err != nil {
		return nil, err
	}

	return &prog, nil
}
