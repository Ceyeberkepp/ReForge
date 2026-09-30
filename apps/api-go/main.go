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
	router.Get("/boot/catalog.ipxe", app.pxeCatalog)
	router.Get("/boot/sources.ipxe", app.pxeSources)
	router.Get("/boot/action.ipxe", app.pxeAction)
	router.Get("/boot/assets/{kind}", app.getBrandingAsset)
	router.Post("/api/hosts/register", app.registerHost)

	router.Group(func(protected chi.Router) {
		protected.Use(app.auth)

		protected.Get("/api/auth/me", app.me)
		protected.Post("/api/auth/logout", app.logout)
		protected.Post("/api/auth/password", app.changePassword)

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
		protected.Post("/api/pxe/assets/{kind}", app.uploadBrandingAsset)

		protected.Get("/api/admin/users", app.listAdminUsers)
		protected.Post("/api/admin/users", app.saveAdminUser)
		protected.Put("/api/admin/users/{id}", app.saveAdminUser)
		protected.Delete("/api/admin/users/{id}", app.deleteAdminUser)
		protected.Get("/api/admin/groups", app.listGroups)
		protected.Post("/api/admin/groups", app.saveGroup)
		protected.Put("/api/admin/groups/{id}", app.saveGroup)
		protected.Delete("/api/admin/groups/{id}", app.deleteGroup)

		protected.Get("/api/isos", app.listISOs)
		protected.Post("/api/isos", app.saveISO)
		protected.Put("/api/isos/{id}", app.saveISO)
		protected.Delete("/api/isos/{id}", app.deleteISO)
		protected.Get("/api/clones", app.listClones)
		protected.Post("/api/clones", app.saveClone)
		protected.Put("/api/clones/{id}", app.saveClone)
		protected.Delete("/api/clones/{id}", app.deleteClone)

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
