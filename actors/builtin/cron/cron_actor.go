package cron

import (
	addr "github.com/filecoin-project/go-address"

	abi "github.com/filecoin-project/specs-actors/actors/abi"
	builtin "github.com/filecoin-project/specs-actors/actors/builtin"
	vmr "github.com/filecoin-project/specs-actors/actors/runtime"
)

type CronActorState struct{}

type CronActor struct {
	// TODO move Entries into the CronActorState struct
	Entries []CronTableEntry
}

type CronTableEntry struct {
	ToAddr    addr.Address
	MethodNum abi.MethodNum
}

func (a *CronActor) Constructor(rt vmr.Runtime) *vmr.EmptyReturn {
	// Nothing. intentionally left blank.
	rt.ValidateImmediateCallerIs(builtin.SystemActorAddr)
	return &vmr.EmptyReturn{}
}

func (a *CronActor) EpochTick(rt vmr.Runtime) *vmr.EmptyReturn {
	rt.ValidateImmediateCallerIs(builtin.SystemActorAddr)

	// a.Entries is basically a static registry for now, loaded
	// in the interpreter static registry.
	for _, entry := range a.Entries {
		rt.SendCatchingErrors(vmr.InvocInput{
			To:     entry.ToAddr,
			Method: entry.MethodNum,
			Params: nil,
			Value:  abi.TokenAmount(0),
		})
	}

	return &vmr.EmptyReturn{}
}