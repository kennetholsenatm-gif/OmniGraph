package baremetal

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

// ProxyDHCP implements a ProxyDHCP service that injects boot parameters
// without managing IP leases. It listens for DHCP Discover broadcasts
// and responds with boot options (next-server, bootfile).
type ProxyDHCP struct {
	mu           sync.RWMutex
	config       ProxyDHCPConfig
	server       *net.UDPConn
	leaseTracker map[string]*ProxyLease
	eventBus     EventBus
	stopCh       chan struct{}
}

// ProxyDHCPConfig holds ProxyDHCP configuration
type ProxyDHCPConfig struct {
	// Network configuration
	ListenAddress string // e.g., "0.0.0.0:67"
	Interface     string // Network interface to listen on
	
	// Boot configuration
	NextServer    string // TFTP/HTTP server IP
	BootFile      string // Boot file path (e.g., "/ipxe.efi")
	HTTPBootURL   string // HTTP boot URL (for UEFI HTTP boot)
	
	// Inventory integration
	InventoryPath string // Path to MAC address inventory
	TargetLookup  func(mac string) (*Target, error)
	
	// Logging
	Verbose       bool
	LogFile       string
}

// ProxyLease represents a proxy lease (for tracking only, not actual IP assignment)
type ProxyLease struct {
	MACAddress  string
	TargetID    string
	OfferedAt   time.Time
	ExpiresAt   time.Time
	BootFile    string
	NextServer  string
}

// DHCP packet constants
const (
	DHCPDiscover = 1
	DHCPOffer    = 2
	DHCPRequest  = 3
	DHCPAck      = 5
	DHCPNak      = 6
	DHCPRelease  = 7
	
	DHCPServerPort   = 67
	DHCPClientPort   = 68
	MaxDHCPacketSize = 576
)

// DHCP options
const (
	OptionSubnetMask       = 1
	OptionRouter           = 3
	OptionDNSServer        = 6
	OptionHostName         = 12
	OptionBootFileSize     = 13
	OptionDomainName       = 15
	OptionBroadcastAddr    = 28
	OptionVendorSpecific   = 43
	OptionRequestedIP      = 50
	OptionIPAddressLease   = 51
	OptionMessageType      = 53
	OptionServerIdentifier = 54
	OptionParameterRequest = 55
	OptionTFTPServerName   = 66
	OptionBootfileName     = 67
	OptionClientGUID       = 97
	OptionHTTPBootURL      = 240 // Custom option for HTTP boot URL
)

// DHCP packet structure
type DHCPPacket struct {
	OpCode     byte
	HardwareType byte
	HardwareLen  byte
	Hops       byte
	TransactionID uint32
	Seconds    uint16
	Flags      uint16
	ClientIP   net.IP
	YourIP     net.IP
	ServerIP   net.IP
	GatewayIP  net.IP
	ClientMAC  net.HardwareAddr
	ServerName [64]byte
	BootFile   [128]byte
	Options    []DHCPOption
}

// DHCPOption represents a DHCP option
type DHCPOption struct {
	Code   byte
	Length byte
	Data   []byte
}

// NewProxyDHCP creates a new ProxyDHCP service
func NewProxyDHCP(config ProxyDHCPConfig) (*ProxyDHCP, error) {
	p := &ProxyDHCP{
		config:       config,
		leaseTracker: make(map[string]*ProxyLease),
		stopCh:       make(chan struct{}),
	}
	
	return p, nil
}

// Start starts the ProxyDHCP service
func (p *ProxyDHCP) Start(ctx context.Context) error {
	// Parse listen address
	addr, err := net.ResolveUDPAddr("udp", p.config.ListenAddress)
	if err != nil {
		return fmt.Errorf("failed to resolve listen address: %w", err)
	}
	
	// Create UDP listener
	p.server, err = net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", p.config.ListenAddress, err)
	}
	
	log.Printf("ProxyDHCP listening on %s", p.config.ListenAddress)
	
	// Start listening for DHCP packets
	go p.listenForPackets(ctx)
	
	// Start lease cleanup routine
	go p.cleanupExpiredLeases(ctx)
	
	return nil
}

// Stop stops the ProxyDHCP service
func (p *ProxyDHCP) Stop() {
	close(p.stopCh)
	if p.server != nil {
		p.server.Close()
	}
}

