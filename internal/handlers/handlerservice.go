package handlers

import (
	"context"
	"strings"

	"git.defalsify.org/vise.git/asm"
	"git.defalsify.org/vise.git/db"
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"

	"git.grassecon.net/urdt/ussd/internal/handlers/application"
	"git.grassecon.net/urdt/ussd/internal/utils"
	"git.grassecon.net/urdt/ussd/remote"
)

type HandlerService interface {
	GetHandler() (*application.Handlers, error)
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

func (ls *LocalHandlerService) GetHandler(accountService remote.AccountServiceInterface) (*application.Handlers, error) {
	replaceSeparatorFunc := func(input string) string {
		return strings.ReplaceAll(input, ":", ls.Cfg.MenuSeparator)
	}

	appHandlers, err := application.NewHandlers(ls.Parser, *ls.UserdataStore, ls.AdminStore, accountService, replaceSeparatorFunc)
	if err != nil {
		return nil, err
	}
	appHandlers = appHandlers.WithPersister(ls.Pe)
	ls.DbRs.AddLocalFunc("set_language", appHandlers.SetLanguage)
	ls.DbRs.AddLocalFunc("create_account", appHandlers.CreateAccount)
	ls.DbRs.AddLocalFunc("save_temporary_pin", appHandlers.SaveTemporaryPin)
	ls.DbRs.AddLocalFunc("verify_create_pin", appHandlers.VerifyCreatePin)
	ls.DbRs.AddLocalFunc("check_identifier", appHandlers.CheckIdentifier)
	ls.DbRs.AddLocalFunc("check_account_status", appHandlers.CheckAccountStatus)
	ls.DbRs.AddLocalFunc("authorize_account", appHandlers.Authorize)
	ls.DbRs.AddLocalFunc("quit", appHandlers.Quit)
	ls.DbRs.AddLocalFunc("check_balance", appHandlers.CheckBalance)
	ls.DbRs.AddLocalFunc("validate_recipient", appHandlers.ValidateRecipient)
	ls.DbRs.AddLocalFunc("transaction_reset", appHandlers.TransactionReset)
	ls.DbRs.AddLocalFunc("invite_valid_recipient", appHandlers.InviteValidRecipient)
	ls.DbRs.AddLocalFunc("max_amount", appHandlers.MaxAmount)
	ls.DbRs.AddLocalFunc("validate_amount", appHandlers.ValidateAmount)
	ls.DbRs.AddLocalFunc("reset_transaction_amount", appHandlers.ResetTransactionAmount)
	ls.DbRs.AddLocalFunc("get_recipient", appHandlers.GetRecipient)
	ls.DbRs.AddLocalFunc("get_sender", appHandlers.GetSender)
	ls.DbRs.AddLocalFunc("get_amount", appHandlers.GetAmount)
	ls.DbRs.AddLocalFunc("reset_incorrect", appHandlers.ResetIncorrectPin)
	ls.DbRs.AddLocalFunc("save_firstname", appHandlers.SaveFirstname)
	ls.DbRs.AddLocalFunc("save_familyname", appHandlers.SaveFamilyname)
	ls.DbRs.AddLocalFunc("save_gender", appHandlers.SaveGender)
	ls.DbRs.AddLocalFunc("save_location", appHandlers.SaveLocation)
	ls.DbRs.AddLocalFunc("save_yob", appHandlers.SaveYob)
	ls.DbRs.AddLocalFunc("save_offerings", appHandlers.SaveOfferings)
	ls.DbRs.AddLocalFunc("reset_account_authorized", appHandlers.ResetAccountAuthorized)
	ls.DbRs.AddLocalFunc("reset_allow_update", appHandlers.ResetAllowUpdate)
	ls.DbRs.AddLocalFunc("get_profile_info", appHandlers.GetProfileInfo)
	ls.DbRs.AddLocalFunc("verify_yob", appHandlers.VerifyYob)
	ls.DbRs.AddLocalFunc("reset_incorrect_date_format", appHandlers.ResetIncorrectYob)
	ls.DbRs.AddLocalFunc("initiate_transaction", appHandlers.InitiateTransaction)
	ls.DbRs.AddLocalFunc("verify_new_pin", appHandlers.VerifyNewPin)
	ls.DbRs.AddLocalFunc("confirm_pin_change", appHandlers.ConfirmPinChange)
	ls.DbRs.AddLocalFunc("quit_with_help", appHandlers.QuitWithHelp)
	ls.DbRs.AddLocalFunc("fetch_community_balance", appHandlers.FetchCommunityBalance)
	ls.DbRs.AddLocalFunc("set_default_voucher", appHandlers.SetDefaultVoucher)
	ls.DbRs.AddLocalFunc("check_vouchers", appHandlers.CheckVouchers)
	ls.DbRs.AddLocalFunc("get_vouchers", appHandlers.GetVoucherList)
	ls.DbRs.AddLocalFunc("view_voucher", appHandlers.ViewVoucher)
	ls.DbRs.AddLocalFunc("set_voucher", appHandlers.SetVoucher)
	ls.DbRs.AddLocalFunc("get_voucher_details", appHandlers.GetVoucherDetails)
	ls.DbRs.AddLocalFunc("reset_valid_pin", appHandlers.ResetValidPin)
	ls.DbRs.AddLocalFunc("check_pin_mismatch", appHandlers.CheckBlockedNumPinMisMatch)
	ls.DbRs.AddLocalFunc("validate_blocked_number", appHandlers.ValidateBlockedNumber)
	ls.DbRs.AddLocalFunc("retrieve_blocked_number", appHandlers.RetrieveBlockedNumber)
	ls.DbRs.AddLocalFunc("reset_unregistered_number", appHandlers.ResetUnregisteredNumber)
	ls.DbRs.AddLocalFunc("reset_others_pin", appHandlers.ResetOthersPin)
	ls.DbRs.AddLocalFunc("save_others_temporary_pin", appHandlers.SaveOthersTemporaryPin)
	ls.DbRs.AddLocalFunc("get_current_profile_info", appHandlers.GetCurrentProfileInfo)
	ls.DbRs.AddLocalFunc("check_transactions", appHandlers.CheckTransactions)
	ls.DbRs.AddLocalFunc("get_transactions", appHandlers.GetTransactionsList)
	ls.DbRs.AddLocalFunc("view_statement", appHandlers.ViewTransactionStatement)
	ls.DbRs.AddLocalFunc("update_all_profile_items", appHandlers.UpdateAllProfileItems)
	ls.DbRs.AddLocalFunc("set_back", appHandlers.SetBack)
	ls.DbRs.AddLocalFunc("show_blocked_account", appHandlers.ShowBlockedAccount)

	return appHandlers, nil
}

// TODO: enable setting of sessionId on engine init time
func (ls *LocalHandlerService) GetEngine() *engine.DefaultEngine {
	en := engine.NewEngine(ls.Cfg, ls.Rs)
	en = en.WithPersister(ls.Pe)
	return en
}
