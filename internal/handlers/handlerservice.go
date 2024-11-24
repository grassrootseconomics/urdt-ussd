package handlers

import (
	"context"

	"git.defalsify.org/vise.git/asm"
	"git.defalsify.org/vise.git/db"
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"

	"git.grassecon.net/urdt/ussd/internal/handlers/ussd"
	"git.grassecon.net/urdt/ussd/internal/utils"
	"git.grassecon.net/urdt/ussd/remote"
)

type HandlerService interface {
	GetHandler() (*ussd.Handlers, error)
}

func getParser(fp string, debug bool) (*asm.FlagParser, error) {
	flagParser := asm.NewFlagParser().WithDebug()
	_, err := flagParser.Load(fp)
	if err != nil {
		return nil, err
	}
	return flagParser, nil
}

type LocalHandlerService struct {
	Parser        *asm.FlagParser
	DbRs          *resource.DbResource
	Pe            *persist.Persister
	UserdataStore *db.Db
	AdminStore    *utils.AdminStore
	Cfg           engine.Config
	Rs            resource.Resource
}

func NewLocalHandlerService(ctx context.Context, fp string, debug bool, dbResource *resource.DbResource, cfg engine.Config, rs resource.Resource) (*LocalHandlerService, error) {
	parser, err := getParser(fp, debug)
	if err != nil {
		return nil, err
	}
	adminstore, err := utils.NewAdminStore(ctx, "admin_numbers")
	if err != nil {
		return nil, err
	}
	return &LocalHandlerService{
		Parser:     parser,
		DbRs:       dbResource,
		AdminStore: adminstore,
		Cfg:        cfg,
		Rs:         rs,
	}, nil
}

func (ls *LocalHandlerService) SetPersister(Pe *persist.Persister) {
	ls.Pe = Pe
}

func (ls *LocalHandlerService) SetDataStore(db *db.Db) {
	ls.UserdataStore = db
}

