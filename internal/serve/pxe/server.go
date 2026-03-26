package pxe

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/kennetholsenatm-gif/omnigraph/internal/baremetal"
)

// HTTPBootServer implements an HTTP server for network boot files
// It serves iPXE binaries, scripts, Cloud-Init configs, and OS images
type HTTPBootServer struct {
	mu       sync.RWMutex
	config   HTTPBootConfig
	server   *http.Server
	router   *http.ServeMux
	eventBus baremetal.EventBus
}

// HTTPBootConfig holds HTTP boot server configuration
type HTTPBootConfig struct {
	// Server configuration
	ListenAddress string
	Port          int
	TLSCertFile   string
	TLSKeyFile    string

	// Asset paths
	IPXEBinaryPath string // Path to iPXE binaries
	CloudInitPath  string // Path to Cloud-Init templates
	OSImagesPath   string // Path to OS images
	ScriptsPath    string // Path to iPXE scripts

	// Boot configuration
	DefaultBootFile  string // Default boot file (e.g., "ipxe.efi")
	HTTPBootURL      string // HTTP boot URL base
	CloudInitEnabled bool   // Enable Cloud-Init generation
	IgnitionEnabled  bool   // Enable CoreOS Ignition generation

	// Inventory integration
	TargetLookup func(mac string) (*baremetal.Target, error)

	// Security
	RequireMACAuth bool     // Require MAC address authentication
	AllowedMACs    []string // List of allowed MAC addresses
	TLSRequired    bool     // Require TLS for all requests

	// Logging
	Verbose bool
	LogFile string
}

// iPXEScript represents a generated iPXE script
type iPXEScript struct {
	Kernel       string
	InitRD       string
	KernelArgs   string
	CloudInitURL string
}

// CloudInitConfig represents Cloud-Init configuration
type CloudInitConfig struct {
	UserData   string
	MetaData   string
	VendorData string
}

// NewHTTPBootServer creates a new HTTP boot server
func NewHTTPBootServer(config HTTPBootConfig) (*HTTPBootServer, error) {
	s := &HTTPBootServer{
		config: config,
		router: http.NewServeMux(),
	}

	// Register routes
	s.registerRoutes()

	return s, nil
}

// registerRoutes registers HTTP routes
func (s *HTTPBootServer) registerRoutes() {
	// iPXE binary serving
	s.router.HandleFunc("/ipxe/", s.handleIPXEBinary)

	// iPXE script generation
	s.router.HandleFunc("/boot/", s.handleBootScript)

	// Cloud-Init serving
	s.router.HandleFunc("/cloud-init/", s.handleCloudInit)

	// Ignition serving (CoreOS)
	s.router.HandleFunc("/ignition/", s.handleIgnition)

	// OS images
	s.router.HandleFunc("/images/", s.handleOSImages)

	// Health check
	s.router.HandleFunc("/health", s.handleHealth)

	// Omnigraph agent
	s.router.HandleFunc("/omnigraph/", s.handleOmnigraphAgent)
}

// Start starts the HTTP boot server
func (s *HTTPBootServer) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.config.ListenAddress, s.config.Port)

	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("HTTP Boot Server listening on %s", addr)

	go func() {
		var err error
		if s.config.TLSCertFile != "" && s.config.TLSKeyFile != "" {
			err = s.server.ListenAndServeTLS(s.config.TLSCertFile, s.config.TLSKeyFile)
		} else {
			err = s.server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP Boot Server error: %v", err)
		}
	}()

	return nil
}

// Stop stops the HTTP boot server
func (s *HTTPBootServer) Stop(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

// handleIPXEBinary serves iPXE binaries
func (s *HTTPBootServer) handleIPXEBinary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract filename from path
	filename := filepath.Base(r.URL.Path)

	// Check if file exists
	filePath := filepath.Join(s.config.IPXEBinaryPath, filename)

	if s.config.Verbose {
		log.Printf("Serving iPXE binary: %s", filePath)
	}

	// Serve file
	http.ServeFile(w, r, filePath)

	// Publish event
	if s.eventBus != nil {
		s.eventBus.Publish(baremetal.Event{
			Type:      "httpboot.binary",
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"file": filename,
				"mac":  r.Header.Get("X-MAC-Address"),
			},
		})
	}
}

