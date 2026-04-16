package application_test

import (
	"context"
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

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func createLicenseService() (*application.LicenseService, *mockLicenseRepo, *mockMachineId) {
	repo := new(mockLicenseRepo)
	machineId := new(mockMachineId)
	svc := application.NewLicenseService(repo, machineId)
	return svc, repo, machineId
}

func createActiveLicense(daysAgo int) *license.License {
	return license.Reconstitute(time.Now().AddDate(0, 0, -daysAgo), "", "")
}

func createExpiredLicense() *license.License {
	return license.Reconstitute(time.Now().AddDate(0, 0, -15), "", "")
}

func createActivatedLicense(daysAgo int) *license.License {
	validator := license.NewLicenseKeyValidator()
	key := validator.GenerateKey("ABCD", "EFGH")
	return license.Reconstitute(time.Now().AddDate(0, 0, -daysAgo), key, "test-machine-id")
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
	svc, repo, _ := createLicenseService()
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
	svc, repo, _ := createLicenseService()
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
	svc, repo, _ := createLicenseService()
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
	svc, repo, _ := createLicenseService()
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
	svc, repo, _ := createLicenseService()
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
	svc, repo, _ := createLicenseService()
	givenLicenseExists(repo, createExpiredLicense())

	// Act
	info, err := svc.GetLicenseInfo(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "expired", info.Status)
}

// ─────────────────────────────────────────────────────────────────────────────
// S-5: Valid license key activates the app
// ─────────────────────────────────────────────────────────────────────────────

// S-5-1: 有效 key，ActivateLicense 回傳 nil error
func TestLicenseService_ActivateLicense_WithValidKey_ReturnsNoError(t *testing.T) {
	// Arrange
	svc, repo, machineId := createLicenseService()
	givenLicenseExists(repo, createActiveLicense(0))
	givenSaveSucceeds(repo)
	givenMachineID(machineId)
	validKey := license.NewLicenseKeyValidator().GenerateKey("ABCD", "EFGH")

	// Act
	err := svc.ActivateLicense(context.Background(), validKey)

	// Assert
	assert.NoError(t, err)
}

// S-5-2: 有效 key 啟用後，CanRunBackgroundProcessor 回傳 true
func TestLicenseService_CanRun_WhenActivated_ReturnsTrue(t *testing.T) {
	// Arrange
	svc, repo, _ := createLicenseService()
	givenLicenseExists(repo, createActivatedLicense(0))

	// Act
	canRun, err := svc.CanRunBackgroundProcessor(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.True(t, canRun)
}

// ─────────────────────────────────────────────────────────────────────────────
// S-6: Invalid license key checksum is rejected
// ─────────────────────────────────────────────────────────────────────────────

// S-6-1: checksum 錯誤的 key，ActivateLicense 回傳 ErrInvalidLicenseKey
func TestLicenseService_ActivateLicense_WithBadChecksum_ReturnsInvalidKeyError(t *testing.T) {
	// Arrange
	svc, _, _ := createLicenseService()

	// Act
	err := svc.ActivateLicense(context.Background(), "DOBBY-ABCD-EFGH-ZZZZ")

	// Assert
	assert.ErrorIs(t, err, license.ErrInvalidLicenseKey)
}

// ─────────────────────────────────────────────────────────────────────────────
// S-7: Malformed license key format is rejected
// ─────────────────────────────────────────────────────────────────────────────

// S-7-1: 格式錯誤的 key，ActivateLicense 回傳 ErrInvalidLicenseKeyFormat
func TestLicenseService_ActivateLicense_WithBadFormat_ReturnsFormatError(t *testing.T) {
	// Arrange
	svc, _, _ := createLicenseService()

	// Act
	err := svc.ActivateLicense(context.Background(), "INVALID-FORMAT")

	// Assert
	assert.ErrorIs(t, err, license.ErrInvalidLicenseKeyFormat)
}

// ─────────────────────────────────────────────────────────────────────────────
// S-8: Activated license allows processor even after trial expires
// ─────────────────────────────────────────────────────────────────────────────

// S-8-1: 20 天前啟用的授權，CanRunBackgroundProcessor 仍回傳 true
func TestLicenseService_CanRun_WhenActivatedAndTrialExpired_ReturnsTrue(t *testing.T) {
	// Arrange
	svc, repo, _ := createLicenseService()
	givenLicenseExists(repo, createActivatedLicense(20))

	// Act
	canRun, err := svc.CanRunBackgroundProcessor(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.True(t, canRun)
}

// S-8-2: 20 天前啟用的授權，GetLicenseInfo 狀態為 "activated"
func TestLicenseService_GetLicenseInfo_WhenActivated_StatusIsActivated(t *testing.T) {
	// Arrange
	svc, repo, _ := createLicenseService()
	givenLicenseExists(repo, createActivatedLicense(20))

	// Act
	info, err := svc.GetLicenseInfo(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "activated", info.Status)
}
