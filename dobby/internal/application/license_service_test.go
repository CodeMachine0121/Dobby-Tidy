package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dobby/filemanager/internal/application"
	"github.com/dobby/filemanager/internal/domain/license"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ─────────────────────────────────────────────────────────────────────────────
// Mocks
// ─────────────────────────────────────────────────────────────────────────────

type mockLicenseRepo struct{ mock.Mock }

func (m *mockLicenseRepo) Load(ctx context.Context) (*license.License, bool, error) {
	args := m.Called(ctx)
	l, _ := args.Get(0).(*license.License)
	return l, args.Bool(1), args.Error(2)
}
func (m *mockLicenseRepo) Save(ctx context.Context, l *license.License) error {
	return m.Called(ctx, l).Error(0)
}

type mockMachineId struct{ mock.Mock }

func (m *mockMachineId) MachineID() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

type mockGumroadVerifier struct{ mock.Mock }

func (m *mockGumroadVerifier) Verify(ctx context.Context, key string) error {
	return m.Called(ctx, key).Error(0)
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func createLicenseService() (*application.LicenseService, *mockLicenseRepo, *mockMachineId, *mockGumroadVerifier) {
	repo := new(mockLicenseRepo)
	machineId := new(mockMachineId)
	verifier := new(mockGumroadVerifier)
	svc := application.NewLicenseService(repo, machineId, verifier)
	return svc, repo, machineId, verifier
}

func givenVerifierSucceeds(v *mockGumroadVerifier) {
	v.On("Verify", mock.Anything, mock.Anything).Return(nil)
}

func givenVerifierReturns(v *mockGumroadVerifier, err error) {
	v.On("Verify", mock.Anything, mock.Anything).Return(err)
}

func createActiveLicense(daysAgo int) *license.License {
	return license.Reconstitute(time.Now().AddDate(0, 0, -daysAgo), "", "")
}

func createExpiredLicense() *license.License {
	return license.Reconstitute(time.Now().AddDate(0, 0, -15), "", "")
}

func createActivatedLicense(daysAgo int) *license.License {
	return license.Reconstitute(time.Now().AddDate(0, 0, -daysAgo), "XXXXXXXX-XXXXXXXX-XXXXXXXX-XXXXXXXX", "test-machine-id")
}

func givenNoLicenseExists(repo *mockLicenseRepo) {
	repo.On("Load", mock.Anything).Return(nil, false, nil)
}

func givenLicenseExists(repo *mockLicenseRepo, l *license.License) {
	repo.On("Load", mock.Anything).Return(l, true, nil)
}

func givenSaveSucceeds(repo *mockLicenseRepo) {
	repo.On("Save", mock.Anything, mock.AnythingOfType("*license.License")).Return(nil)
}

func givenMachineID(m *mockMachineId) {
	m.On("MachineID").Return("test-machine-id", nil)
}

// ─────────────────────────────────────────────────────────────────────────────
// S-1: First launch initializes a 14-day trial
// ─────────────────────────────────────────────────────────────────────────────

// S-1-1: 無授權記錄時 Save 被呼叫
func TestLicenseService_InitializeTrial_WhenNoRecord_SaveIsCalled(t *testing.T) {
	// Arrange
	svc, repo, _, _ := createLicenseService()
	givenNoLicenseExists(repo)
	givenSaveSucceeds(repo)

	// Act
	err := svc.InitializeTrial(context.Background())

	// Assert
	assert.NoError(t, err)
	repo.AssertCalled(t, "Save", mock.Anything, mock.AnythingOfType("*license.License"))
}

// S-1-2: 已有授權記錄時 Save 不被呼叫（idempotent）
func TestLicenseService_InitializeTrial_WhenRecordExists_SaveIsNotCalled(t *testing.T) {
	// Arrange
	svc, repo, _, _ := createLicenseService()
	givenLicenseExists(repo, createActiveLicense(0))

	// Act
	err := svc.InitializeTrial(context.Background())

	// Assert
	assert.NoError(t, err)
	repo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
}

// ─────────────────────────────────────────────────────────────────────────────
// S-2: Trial active within 14-day window
// ─────────────────────────────────────────────────────────────────────────────

// S-2-1: 7 天前開始的試用，CanRunBackgroundProcessor 回傳 true
func TestLicenseService_CanRun_WhenTrialActive_ReturnsTrue(t *testing.T) {
	// Arrange
	svc, repo, _, _ := createLicenseService()
	givenLicenseExists(repo, createActiveLicense(7))

	// Act
	canRun, err := svc.CanRunBackgroundProcessor(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.True(t, canRun)
}

// S-2-2: 7 天前開始的試用，GetLicenseInfo 狀態為 "active"
func TestLicenseService_GetLicenseInfo_WhenTrialActive_StatusIsActive(t *testing.T) {
	// Arrange
	svc, repo, _, _ := createLicenseService()
	givenLicenseExists(repo, createActiveLicense(7))

	// Act
	info, err := svc.GetLicenseInfo(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "active", info.Status)
}

// ─────────────────────────────────────────────────────────────────────────────
// S-3: Trial expires after 14 days
// ─────────────────────────────────────────────────────────────────────────────

// S-3-1: 15 天前開始的試用，CanRunBackgroundProcessor 回傳 false
func TestLicenseService_CanRun_WhenTrialExpired_ReturnsFalse(t *testing.T) {
	// Arrange
	svc, repo, _, _ := createLicenseService()
	givenLicenseExists(repo, createExpiredLicense())

	// Act
	canRun, err := svc.CanRunBackgroundProcessor(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.False(t, canRun)
}

// S-3-2: 15 天前開始的試用，GetLicenseInfo 狀態為 "expired"
func TestLicenseService_GetLicenseInfo_WhenTrialExpired_StatusIsExpired(t *testing.T) {
	// Arrange
	svc, repo, _, _ := createLicenseService()
	givenLicenseExists(repo, createExpiredLicense())

	// Act
	info, err := svc.GetLicenseInfo(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "expired", info.Status)
}

// ─────────────────────────────────────────────────────────────────────────────
// S-1: 有效 key 首次啟用成功
// ─────────────────────────────────────────────────────────────────────────────

// S-1-1: verifier 成功，ActivateLicense 回傳 nil error
func TestLicenseService_ActivateLicense_WhenVerifierSucceeds_ReturnsNoError(t *testing.T) {
	// Arrange
	svc, repo, machineId, verifier := createLicenseService()
	givenLicenseExists(repo, createActiveLicense(0))
	givenSaveSucceeds(repo)
	givenMachineID(machineId)
	givenVerifierSucceeds(verifier)

	// Act
	err := svc.ActivateLicense(context.Background(), "XXXXXXXX-XXXXXXXX-XXXXXXXX-XXXXXXXX")

	// Assert
	assert.NoError(t, err)
}

// S-1-2: verifier 成功，Save 被呼叫
func TestLicenseService_ActivateLicense_WhenVerifierSucceeds_SaveIsCalled(t *testing.T) {
	// Arrange
	svc, repo, machineId, verifier := createLicenseService()
	givenLicenseExists(repo, createActiveLicense(0))
	givenSaveSucceeds(repo)
	givenMachineID(machineId)
	givenVerifierSucceeds(verifier)

	// Act
	_ = svc.ActivateLicense(context.Background(), "XXXXXXXX-XXXXXXXX-XXXXXXXX-XXXXXXXX")

	// Assert
	repo.AssertCalled(t, "Save", mock.Anything, mock.AnythingOfType("*license.License"))
}

// ─────────────────────────────────────────────────────────────────────────────
// S-2: 無效 key 被拒絕
// ─────────────────────────────────────────────────────────────────────────────

// S-2-1: verifier 回傳 ErrInvalidLicenseKey，ActivateLicense 傳遞該錯誤
func TestLicenseService_ActivateLicense_WhenVerifierRejectsKey_ReturnsErrInvalidLicenseKey(t *testing.T) {
	// Arrange
	svc, repo, _, verifier := createLicenseService()
	givenLicenseExists(repo, createActiveLicense(0))
	givenVerifierReturns(verifier, license.ErrInvalidLicenseKey)

	// Act
	err := svc.ActivateLicense(context.Background(), "bad-key")

	// Assert
	assert.ErrorIs(t, err, license.ErrInvalidLicenseKey)
}

// S-2-2: verifier 回傳錯誤，Save 不被呼叫
func TestLicenseService_ActivateLicense_WhenVerifierFails_SaveIsNotCalled(t *testing.T) {
	// Arrange
	svc, repo, _, verifier := createLicenseService()
	givenLicenseExists(repo, createActiveLicense(0))
	givenVerifierReturns(verifier, license.ErrInvalidLicenseKey)

	// Act
	_ = svc.ActivateLicense(context.Background(), "bad-key")

	// Assert
	repo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
}

// ─────────────────────────────────────────────────────────────────────────────
// S-3: 已在其他機器使用的 key 被拒絕
// ─────────────────────────────────────────────────────────────────────────────

// S-3-1: verifier 回傳 ErrLicenseAlreadyUsed
func TestLicenseService_ActivateLicense_WhenKeyAlreadyUsed_ReturnsErrLicenseAlreadyUsed(t *testing.T) {
	// Arrange
	svc, repo, _, verifier := createLicenseService()
	givenLicenseExists(repo, createActiveLicense(0))
	givenVerifierReturns(verifier, license.ErrLicenseAlreadyUsed)

	// Act
	err := svc.ActivateLicense(context.Background(), "used-key")

	// Assert
	assert.ErrorIs(t, err, license.ErrLicenseAlreadyUsed)
}

// ─────────────────────────────────────────────────────────────────────────────
// S-4: 網路失敗時回傳錯誤
// ─────────────────────────────────────────────────────────────────────────────

// S-4-1: verifier 回傳網路錯誤，ActivateLicense 傳遞該錯誤
func TestLicenseService_ActivateLicense_WhenNetworkFails_ReturnsError(t *testing.T) {
	// Arrange
	svc, repo, _, verifier := createLicenseService()
	givenLicenseExists(repo, createActiveLicense(0))
	networkErr := errors.New("無法連線至驗證伺服器，請確認網路連線")
	givenVerifierReturns(verifier, networkErr)

	// Act
	err := svc.ActivateLicense(context.Background(), "any-key")

	// Assert
	assert.Error(t, err)
}

// ─────────────────────────────────────────────────────────────────────────────
// S-5: 已啟用的 license 不可重複啟用
// ─────────────────────────────────────────────────────────────────────────────

// S-5-1: 已啟用狀態，ActivateLicense 回傳 ErrAlreadyActivated
func TestLicenseService_ActivateLicense_WhenAlreadyActivated_ReturnsErrAlreadyActivated(t *testing.T) {
	// Arrange
	svc, repo, _, _ := createLicenseService()
	givenLicenseExists(repo, createActivatedLicense(0))

	// Act
	err := svc.ActivateLicense(context.Background(), "any-key")

	// Assert
	assert.ErrorIs(t, err, license.ErrAlreadyActivated)
}

// S-5-2: 已啟用狀態，verifier 不被呼叫
func TestLicenseService_ActivateLicense_WhenAlreadyActivated_VerifierIsNotCalled(t *testing.T) {
	// Arrange
	svc, repo, _, verifier := createLicenseService()
	givenLicenseExists(repo, createActivatedLicense(0))

	// Act
	_ = svc.ActivateLicense(context.Background(), "any-key")

	// Assert
	verifier.AssertNotCalled(t, "Verify", mock.Anything, mock.Anything)
}

// ─────────────────────────────────────────────────────────────────────────────
// 既有測試：啟用後 CanRun / GetLicenseInfo
// ─────────────────────────────────────────────────────────────────────────────

func TestLicenseService_CanRun_WhenActivated_ReturnsTrue(t *testing.T) {
	// Arrange
	svc, repo, _, _ := createLicenseService()
	givenLicenseExists(repo, createActivatedLicense(0))

	// Act
	canRun, err := svc.CanRunBackgroundProcessor(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.True(t, canRun)
}

func TestLicenseService_CanRun_WhenActivatedAndTrialExpired_ReturnsTrue(t *testing.T) {
	// Arrange
	svc, repo, _, _ := createLicenseService()
	givenLicenseExists(repo, createActivatedLicense(20))

	// Act
	canRun, err := svc.CanRunBackgroundProcessor(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.True(t, canRun)
}

func TestLicenseService_GetLicenseInfo_WhenActivated_StatusIsActivated(t *testing.T) {
	// Arrange
	svc, repo, _, _ := createLicenseService()
	givenLicenseExists(repo, createActivatedLicense(20))

	// Act
	info, err := svc.GetLicenseInfo(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "activated", info.Status)
}
