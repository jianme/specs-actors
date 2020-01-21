package runtime

import (
	addr "github.com/filecoin-project/go-address"
	cid "github.com/ipfs/go-cid"

	abi "github.com/filecoin-project/specs-actors/actors/abi"
	crypto "github.com/filecoin-project/specs-actors/actors/crypto"
	exitcode "github.com/filecoin-project/specs-actors/actors/runtime/exitcode"
	indices "github.com/filecoin-project/specs-actors/actors/runtime/indices"
)

// Runtime is the VM's internal runtime object.
// this is everything that is accessible to actors, beyond parameters.
type Runtime interface {
	CurrEpoch() abi.ChainEpoch

	// Randomness returns a (pseudo)random string for the given epoch and tag.
	GetRandomness(epoch abi.ChainEpoch) abi.RandomnessSeed

	// The address of the immediate calling actor.
	// Not necessarily the actor in the From field of the initial on-chain Message.
	// Always an ID-address.
	ImmediateCaller() addr.Address
	ValidateImmediateCallerAcceptAny()
	ValidateImmediateCallerIs(addrs ...addr.Address)
	ValidateImmediateCallerType(types ...abi.ActorCodeID)

	// The address of the actor receiving the message. Always an ID-address.
	CurrReceiver() addr.Address

	// The actor who mined the block in which the initial on-chain message appears.
	// Always an ID-address.
	ToplevelBlockWinner() addr.Address

	AcquireState() ActorStateHandle

	// Throw an error indicating a failure condition has occurred, from which the given actor
	// code is unable to recover.
	Abort(errExitCode exitcode.ExitCode, msg string)

	// Calls Abort with InvalidArguments_User.
	AbortArgMsg(msg string)

	// Calls Abort with InconsistentState_User.
	AbortStateMsg(msg string)

	// Calls Abort with InsufficientFunds_User.
	AbortFundsMsg(msg string)

	// Calls Abort with RuntimeAPIError.
	// For internal use only (not in actor code).
	AbortAPI(msg string)

	CurrentBalance() abi.TokenAmount
	ValueReceived() abi.TokenAmount

	// Look up the current values of several system-wide economic indices.
	CurrIndices() indices.Indices

	// Look up the code ID of a given actor address.
	GetActorCodeID(addr addr.Address) (ret abi.ActorCodeID, ok bool)

	// Run a (pure function) computation, consuming the gas cost associated with that function.
	// This mechanism is intended to capture the notion of an ABI between the VM and native
	// functions, and should be used for any function whose computation is expensive.
	Compute(ComputeFunctionID, args []interface{}) interface{}

	// Sends a message to another actor.
	// If the invoked method does not return successfully, this caller will be aborted too.
	Send(toAddr addr.Address, methodNum abi.MethodNum, params abi.MethodParams, value abi.TokenAmount) SendReturn
	SendQuery(toAddr addr.Address, methodNum abi.MethodNum, params abi.MethodParams) SendReturn
	SendFunds(toAddr addr.Address, value abi.TokenAmount)

	// Sends a message to another actor, trapping an unsuccessful execution.
	// This may only be invoked by the singleton Cron actor.
	SendCatchingErrors(input InvocInput) (ret SendReturn, exitCode exitcode.ExitCode)

	// Computes an address for a new actor. The returned address is intended to uniquely refer to
	// the actor even in the event of a chain re-org (whereas an ID-address might refer to a
	// different actor after messages are re-ordered).
	// Always an ActorExec address.
	NewActorAddress() addr.Address

	// Creates an actor with code `codeID` and address `address`, with empty state. May only be called by InitActor.
	CreateActor(codeId abi.ActorCodeID, address addr.Address)

	// Deletes an actor in the state tree. May only be called by the actor itself,
	// or by StoragePowerActor in the case of StorageMinerActors.
	DeleteActor(address addr.Address)

	// Retrieves and deserializes an object from the store into o. Returns whether successful.
	IpldGet(c cid.Cid, o interface{}) bool
	// Serializes and stores an object, returning its CID.
	IpldPut(x interface{}) cid.Cid

	// Provides the system call interface.
	Syscalls() Syscalls
}

type Syscalls interface {
	// Verifies that a signature is valid for an address and plaintext.
	VerifySignature(signature crypto.Signature, signer addr.Address, plaintext []byte) bool
	// Computes an unsealed sector CID (CommD) from its constituent piece CIDs (CommPs) and sizes.
	ComputeUnsealedSectorCID(sectorSize abi.SectorSize, pieces []abi.PieceInfo) (abi.UnsealedSectorCID, error)
	// Verifies a sector seal proof.
	VerifySeal(sectorSize abi.SectorSize, vi abi.SealVerifyInfo) bool
	// Verifies a proof of spacetime.
	VerifyPoSt(sectorSize abi.SectorSize, vi abi.PoStVerifyInfo) bool
}

// The return type from a message send from one actor to another. This abstracts over the internal representation of
// the return, in particular whether it has been serialized to bytes or just passed through.
// Production code is expected to de/serialize, but test and other code may pass the value straight through.
type SendReturn interface {
	Into(interface{}) error
}

type InvocInput struct {
	To     addr.Address
	Method abi.MethodNum
	Params abi.MethodParams
	Value  abi.TokenAmount
}

type ActorStateHandle interface {
	UpdateRelease(newStateCID abi.ActorSubstateCID)
	Release(checkStateCID abi.ActorSubstateCID)
	Take() abi.ActorSubstateCID
}

type ComputeFunctionID int64

const (
	Compute_VerifySignature = ComputeFunctionID(1)
)