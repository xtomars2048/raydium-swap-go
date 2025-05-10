package trade

import (
	"context"

	ag_solanago "github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/xtomars2048/raydium-swap-go/raydium/constants"
	"github.com/xtomars2048/raydium-swap-go/raydium/layouts"
	"github.com/xtomars2048/raydium-swap-go/raydium/utils"
)

type Trade struct {
	Connection *rpc.Client
	SignerPub  *ag_solanago.PublicKey
}

type FeeConfig struct {
	MicroLamports uint64
}

func New(connection *rpc.Client, signerP *ag_solanago.PublicKey) *Trade {
	return &Trade{
		Connection: connection,
		SignerPub:  signerP,
	}
}

func (t *Trade) MakeRawSwapTx(poolKeys *layouts.ApiPoolInfoV4, amountIn *utils.TokenAmount, minAmountOut *utils.TokenAmount, feeConfig FeeConfig, faultAdd string, feeValue uint64) (*ag_solanago.Transaction, error) {
	recent, err := t.Connection.GetLatestBlockhash(context.Background(), rpc.CommitmentFinalized)

	if err != nil {
		return &ag_solanago.Transaction{}, err
	}

	var instructions []ag_solanago.Instruction = []ag_solanago.Instruction{
		computebudget.NewSetComputeUnitLimitInstruction(600000).Build(),
		computebudget.NewSetComputeUnitPriceInstruction(feeConfig.MicroLamports).Build(),
	}

	tokenAccountIn, err := t.selectOrCreateAccount(amountIn, &instructions, "in")

	if err != nil {
		return &ag_solanago.Transaction{}, err
	}

	tokenAccountOut, err := t.selectOrCreateAccount(minAmountOut, &instructions, "out")

	if err != nil {
		return &ag_solanago.Transaction{}, err
	}
	instr, err := NewSwapV4Instruction(
		t.Connection,
		poolKeys,
		uint64(amountIn.Amount),
		uint64(minAmountOut.Amount),
		tokenAccountIn,
		tokenAccountOut,
		*t.SignerPub,
	)

	if err != nil {
		return &ag_solanago.Transaction{}, err
	}

	instructions = append(instructions, instr)

	if amountIn.Token.Mint == constants.WSOl.String() {
		closeAccInst, err := token.NewCloseAccountInstruction(
			tokenAccountIn,
			*t.SignerPub,
			*t.SignerPub,
			[]ag_solanago.PublicKey{},
		).ValidateAndBuild()

		if err != nil {
			return &ag_solanago.Transaction{}, err
		}

		instructions = append(instructions, closeAccInst)
	} else if minAmountOut.Token.Mint == constants.WSOl.String() {
		closeAccInst, err := token.NewCloseAccountInstruction(
			tokenAccountOut,
			*t.SignerPub,
			*t.SignerPub,
			[]ag_solanago.PublicKey{},
		).ValidateAndBuild()

		if err != nil {
			return &ag_solanago.Transaction{}, err
		}

		instructions = append(instructions, closeAccInst)
	}

	if faultAdd != "" && feeValue > 0 {
		recipientAccount := ag_solanago.MustPublicKeyFromBase58(faultAdd)

		payFaultInst := system.NewTransferInstruction(feeValue, *t.SignerPub, recipientAccount).Build()
		instructions = append(instructions, payFaultInst)
	}

	tx, err := ag_solanago.NewTransaction(
		instructions,
		recent.Value.Blockhash,
		ag_solanago.TransactionPayer(*t.SignerPub),
	)

	return tx, err
}

func (t *Trade) selectOrCreateAccount(amount *utils.TokenAmount, insttr *[]ag_solanago.Instruction, side string) (ag_solanago.PublicKey, error) {
	acc, err := t.Connection.GetTokenAccountsByOwner(context.Background(), *t.SignerPub, &rpc.GetTokenAccountsConfig{Mint: amount.Token.PublicKey().ToPointer()}, &rpc.GetTokenAccountsOpts{
		Encoding: "jsonParsed",
	})
	if err != nil {
		return ag_solanago.PublicKey{}, err
	}

	if len(acc.Value) > 0 {
		return acc.Value[0].Pubkey, nil
	}

	ataAddress, _, err := ag_solanago.FindAssociatedTokenAddress(*t.SignerPub, amount.Token.PublicKey())

	if err != nil {
		return ag_solanago.PublicKey{}, err
	}

	rentCost, err := t.Connection.GetMinimumBalanceForRentExemption(context.Background(), 165, rpc.CommitmentConfirmed)

	if err != nil {
		return ag_solanago.PublicKey{}, err
	}

	accountLamports := rentCost

	if amount.Mint == constants.WSOl.String() {
		if side == "in" {
			accountLamports += uint64(amount.Amount)
		}

		pubKey, seed, err := t.generatePubkeyWithSeed(*t.SignerPub, token.ProgramID)

		if err != nil {
			return ag_solanago.PublicKey{}, err
		}
		createInst, err := system.NewCreateAccountWithSeedInstruction(
			*t.SignerPub,
			seed,
			accountLamports,
			165,
			token.ProgramID,
			*t.SignerPub,
			pubKey,
			*t.SignerPub,
		).ValidateAndBuild()

		if err != nil {
			return ag_solanago.PublicKey{}, err
		}

		initInst, err := token.NewInitializeAccountInstruction(
			pubKey,
			constants.WSOl,
			*t.SignerPub,
			ag_solanago.SysVarRentPubkey,
		).ValidateAndBuild()

		if err != nil {
			return ag_solanago.PublicKey{}, err
		}

		*insttr = append(*insttr, createInst)
		*insttr = append(*insttr, initInst)

		return pubKey, nil
	}

	createAtaInst, err := associatedtokenaccount.NewCreateInstruction(
		*t.SignerPub,
		*t.SignerPub,
		amount.Token.PublicKey(),
	).ValidateAndBuild()

	if err != nil {
		return ag_solanago.PublicKey{}, err
	}
	*insttr = append(*insttr, createAtaInst)

	return ataAddress, nil
}

func (t *Trade) generatePubkeyWithSeed(from ag_solanago.PublicKey, programId ag_solanago.PublicKey) (ag_solanago.PublicKey, string, error) {
	seed := ag_solanago.NewWallet().PublicKey().String()[0:32]
	publicKey, err := ag_solanago.CreateWithSeed(from, seed, programId)

	return publicKey, seed, err
}
