package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type App struct {
	db  *gorm.DB
	cfg Config
	log *slog.Logger
}

func main() {
	cfg := loadConfig()
	db, err := openDB(cfg)
	if err != nil {
		panic(err)
	}
	if err := migrate(db); err != nil {
		panic(err)
	}

	app := &App{
		db:  db,
		cfg: cfg,
		log: slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}
	if err := app.bootstrapAdmin(); err != nil {
		panic(err)
	}

	router := chi.NewRouter()
	router.Use(app.securityHeaders)
	router.Use(app.cors)

	router.Get("/health", app.health)
	router.Post("/api/auth/login", app.login)

	router.Get("/boot/ipxe", app.ipxe)
	router.Get("/boot/deploy.ipxe", app.deployIPXE)
	router.Get("/boot/register.ipxe", app.registerIPXE)
	router.Post("/api/hosts/register", app.registerHost)

	router.Group(func(protected chi.Router) {
		protected.Use(app.auth)

		protected.Get("/api/auth/me", app.me)
		protected.Post("/api/auth/logout", app.logout)\n\t\tprotected.Post("/api/auth/password", app.changePassword)

		protected.Get("/api/dashboard", app.dashboard)

		protected.Get("/api/images", app.listImages)
		protected.Post("/api/images", app.saveImage)
		protected.Put("/api/images/{id}", app.saveImage)
		protected.Delete("/api/images/{id}", app.deleteImage)

		protected.Get("/api/departments", app.listDepartments)
		protected.Post("/api/departments", app.saveDepartment)
		protected.Put("/api/departments/{id}", app.saveDepartment)
		protected.Delete("/api/departments/{id}", app.deleteDepartment)

		protected.Get("/api/software", app.listSoftware)
		protected.Post("/api/software", app.saveSoftware)
		protected.Put("/api/software/{id}", app.saveSoftware)
		protected.Delete("/api/software/{id}", app.deleteSoftware)
		protected.Post("/api/software/bootstrap", app.bootstrapSoftware)

		protected.Get("/api/hosts", app.listHosts)
		protected.Post("/api/hosts", app.saveHost)
		protected.Put("/api/hosts/{id}", app.saveHost)
		protected.Delete("/api/hosts/{id}", app.deleteHost)

		protected.Get("/api/directory", app.getDirectory)
		protected.Put("/api/directory", app.saveDirectory)
		protected.Post("/api/directory/test", app.testDirectory)

		protected.Get("/api/deployments", app.listDeployments)
		protected.Post("/api/deployments", app.createDeployment)
		protected.Post("/api/deployments/{id}/{action}", app.deploymentAction)
		protected.Get("/api/deployments/{id}/plan", app.deploymentPlan)

		protected.Get("/api/pxe", app.getPXE)
		protected.Put("/api/pxe", app.savePXE)

		protected.Get("/api/audit", app.listAudit)
	})

	router.Group(func(worker chi.Router) {
		worker.Use(app.workerAuth)
		worker.Get("/api/worker/jobs/next", app.workerNextJob)
		worker.Get("/api/worker/deployments/{id}/plan", app.deploymentPlan)
		worker.Patch("/api/worker/deployments/{id}", app.workerUpdate)
	})

	server := &http.Server{
		Addr:              cfg.Listen,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	app.log.Info("ReForge API starting", "listen", cfg.Listen, "database", cfg.DBDriver)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}
