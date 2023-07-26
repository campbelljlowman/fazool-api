// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/campbelljlowman/fazool-api/account (interfaces: AccountService)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	model "github.com/campbelljlowman/fazool-api/graph/model"
	gomock "go.uber.org/mock/gomock"
)

// MockAccountService is a mock of AccountService interface.
type MockAccountService struct {
	ctrl     *gomock.Controller
	recorder *MockAccountServiceMockRecorder
}

// MockAccountServiceMockRecorder is the mock recorder for MockAccountService.
type MockAccountServiceMockRecorder struct {
	mock *MockAccountService
}

// NewMockAccountService creates a new mock instance.
func NewMockAccountService(ctrl *gomock.Controller) *MockAccountService {
	mock := &MockAccountService{ctrl: ctrl}
	mock.recorder = &MockAccountServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAccountService) EXPECT() *MockAccountServiceMockRecorder {
	return m.recorder
}

// AddBonusVotes mocks base method.
func (m *MockAccountService) AddBonusVotes(arg0, arg1 int) *model.Account {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddBonusVotes", arg0, arg1)
	ret0, _ := ret[0].(*model.Account)
	return ret0
}

// AddBonusVotes indicates an expected call of AddBonusVotes.
func (mr *MockAccountServiceMockRecorder) AddBonusVotes(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddBonusVotes", reflect.TypeOf((*MockAccountService)(nil).AddBonusVotes), arg0, arg1)
}

// CheckIfEmailHasAccount mocks base method.
func (m *MockAccountService) CheckIfEmailHasAccount(arg0 string) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckIfEmailHasAccount", arg0)
	ret0, _ := ret[0].(bool)
	return ret0
}

// CheckIfEmailHasAccount indicates an expected call of CheckIfEmailHasAccount.
func (mr *MockAccountServiceMockRecorder) CheckIfEmailHasAccount(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckIfEmailHasAccount", reflect.TypeOf((*MockAccountService)(nil).CheckIfEmailHasAccount), arg0)
}

// CreateAccount mocks base method.
func (m *MockAccountService) CreateAccount(arg0, arg1, arg2, arg3 string, arg4 model.AccountType, arg5 model.VoterType, arg6 int, arg7 model.StreamingService) int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateAccount", arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7)
	ret0, _ := ret[0].(int)
	return ret0
}

// CreateAccount indicates an expected call of CreateAccount.
func (mr *MockAccountServiceMockRecorder) CreateAccount(arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateAccount", reflect.TypeOf((*MockAccountService)(nil).CreateAccount), arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7)
}

// DeleteAccount mocks base method.
func (m *MockAccountService) DeleteAccount(arg0 int) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "DeleteAccount", arg0)
}

// DeleteAccount indicates an expected call of DeleteAccount.
func (mr *MockAccountServiceMockRecorder) DeleteAccount(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteAccount", reflect.TypeOf((*MockAccountService)(nil).DeleteAccount), arg0)
}

// GetAccountActiveSession mocks base method.
func (m *MockAccountService) GetAccountActiveSession(arg0 int) int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAccountActiveSession", arg0)
	ret0, _ := ret[0].(int)
	return ret0
}

// GetAccountActiveSession indicates an expected call of GetAccountActiveSession.
func (mr *MockAccountServiceMockRecorder) GetAccountActiveSession(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAccountActiveSession", reflect.TypeOf((*MockAccountService)(nil).GetAccountActiveSession), arg0)
}

// GetAccountFromEmail mocks base method.
func (m *MockAccountService) GetAccountFromEmail(arg0 string) *model.Account {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAccountFromEmail", arg0)
	ret0, _ := ret[0].(*model.Account)
	return ret0
}

// GetAccountFromEmail indicates an expected call of GetAccountFromEmail.
func (mr *MockAccountServiceMockRecorder) GetAccountFromEmail(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAccountFromEmail", reflect.TypeOf((*MockAccountService)(nil).GetAccountFromEmail), arg0)
}

// GetAccountFromID mocks base method.
func (m *MockAccountService) GetAccountFromID(arg0 int) *model.Account {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAccountFromID", arg0)
	ret0, _ := ret[0].(*model.Account)
	return ret0
}

// GetAccountFromID indicates an expected call of GetAccountFromID.
func (mr *MockAccountServiceMockRecorder) GetAccountFromID(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAccountFromID", reflect.TypeOf((*MockAccountService)(nil).GetAccountFromID), arg0)
}

// GetAccountIDAndPassHash mocks base method.
func (m *MockAccountService) GetAccountIDAndPassHash(arg0 string) (int, string) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAccountIDAndPassHash", arg0)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(string)
	return ret0, ret1
}