func (ls *LocalHandlerService) GetHandler(accountService remote.AccountServiceInterface) (*ussd.Handlers, error) {
	ussdHandlers, err := ussd.NewHandlers(ls.Parser, *ls.UserdataStore, ls.AdminStore, accountService)
	if err != nil {
		return nil, err
	}
	ussdHandlers = ussdHandlers.WithPersister(ls.Pe)
	ls.DbRs.AddLocalFunc("set_language", ussdHandlers.SetLanguage)
	ls.DbRs.AddLocalFunc("create_account", ussdHandlers.CreateAccount)
	ls.DbRs.AddLocalFunc("save_temporary_pin", ussdHandlers.SaveTemporaryPin)
	ls.DbRs.AddLocalFunc("verify_create_pin", ussdHandlers.VerifyCreatePin)
	ls.DbRs.AddLocalFunc("check_identifier", ussdHandlers.CheckIdentifier)
	ls.DbRs.AddLocalFunc("check_account_status", ussdHandlers.CheckAccountStatus)
	ls.DbRs.AddLocalFunc("authorize_account", ussdHandlers.Authorize)
	ls.DbRs.AddLocalFunc("quit", ussdHandlers.Quit)
	ls.DbRs.AddLocalFunc("check_balance", ussdHandlers.CheckBalance)
	ls.DbRs.AddLocalFunc("validate_recipient", ussdHandlers.ValidateRecipient)
	ls.DbRs.AddLocalFunc("transaction_reset", ussdHandlers.TransactionReset)
	ls.DbRs.AddLocalFunc("invite_valid_recipient", ussdHandlers.InviteValidRecipient)
	ls.DbRs.AddLocalFunc("max_amount", ussdHandlers.MaxAmount)
	ls.DbRs.AddLocalFunc("validate_amount", ussdHandlers.ValidateAmount)
	ls.DbRs.AddLocalFunc("reset_transaction_amount", ussdHandlers.ResetTransactionAmount)
	ls.DbRs.AddLocalFunc("get_recipient", ussdHandlers.GetRecipient)
	ls.DbRs.AddLocalFunc("get_sender", ussdHandlers.GetSender)
	ls.DbRs.AddLocalFunc("get_amount", ussdHandlers.GetAmount)
	ls.DbRs.AddLocalFunc("reset_incorrect", ussdHandlers.ResetIncorrectPin)
	ls.DbRs.AddLocalFunc("save_firstname", ussdHandlers.SaveFirstname)
	ls.DbRs.AddLocalFunc("save_familyname", ussdHandlers.SaveFamilyname)
	ls.DbRs.AddLocalFunc("save_gender", ussdHandlers.SaveGender)
	ls.DbRs.AddLocalFunc("save_location", ussdHandlers.SaveLocation)
	ls.DbRs.AddLocalFunc("save_yob", ussdHandlers.SaveYob)
	ls.DbRs.AddLocalFunc("save_offerings", ussdHandlers.SaveOfferings)
	ls.DbRs.AddLocalFunc("reset_account_authorized", ussdHandlers.ResetAccountAuthorized)
	ls.DbRs.AddLocalFunc("reset_allow_update", ussdHandlers.ResetAllowUpdate)
	ls.DbRs.AddLocalFunc("get_profile_info", ussdHandlers.GetProfileInfo)
	ls.DbRs.AddLocalFunc("verify_yob", ussdHandlers.VerifyYob)
	ls.DbRs.AddLocalFunc("reset_incorrect_date_format", ussdHandlers.ResetIncorrectYob)
	ls.DbRs.AddLocalFunc("initiate_transaction", ussdHandlers.InitiateTransaction)
	ls.DbRs.AddLocalFunc("verify_new_pin", ussdHandlers.VerifyNewPin)
	ls.DbRs.AddLocalFunc("confirm_pin_change", ussdHandlers.ConfirmPinChange)
	ls.DbRs.AddLocalFunc("quit_with_help", ussdHandlers.QuitWithHelp)
	ls.DbRs.AddLocalFunc("fetch_community_balance", ussdHandlers.FetchCommunityBalance)
	ls.DbRs.AddLocalFunc("set_default_voucher", ussdHandlers.SetDefaultVoucher)
	ls.DbRs.AddLocalFunc("check_vouchers", ussdHandlers.CheckVouchers)
	ls.DbRs.AddLocalFunc("get_vouchers", ussdHandlers.GetVoucherList)
	ls.DbRs.AddLocalFunc("view_voucher", ussdHandlers.ViewVoucher)
	ls.DbRs.AddLocalFunc("set_voucher", ussdHandlers.SetVoucher)
	ls.DbRs.AddLocalFunc("get_voucher_details", ussdHandlers.GetVoucherDetails)
	ls.DbRs.AddLocalFunc("reset_valid_pin", ussdHandlers.ResetValidPin)
	ls.DbRs.AddLocalFunc("check_pin_mismatch", ussdHandlers.CheckPinMisMatch)
	ls.DbRs.AddLocalFunc("validate_blocked_number", ussdHandlers.ValidateBlockedNumber)
	ls.DbRs.AddLocalFunc("retrieve_blocked_number", ussdHandlers.RetrieveBlockedNumber)
	ls.DbRs.AddLocalFunc("reset_unregistered_number", ussdHandlers.ResetUnregisteredNumber)
	ls.DbRs.AddLocalFunc("reset_others_pin", ussdHandlers.ResetOthersPin)
	ls.DbRs.AddLocalFunc("save_others_temporary_pin", ussdHandlers.SaveOthersTemporaryPin)
	ls.DbRs.AddLocalFunc("get_current_profile_info", ussdHandlers.GetCurrentProfileInfo)
	ls.DbRs.AddLocalFunc("check_transactions", ussdHandlers.CheckTransactions)
	ls.DbRs.AddLocalFunc("get_transactions", ussdHandlers.GetTransactionsList)
	ls.DbRs.AddLocalFunc("view_statement", ussdHandlers.ViewTransactionStatement)

	return ussdHandlers, nil
}

// TODO: enable setting of sessionId on engine init time
func (ls *LocalHandlerService) GetEngine() *engine.DefaultEngine {
	en := engine.NewEngine(ls.Cfg, ls.Rs)
	en = en.WithPersister(ls.Pe)
	return en
}