// handleBootScript generates and serves iPXE scripts
func (s *HTTPBootServer) handleBootScript(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract MAC address from query parameter
	mac := r.URL.Query().Get("mac")
	if mac == "" {
		// Try to get from header
		mac = r.Header.Get("X-MAC-Address")
	}

	if mac == "" {
		http.Error(w, "MAC address required", http.StatusBadRequest)
		return
	}

	// Look up target
	target, err := s.config.TargetLookup(mac)
	if err != nil {
		if s.config.Verbose {
			log.Printf("Target not found for MAC %s: %v", mac, err)
		}
		http.Error(w, "Target not found", http.StatusNotFound)
		return
	}

	// Generate iPXE script
	script := s.generateIPXEScript(target, mac)

	if s.config.Verbose {
		log.Printf("Generated iPXE script for MAC %s (target %s)", mac, target.ID)
	}

	// Set content type
	w.Header().Set("Content-Type", "text/plain")

	// Write script
	_, err = w.Write([]byte(script))
	if err != nil {
		log.Printf("Error writing iPXE script: %v", err)
		return
	}

	// Publish event
	if s.eventBus != nil {
		s.eventBus.Publish(baremetal.Event{
			Type:      "httpboot.script",
			TargetID:  target.ID,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"mac": mac,
			},
		})
	}
}

// generateIPXEScript generates an iPXE script for a target
func (s *HTTPBootServer) generateIPXEScript(target *baremetal.Target, mac string) string {
	var script string

	// Set timeout
	script += "#!ipxe\n"
	script += "echo OmniGraph iPXE Boot Script\n"
	script += "echo Target: " + target.ID + "\n"
	script += "echo MAC: " + mac + "\n\n"

	// Determine boot mode
	switch target.BootMode {
	case "http":
		// UEFI HTTP Boot
		script += s.generateHTTPBootScript(target, mac)
	case "ipxe":
		// iPXE chain loading
		script += s.generateIPXEChainScript(target, mac)
	case "pxe":
		// Traditional PXE
		script += s.generatePXEScript(target, mac)
	default:
		// Default to HTTP boot
		script += s.generateHTTPBootScript(target, mac)
	}

	return script
}

// generateHTTPBootScript generates HTTP boot script
func (s *HTTPBootServer) generateHTTPBootScript(target *baremetal.Target, mac string) string {
	var script string

	script += "# HTTP Boot\n"
	script += "echo Booting via HTTP...\n\n"

	// Set kernel and initrd
	kernelURL := fmt.Sprintf("%s/images/%s/vmlinuz", s.config.HTTPBootURL, target.OSProfile)
	initrdURL := fmt.Sprintf("%s/images/%s/initrd", s.config.HTTPBootURL, target.OSProfile)

	script += fmt.Sprintf("kernel %s\n", kernelURL)
	script += fmt.Sprintf("initrd %s\n", initrdURL)

	// Build kernel arguments
	kernelArgs := s.buildKernelArgs(target, mac)
	script += fmt.Sprintf("imgargs vmlinuz %s\n", kernelArgs)

	script += "boot\n"

	return script
}

// generateIPXEChainScript generates iPXE chain loading script
func (s *HTTPBootServer) generateIPXEChainScript(target *baremetal.Target, mac string) string {
	var script string

	script += "# iPXE Chain Loading\n"
	script += "echo Chain loading iPXE...\n\n"

	// Chain to next iPXE script
	nextScriptURL := fmt.Sprintf("%s/boot/?mac=%s", s.config.HTTPBootURL, mac)

	script += fmt.Sprintf("chain %s\n", nextScriptURL)

	return script
}

// generatePXEScript generates traditional PXE script
func (s *HTTPBootServer) generatePXEScript(target *baremetal.Target, mac string) string {
	var script string

	script += "# Traditional PXE\n"
	script += "echo Booting via PXE...\n\n"

	// For traditional PXE, we'll use iPXE to chain to HTTP boot
	script += "# Chain to HTTP boot for modern systems\n"
	httpBootURL := fmt.Sprintf("%s/boot/?mac=%s&mode=http", s.config.HTTPBootURL, mac)
	script += fmt.Sprintf("chain %s\n", httpBootURL)

	return script
}