// GetAccountIDAndPassHash indicates an expected call of GetAccountIDAndPassHash.
func (mr *MockAccountServiceMockRecorder) GetAccountIDAndPassHash(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAccountIDAndPassHash", reflect.TypeOf((*MockAccountService)(nil).GetAccountIDAndPassHash), arg0)
}

// GetAccountType mocks base method.
func (m *MockAccountService) GetAccountType(arg0 int) model.AccountType {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAccountType", arg0)
	ret0, _ := ret[0].(model.AccountType)
	return ret0
}

// GetAccountType indicates an expected call of GetAccountType.
func (mr *MockAccountServiceMockRecorder) GetAccountType(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAccountType", reflect.TypeOf((*MockAccountService)(nil).GetAccountType), arg0)
}

// GetSpotifyRefreshToken mocks base method.
func (m *MockAccountService) GetSpotifyRefreshToken(arg0 int) string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSpotifyRefreshToken", arg0)
	ret0, _ := ret[0].(string)
	return ret0
}

// GetSpotifyRefreshToken indicates an expected call of GetSpotifyRefreshToken.
func (mr *MockAccountServiceMockRecorder) GetSpotifyRefreshToken(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSpotifyRefreshToken", reflect.TypeOf((*MockAccountService)(nil).GetSpotifyRefreshToken), arg0)
}

// GetVoterTypeAndBonusVotes mocks base method.
func (m *MockAccountService) GetVoterTypeAndBonusVotes(arg0 int) (model.VoterType, int) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetVoterTypeAndBonusVotes", arg0)
	ret0, _ := ret[0].(model.VoterType)
	ret1, _ := ret[1].(int)
	return ret0, ret1
}

// GetVoterTypeAndBonusVotes indicates an expected call of GetVoterTypeAndBonusVotes.
func (mr *MockAccountServiceMockRecorder) GetVoterTypeAndBonusVotes(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVoterTypeAndBonusVotes", reflect.TypeOf((*MockAccountService)(nil).GetVoterTypeAndBonusVotes), arg0)
}

// SetAccountActiveSession mocks base method.
func (m *MockAccountService) SetAccountActiveSession(arg0, arg1 int) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetAccountActiveSession", arg0, arg1)
}

// SetAccountActiveSession indicates an expected call of SetAccountActiveSession.
func (mr *MockAccountServiceMockRecorder) SetAccountActiveSession(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetAccountActiveSession", reflect.TypeOf((*MockAccountService)(nil).SetAccountActiveSession), arg0, arg1)
}

// SetAccountType mocks base method.
func (m *MockAccountService) SetAccountType(arg0 int, arg1 model.AccountType) *model.Account {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetAccountType", arg0, arg1)
	ret0, _ := ret[0].(*model.Account)
	return ret0
}

// SetAccountType indicates an expected call of SetAccountType.
func (mr *MockAccountServiceMockRecorder) SetAccountType(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetAccountType", reflect.TypeOf((*MockAccountService)(nil).SetAccountType), arg0, arg1)
}

// SetSpotifyRefreshToken mocks base method.
func (m *MockAccountService) SetSpotifyRefreshToken(arg0 int, arg1 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetSpotifyRefreshToken", arg0, arg1)
}

// SetSpotifyRefreshToken indicates an expected call of SetSpotifyRefreshToken.
func (mr *MockAccountServiceMockRecorder) SetSpotifyRefreshToken(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetSpotifyRefreshToken", reflect.TypeOf((*MockAccountService)(nil).SetSpotifyRefreshToken), arg0, arg1)
}

// SetVoterType mocks base method.
func (m *MockAccountService) SetVoterType(arg0 int, arg1 model.VoterType) *model.Account {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetVoterType", arg0, arg1)
	ret0, _ := ret[0].(*model.Account)
	return ret0
}

// SetVoterType indicates an expected call of SetVoterType.
func (mr *MockAccountServiceMockRecorder) SetVoterType(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetVoterType", reflect.TypeOf((*MockAccountService)(nil).SetVoterType), arg0, arg1)
}

// SubtractBonusVotes mocks base method.
func (m *MockAccountService) SubtractBonusVotes(arg0, arg1 int) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SubtractBonusVotes", arg0, arg1)
}

// SubtractBonusVotes indicates an expected call of SubtractBonusVotes.
func (mr *MockAccountServiceMockRecorder) SubtractBonusVotes(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubtractBonusVotes", reflect.TypeOf((*MockAccountService)(nil).SubtractBonusVotes), arg0, arg1)
}