// listenForPackets listens for incoming DHCP packets
func (p *ProxyDHCP) listenForPackets(ctx context.Context) {
	buffer := make([]byte, MaxDHCPacketSize)
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-p.stopCh:
			return
		default:
			// Set read deadline
			p.server.SetReadDeadline(time.Now().Add(1 * time.Second))
			
			n, addr, err := p.server.ReadFromUDP(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue // Timeout, continue listening
				}
				log.Printf("Error reading UDP packet: %v", err)
				continue
			}
			
			// Parse DHCP packet
			packet, err := p.parseDHCPPacket(buffer[:n])
			if err != nil {
				log.Printf("Error parsing DHCP packet: %v", err)
				continue
			}
			
			// Handle DHCP packet
			go p.handleDHCPPacket(ctx, packet, addr)
		}
	}
}

// parseDHCPPacket parses a raw DHCP packet
func (p *ProxyDHCP) parseDHCPPacket(data []byte) (*DHCPPacket, error) {
	if len(data) < 240 {
		return nil, fmt.Errorf("DHCP packet too short: %d bytes", len(data))
	}
	
	packet := &DHCPPacket{
		OpCode:       data[0],
		HardwareType: data[1],
		HardwareLen:  data[2],
		Hops:         data[3],
		TransactionID: binary.BigEndian.Uint32(data[4:8]),
		Seconds:      binary.BigEndian.Uint16(data[8:10]),
		Flags:        binary.BigEndian.Uint16(data[10:12]),
		ClientIP:     net.IP(data[12:16]),
		YourIP:       net.IP(data[16:20]),
		ServerIP:     net.IP(data[20:24]),
		GatewayIP:    net.IP(data[24:28]),
	}
	
	// Parse client MAC address
	macLen := int(packet.HardwareLen)
	if macLen > 16 {
		macLen = 16
	}
	packet.ClientMAC = net.HardwareAddr(data[28 : 28+macLen])
	
	// Parse server name
	copy(packet.ServerName[:], data[44:108])
	
	// Parse boot file
	copy(packet.BootFile[:], data[108:236])
	
	// Parse options
	if len(data) > 236 {
		options, err := p.parseDHCPOptions(data[236:])
		if err == nil {
			packet.Options = options
		}
	}
	
	return packet, nil
}

// parseDHCPOptions parses DHCP options from raw data
func (p *ProxyDHCP) parseDHCPOptions(data []byte) ([]DHCPOption, error) {
	var options []DHCPOption
	i := 0
	
	// Skip magic cookie (99, 130, 83, 99)
	if len(data) < 4 || data[0] != 99 || data[1] != 130 || data[2] != 83 || data[3] != 99 {
		return nil, fmt.Errorf("invalid DHCP magic cookie")
	}
	i = 4
	
	for i < len(data) {
		if data[i] == 255 { // End option
			break
		}
		
		if data[i] == 0 { // Pad option
			i++
			continue
		}
		
		if i+1 >= len(data) {
			break
		}
		
		code := data[i]
		length := int(data[i+1])
		
		if i+2+length > len(data) {
			break
		}
		
		option := DHCPOption{
			Code:   code,
			Length: byte(length),
			Data:   data[i+2 : i+2+length],
		}
		
		options = append(options, option)
		i += 2 + length
	}
	
	return options, nil
}

// handleDHCPPacket handles an incoming DHCP packet
func (p *ProxyDHCP) handleDHCPPacket(ctx context.Context, packet *DHCPPacket, addr *net.UDPAddr) {
	// Get message type from options
	messageType := p.getMessageType(packet.Options)
	
	switch messageType {
	case DHCPDiscover:
		p.handleDiscover(ctx, packet, addr)
	case DHCPRequest:
		p.handleRequest(ctx, packet, addr)
	case DHCPRelease:
		p.handleRelease(ctx, packet, addr)
	default:
		if p.config.Verbose {
			log.Printf("Ignoring DHCP message type %d from %s", messageType, packet.ClientMAC)
		}
	}
}