// buildKernelArgs builds kernel arguments for OS installation
func (s *HTTPBootServer) buildKernelArgs(target *baremetal.Target, mac string) string {
	var args string

	// Network configuration
	if target.Network.IPAddress != "" {
		args += fmt.Sprintf("ip=%s::%s:%s:%s:eno1:off ",
			target.Network.IPAddress,
			target.Network.Gateway,
			target.Network.Gateway, // Assuming gateway is also DNS
			target.ID)
	}

	// Cloud-Init configuration
	if s.config.CloudInitEnabled {
		cloudInitURL := fmt.Sprintf("%s/cloud-init/user-data?mac=%s", s.config.HTTPBootURL, mac)
		args += fmt.Sprintf("ds=nocloud-net;s=%s ", cloudInitURL)
	}

	// Ignition configuration (CoreOS)
	if s.config.IgnitionEnabled {
		ignitionURL := fmt.Sprintf("%s/ignition/config?mac=%s", s.config.HTTPBootURL, mac)
		args += fmt.Sprintf("ignition.config.url=%s ", ignitionURL)
	}

	// Auto-install parameters
	args += "auto-install "
	args += "console=tty0 "
	args += "console=ttyS0,115200n8 "

	return args
}

// handleCloudInit serves Cloud-Init configurations
func (s *HTTPBootServer) handleCloudInit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract MAC address
	mac := r.URL.Query().Get("mac")
	if mac == "" {
		mac = r.Header.Get("X-MAC-Address")
	}

	if mac == "" {
		http.Error(w, "MAC address required", http.StatusBadRequest)
		return
	}

	// Look up target
	target, err := s.config.TargetLookup(mac)
	if err != nil {
		http.Error(w, "Target not found", http.StatusNotFound)
		return
	}

	// Determine which Cloud-Init file to serve
	path := r.URL.Path

	var content string
	var contentType string

	switch {
	case path == "/cloud-init/user-data" || path == "/cloud-init/user-data/":
		content = s.generateCloudInitUserData(target, mac)
		contentType = "text/plain"
	case path == "/cloud-init/meta-data" || path == "/cloud-init/meta-data/":
		content = s.generateCloudInitMetaData(target, mac)
		contentType = "text/plain"
	case path == "/cloud-init/vendor-data" || path == "/cloud-init/vendor-data/":
		content = s.generateCloudInitVendorData(target, mac)
		contentType = "text/plain"
	default:
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if s.config.Verbose {
		log.Printf("Serving Cloud-Init %s for MAC %s (target %s)", path, mac, target.ID)
	}

	w.Header().Set("Content-Type", contentType)
	w.Write([]byte(content))
}

// generateCloudInitUserData generates Cloud-Init user-data
func (s *HTTPBootServer) generateCloudInitUserData(target *baremetal.Target, mac string) string {
	var userData string

	userData += "#cloud-config\n"
	userData += "# OmniGraph Cloud-Init Configuration\n"
	userData += fmt.Sprintf("# Target: %s\n", target.ID)
	userData += fmt.Sprintf("# MAC: %s\n\n", mac)

	// Hostname
	userData += fmt.Sprintf("hostname: %s\n", target.ID)

	// Users
	userData += "users:\n"
	userData += "  - default\n"
	userData += "  - name: ansible\n"
	userData += "    sudo: ALL=(ALL) NOPASSWD:ALL\n"
	userData += "    shell: /bin/bash\n"
	userData += "    ssh_authorized_keys:\n"
	userData += "      - ssh-rsa AAAA... # Replace with actual key\n"

	// Packages
	userData += "\npackages:\n"
	userData += "  - openssh-server\n"
	userData += "  - curl\n"
	userData += "  - wget\n"
	userData += "  - python3\n"

	// Run commands
	userData += "\nruncmd:\n"
	userData += "  - systemctl enable ssh\n"
	userData += "  - systemctl start ssh\n"
	userData += "  - curl -fsSL https://omnigraph.example.com/agent/install.sh | bash\n"

	// Write files
	userData += "\nwrite_files:\n"
	userData += "  - path: /etc/omnigraph/target-id\n"
	userData += "    content: |\n"
	userData += fmt.Sprintf("      %s\n", target.ID)

	return userData
}

