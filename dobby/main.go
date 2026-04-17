package main

import (
	"embed"
	"log"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"github.com/dobby/filemanager/internal/application"
	"github.com/dobby/filemanager/internal/infrastructure/filesystem"
	infralicense "github.com/dobby/filemanager/internal/infrastructure/license"
	"github.com/dobby/filemanager/internal/infrastructure/persistence"
)

// assets embeds the compiled frontend.
// Build the React app into frontend/dist/ before running `wails build`.
//
//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// ── Database ──────────────────────────────────────────────────────────────
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("cannot determine home directory:", err)
	}
	dataDir := filepath.Join(home, ".dobby")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		log.Fatal("cannot create data directory:", err)
	}
	dsn := filepath.Join(dataDir, "dobby.db")

	db, err := persistence.Open(dsn)
	if err != nil {
		log.Fatal("database init:", err)
	}
	defer db.Close()

	// ── Repositories ─────────────────────────────────────────────────────────
	ruleRepo := persistence.NewSQLiteRuleRepository(db)
	jobRepo := persistence.NewSQLiteProcessingJobRepository(db)
	logRepo := persistence.NewSQLiteOperationLogRepository(db)

	// ── Application services ─────────────────────────────────────────────────
	ruleSvc := application.NewRuleService(ruleRepo)
	logSvc := application.NewLogService(logRepo)
	processorSvc := application.NewBackgroundProcessorService(
		ruleRepo,
		jobRepo,
		logRepo,
		filesystem.NewOSFileSystem(),
		logRepo,
	)

	// ── License service ───────────────────────────────────────────────────────
	licenseRepo := infralicense.NewLocalLicenseRepository(dataDir)
	machineId := infralicense.NewMachineIdProvider()
	gumroadVerifier := infralicense.NewProductionGumroadVerifier()
	licenseSvc := application.NewLicenseService(licenseRepo, machineId, gumroadVerifier)
	processorSvc.SetLicenseGuard(licenseSvc)

	// ── Wails app ─────────────────────────────────────────────────────────────
	app := NewApp(ruleSvc, logSvc, processorSvc, licenseSvc)

	if err := wails.Run(&options.App{
		Title:     "Dobby — 檔案管家",
		Width:     1280,
		Height:    800,
		MinWidth:  900,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.startup,
		Bind: []any{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
	}); err != nil {
		log.Fatal("wails.Run:", err)
	}
}