// handleDiscover handles a DHCP Discover packet
func (p *ProxyDHCP) handleDiscover(ctx context.Context, packet *DHCPPacket, addr *net.UDPAddr) {
	mac := packet.ClientMAC.String()
	
	if p.config.Verbose {
		log.Printf("DHCP Discover from %s", mac)
	}
	
	// Check if this MAC is in our inventory
	target, err := p.config.TargetLookup(mac)
	if err != nil {
		if p.config.Verbose {
			log.Printf("MAC %s not in inventory, ignoring", mac)
		}
		return
	}
	
	// Create proxy lease
	lease := &ProxyLease{
		MACAddress: mac,
		TargetID:   target.ID,
		OfferedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(5 * time.Minute),
		BootFile:   p.config.BootFile,
		NextServer: p.config.NextServer,
	}
	
	// Store lease
	p.mu.Lock()
	p.leaseTracker[mac] = lease
	p.mu.Unlock()
	
	// Build DHCP Offer
	offer := p.buildDHCPOffer(packet, target)
	
	// Send DHCP Offer
	err = p.sendDHCPPacket(offer, addr)
	if err != nil {
		log.Printf("Error sending DHCP Offer to %s: %v", mac, err)
		return
	}
	
	if p.config.Verbose {
		log.Printf("Sent DHCP Offer to %s (next-server=%s, bootfile=%s)", 
			mac, p.config.NextServer, p.config.BootFile)
	}
	
	// Publish event
	if p.eventBus != nil {
		p.eventBus.Publish(Event{
			Type:      "proxydhcp.offer",
			TargetID:  target.ID,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"mac":        mac,
				"nextServer": p.config.NextServer,
				"bootFile":   p.config.BootFile,
			},
		})
	}
}

// handleRequest handles a DHCP Request packet
func (p *ProxyDHCP) handleRequest(ctx context.Context, packet *DHCPPacket, addr *net.UDPAddr) {
	mac := packet.ClientMAC.String()
	
	if p.config.Verbose {
		log.Printf("DHCP Request from %s", mac)
	}
	
	// Check if we have a lease for this MAC
	p.mu.RLock()
	lease, exists := p.leaseTracker[mac]
	p.mu.RUnlock()
	
	if !exists {
		if p.config.Verbose {
			log.Printf("No lease found for %s, ignoring", mac)
		}
		return
	}
	
	// Build DHCP Ack
	ack := p.buildDHCPAck(packet)
	
	// Send DHCP Ack
	err := p.sendDHCPPacket(ack, addr)
	if err != nil {
		log.Printf("Error sending DHCP Ack to %s: %v", mac, err)
		return
	}
	
	if p.config.Verbose {
		log.Printf("Sent DHCP Ack to %s", mac)
	}
}

// handleRelease handles a DHCP Release packet
func (p *ProxyDHCP) handleRelease(ctx context.Context, packet *DHCPPacket, addr *net.UDPAddr) {
	mac := packet.ClientMAC.String()
	
	if p.config.Verbose {
		log.Printf("DHCP Release from %s", mac)
	}
	
	// Remove lease
	p.mu.Lock()
	delete(p.leaseTracker, mac)
	p.mu.Unlock()
}

// buildDHCPOffer builds a DHCP Offer packet
func (p *ProxyDHCP) buildDHCPOffer(discover *DHCPPacket, target *Target) *DHCPPacket {
	offer := &DHCPPacket{
		OpCode:       DHCPOffer,
		HardwareType: discover.HardwareType,
		HardwareLen:  discover.HardwareLen,
		Hops:         discover.Hops,
		TransactionID: discover.TransactionID,
		Seconds:      0,
		Flags:        discover.Flags,
		ClientIP:     net.IPv4zero,
		YourIP:       net.IPv4zero, // We don't assign IPs
		ServerIP:     net.ParseIP(p.config.NextServer),
		GatewayIP:    net.IPv4zero,
		ClientMAC:    discover.ClientMAC,
	}
	
	// Copy server name
	copy(offer.ServerName[:], "OmniGraph-ProxyDHCP")
	
	// Copy boot file
	copy(offer.BootFile[:], p.config.BootFile)
	
	// Build options
	offer.Options = []DHCPOption{
		{Code: OptionMessageType, Length: 1, Data: []byte{DHCPOffer}},
		{Code: OptionServerIdentifier, Length: 4, Data: net.ParseIP(p.config.NextServer).To4()},
		{Code: OptionTFTPServerName, Length: byte(len(p.config.NextServer)), Data: []byte(p.config.NextServer)},
		{Code: OptionBootfileName, Length: byte(len(p.config.BootFile)), Data: []byte(p.config.BootFile)},
	}
	
	// Add HTTP boot URL if configured
	if p.config.HTTPBootURL != "" {
		offer.Options = append(offer.Options, DHCPOption{
			Code:   OptionHTTPBootURL,
			Length: byte(len(p.config.HTTPBootURL)),
			Data:   []byte(p.config.HTTPBootURL),
		})
	}
	
	return offer
}