// generateCloudInitMetaData generates Cloud-Init meta-data
func (s *HTTPBootServer) generateCloudInitMetaData(target *baremetal.Target, mac string) string {
	var metaData string

	metaData += fmt.Sprintf("instance-id: %s\n", target.ID)
	metaData += fmt.Sprintf("local-hostname: %s\n", target.ID)

	return metaData
}

// generateCloudInitVendorData generates Cloud-Init vendor-data
func (s *HTTPBootServer) generateCloudInitVendorData(target *baremetal.Target, mac string) string {
	var vendorData string

	vendorData += "#cloud-config\n"
	vendorData += "# OmniGraph Vendor Data\n"

	return vendorData
}

// handleIgnition serves CoreOS Ignition configurations
func (s *HTTPBootServer) handleIgnition(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract MAC address
	mac := r.URL.Query().Get("mac")
	if mac == "" {
		mac = r.Header.Get("X-MAC-Address")
	}

	if mac == "" {
		http.Error(w, "MAC address required", http.StatusBadRequest)
		return
	}

	// Look up target
	target, err := s.config.TargetLookup(mac)
	if err != nil {
		http.Error(w, "Target not found", http.StatusNotFound)
		return
	}

	// Generate Ignition config
	config := s.generateIgnitionConfig(target, mac)

	if s.config.Verbose {
		log.Printf("Serving Ignition config for MAC %s (target %s)", mac, target.ID)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(config))
}

// generateIgnitionConfig generates CoreOS Ignition configuration
func (s *HTTPBootServer) generateIgnitionConfig(target *baremetal.Target, mac string) string {
	var config string

	config += "{\n"
	config += "  \"ignition\": { \"version\": \"3.3.0\" },\n"
	config += fmt.Sprintf("  \"systemd\": { \"units\": [{ \"name\": \"omnigraph-agent.service\", \"enabled\": true, \"contents\": \"[Unit]\\nDescription=OmniGraph Agent\\nAfter=network-online.target\\nWants=network-online.target\\n\\n[Service]\\nType=simple\\nExecStart=/usr/local/bin/omnigraph-agent\\nRestart=always\\n\\n[Install]\\nWantedBy=multi-user.target\" }] },\n")
	config += fmt.Sprintf("  \"passwd\": { \"users\": [{ \"name\": \"ansible\", \"sshAuthorizedKeys\": [\"ssh-rsa AAAA...\"] }] }\n")
	config += "}\n"

	return config
}

// handleOSImages serves OS images
func (s *HTTPBootServer) handleOSImages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract path
	path := r.URL.Path

	if s.config.Verbose {
		log.Printf("Serving OS image: %s", path)
	}

	// Serve file
	filePath := filepath.Join(s.config.OSImagesPath, path)
	http.ServeFile(w, r, filePath)
}

// handleHealth handles health check requests
func (s *HTTPBootServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"healthy","service":"omnigraph-http-boot"}`))
}

// handleOmnigraphAgent serves the OmniGraph agent installer
func (s *HTTPBootServer) handleOmnigraphAgent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract path
	path := r.URL.Path

	// Serve agent installer or binary
	if path == "/omnigraph/agent/install.sh" {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("#!/bin/bash\n# OmniGraph Agent Installer\necho 'Installing OmniGraph agent...'\n# Add actual installation logic here\n"))
	} else {
		// Serve agent binary
		filePath := filepath.Join(s.config.ScriptsPath, path)
		http.ServeFile(w, r, filePath)
	}
}

// SetEventBus sets the event bus for publishing events
func (s *HTTPBootServer) SetEventBus(bus baremetal.EventBus) {
	s.eventBus = bus
}