// buildDHCPAck builds a DHCP Ack packet
func (p *ProxyDHCP) buildDHCPAck(request *DHCPPacket) *DHCPPacket {
	ack := &DHCPPacket{
		OpCode:       DHCPAck,
		HardwareType: request.HardwareType,
		HardwareLen:  request.HardwareLen,
		Hops:         request.Hops,
		TransactionID: request.TransactionID,
		Seconds:      0,
		Flags:        request.Flags,
		ClientIP:     request.ClientIP,
		YourIP:       request.YourIP,
		ServerIP:     net.ParseIP(p.config.NextServer),
		GatewayIP:    request.GatewayIP,
		ClientMAC:    request.ClientMAC,
	}
	
	// Copy server name
	copy(ack.ServerName[:], "OmniGraph-ProxyDHCP")
	
	// Copy boot file
	copy(ack.BootFile[:], p.config.BootFile)
	
	// Build options
	ack.Options = []DHCPOption{
		{Code: OptionMessageType, Length: 1, Data: []byte{DHCPAck}},
		{Code: OptionServerIdentifier, Length: 4, Data: net.ParseIP(p.config.NextServer).To4()},
		{Code: OptionIPAddressLease, Length: 4, Data: []byte{0, 0, 0, 0}}, // No actual lease
	}
	
	return ack
}

// sendDHCPPacket sends a DHCP packet to the client
func (p *ProxyDHCP) sendDHCPPacket(packet *DHCPPacket, addr *net.UDPAddr) error {
	// Build raw packet
	data, err := p.buildRawPacket(packet)
	if err != nil {
		return fmt.Errorf("failed to build raw packet: %w", err)
	}
	
	// Send packet
	_, err = p.server.WriteToUDP(data, addr)
	if err != nil {
		return fmt.Errorf("failed to send packet: %w", err)
	}
	
	return nil
}

// buildRawPacket builds a raw DHCP packet from a DHCPPacket struct
func (p *ProxyDHCP) buildRawPacket(packet *DHCPPacket) ([]byte, error) {
	data := make([]byte, MaxDHCPacketSize)
	
	// Fill in header
	data[0] = packet.OpCode
	data[1] = packet.HardwareType
	data[2] = packet.HardwareLen
	data[3] = packet.Hops
	binary.BigEndian.PutUint32(data[4:8], packet.TransactionID)
	binary.BigEndian.PutUint16(data[8:10], packet.Seconds)
	binary.BigEndian.PutUint16(data[10:12], packet.Flags)
	copy(data[12:16], packet.ClientIP.To4())
	copy(data[16:20], packet.YourIP.To4())
	copy(data[20:24], packet.ServerIP.To4())
	copy(data[24:28], packet.GatewayIP.To4())
	copy(data[28:44], packet.ClientMAC)
	copy(data[44:108], packet.ServerName[:])
	copy(data[108:236], packet.BootFile[:])
	
	// Add magic cookie
	data[236] = 99
	data[237] = 130
	data[238] = 83
	data[239] = 99
	
	// Add options
	i := 240
	for _, opt := range packet.Options {
		if i+2+len(opt.Data) > MaxDHCPacketSize {
			break
		}
		data[i] = opt.Code
		data[i+1] = opt.Length
		copy(data[i+2:], opt.Data)
		i += 2 + len(opt.Data)
	}
	
	// Add end option
	if i < MaxDHCPacketSize {
		data[i] = 255
		i++
	}
	
	return data[:i], nil
}

// getMessageType extracts the message type from DHCP options
func (p *ProxyDHCP) getMessageType(options []DHCPOption) byte {
	for _, opt := range options {
		if opt.Code == OptionMessageType && len(opt.Data) > 0 {
			return opt.Data[0]
		}
	}
	return 0
}

// cleanupExpiredLeases periodically cleans up expired leases
func (p *ProxyDHCP) cleanupExpiredLeases(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-p.stopCh:
			return
		case <-ticker.C:
			p.mu.Lock()
			now := time.Now()
			for mac, lease := range p.leaseTracker {
				if now.After(lease.ExpiresAt) {
					delete(p.leaseTracker, mac)
					if p.config.Verbose {
						log.Printf("Cleaned up expired lease for %s", mac)
					}
				}
			}
			p.mu.Unlock()
		}
	}
}

// GetLeases returns all active proxy leases
func (p *ProxyDHCP) GetLeases() map[string]*ProxyLease {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	leases := make(map[string]*ProxyLease)
	for k, v := range p.leaseTracker {
		leases[k] = v
	}
	
	return leases
}

// SetEventBus sets the event bus for publishing events
func (p *ProxyDHCP) SetEventBus(bus EventBus) {
	p.eventBus = bus
}